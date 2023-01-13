package main

import (
	"api/model"
	"api/repository"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	userRepo   repository.Repository[model.User]
	sudokuRepo repository.Repository[model.Sudoku]
	channel    *amqp.Channel
	msgs       <-chan amqp.Delivery
}

func main() {
	ctx := context.Background()
	client, err := setupMongo(ctx)
	if err != nil {
		log.Fatal(err)
	}

	rabbitConn, chl, err := setupRabbit()
	if err != nil {
		log.Fatal(err)
	}

	defer rabbitConn.Close()
	defer chl.Close()

	if err = setupQueue(chl, "StartPuzzle"); err != nil {
		log.Fatal(err)
	}

	if err = setupQueue(chl, "GeneratePuzzle"); err != nil {
		log.Fatal(err)
	}

	//consume messages from the GeneratePuzzle RabbitMQ Queue (assuming that all queued messages here mean that a new puzzle has been generated)
	msgs, err := chl.Consume(
		"GeneratePuzzle", // queue
		"StartPuzzle",    // consumer
		true,             // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)

	if err != nil {
		log.Fatal(err)
	}

	userRepo := repository.New[model.User](ctx, client, "SudokuDB", "Users")
	sudokuRepo := repository.New[model.Sudoku](ctx, client, "SudokuDB", "Sudokus")
	handler := Handler{userRepo, sudokuRepo, chl, msgs}
	router := mux.NewRouter()

	router.HandleFunc("/", handler.startSudoku).Methods("POST")

	if err := http.ListenAndServe(":8081", router); err != nil {
		log.Fatal(err)
	}
}

func (self *Handler) startSudoku(writer http.ResponseWriter, request *http.Request) {
	userIdString := request.Header.Get("Authorization")
	userId, err := primitive.ObjectIDFromHex(userIdString)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(userId)
	user, err := self.userRepo.GetById(userId)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(user)
	completed := make([]primitive.ObjectID, 0, len(user.Sudokus))

	// check if the user has any unfinished sudokus
	for _, progress := range user.Sudokus {
		if !progress.IsSolved() {
			filter := bson.M{"_id": user.Id, "sudokus.sudokuid": progress.SudokuId}
			update := bson.M{"$set": bson.M{"sudokus.$.datestarted": time.Now()}}
			if err := self.userRepo.UpdateWhere(filter, update); err != nil {
				log.Fatal(err)
			}

			user.CurrentSudokuId = progress.SudokuId
			setup := bson.M{"$set": bson.M{"currentsudokuid": user.CurrentSudokuId}}
			if err := self.userRepo.Update(user.Id, setup); err != nil {
				log.Fatal(err)
			}

			self.returnSudoku(writer, progress.SudokuId)
			return
		}

		completed = append(completed, progress.SudokuId)
	}

	// check if there are any sudokus the user hasn't done
	filter := bson.M{"_id": bson.M{"$not": bson.M{"$in": completed}}}
	uncompletedSudokus, err := self.sudokuRepo.GetAllWhere(filter)
	if err != nil {
		log.Fatal(err)
	}

	for _, sudoku := range uncompletedSudokus {
		self.onSudokuCreated(sudoku.Id, user)
		self.returnSudoku(writer, sudoku.Id)
		return
	}

	// genreate a new sudoku
	sudokuId := primitive.NewObjectID()
	bytes, _ := sudokuId.MarshalText()

	self.channel.Publish(
		"",
		"StartPuzzle",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        bytes,
		},
	)

	forever := make(chan struct{})

	go func() {
		for d := range self.msgs {
			log.Printf("Received a message: %s", d.Body)
			self.onSudokuCreated(sudokuId, user)
			self.returnSudoku(writer, sudokuId)
			// cancel the forever
			forever <- struct{}{}
			return
		}
	}()

	<-forever
}

func (self *Handler) onSudokuCreated(sudokuId primitive.ObjectID, user model.User) {
	progress := model.NewSolvedSudoku(sudokuId)
	update := bson.M{"$push": bson.M{"sudokus": progress}}

	if err := self.userRepo.Update(user.Id, update); err != nil {
		log.Fatal(err)
	}

	user.CurrentSudokuId = sudokuId
	setup := bson.M{"$set": bson.M{"currentsudokuid": user.CurrentSudokuId}}
	if err := self.userRepo.Update(user.Id, setup); err != nil {
		log.Fatal(err)
	}

}

func (self *Handler) returnSudoku(writer http.ResponseWriter, sudokuId primitive.ObjectID) {
	sudoku, err := self.sudokuRepo.GetById(sudokuId)
	if err != nil {
		log.Fatal(err)
	}

	response := model.StartSudokuResponse{Id: sudokuId, StartState: sudoku.StartState}
	writeResponse(writer, http.StatusOK, response)
}

func setupMongo(ctx context.Context) (*mongo.Client, error) {
	user := os.Getenv("MONGODB_USER")
	pass := os.Getenv("MONGODB_PASSWORD")
	conn := os.Getenv("MONGODB_CONNECTION")
	connString := fmt.Sprintf("mongodb://%s:%s@%s:27017", user, pass, conn)
	log.Println(connString)
	return repository.NewClient(ctx, connString)
}

func setupRabbit() (*amqp.Connection, *amqp.Channel, error) {
	user := os.Getenv("RABBITMQ_USER")
	pass := os.Getenv("RABBITMQ_PASSWORD")
	hostname := os.Getenv("RABBITMQ_CONNECTION")
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:5672", user, pass, hostname))
	if err != nil {
		return nil, nil, err
	}

	chl, err := conn.Channel()
	return conn, chl, err
}

func setupQueue(chl *amqp.Channel, queueName string) error {
	_, err := chl.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)

	return err
}

func writeResponse[T any](w http.ResponseWriter, code int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Fatal(err)
	}
}
