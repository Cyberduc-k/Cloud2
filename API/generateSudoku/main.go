package main

import (
	"api/model"
	"api/repository"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	/*"go.mongodb.org/mongo-driver/mongo/options"*/)

type Handler struct {
	sudokuRepo repository.Repository[model.Sudoku]
	channel    *amqp.Channel
}

type MyObjectID string

func (id MyObjectID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	p, err := primitive.ObjectIDFromHex(string(id))
	if err != nil {
		return bsontype.Null, nil, err
	}

	return bson.MarshalValue(p)
}

func main() {

	//context & background
	ctx := context.Background()

	//client and mongoDB setup
	client, err := setupMongo(ctx)
	if err != nil {
		log.Fatal(err)
	}

	//rabbitMQ Queue connection setup
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

	//var definition/creation of repo, handler & router (do I need these here?)
	sudokuRepo := repository.New[model.Sudoku](ctx, client, "SudokuDB", "Sudokus")
	handler := Handler{sudokuRepo, chl}

	//consume messages from the StartPuzzle RabbitMQ Queue (assuming that all queued messages here mean that a new puzzle has to be generated)
	msgs, err := chl.Consume(
		"StartPuzzle",    // queue
		"GeneratePuzzle", // consumer
		true,             // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)

	//forever := make(chan bool)
	//var forever chan struct{}
	forever := make(chan struct{})

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			/*fmt.Println(err)*/
			generateSudoku(d.Body, handler)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func generateSudoku(sudokuId []byte, handler Handler) {
	newSudoku := model.GenerateSudoku()

	// set sudoku id
	if err := newSudoku.Id.UnmarshalText(sudokuId); err != nil {
		log.Fatal(err)
	}

	log.Println(newSudoku.StartState)
	log.Println(newSudoku.Solution)

	//insert the new sudoku puzzle into the context sudoku repository (connected to the mongodb sudoku database)
	handler.sudokuRepo.Insert(newSudoku)

	// communicate to startSudoku that the sudoku has been generated
	handler.channel.Publish(
		"",
		"GeneratePuzzle",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte{0},
		},
	)
}

// mongo db connection function
func setupMongo(ctx context.Context) (*mongo.Client, error) {
	user := os.Getenv("MONGODB_USER")
	pass := os.Getenv("MONGODB_PASSWORD")
	conn := os.Getenv("MONGODB_CONNECTION")
	connString := fmt.Sprintf("mongodb://%s:%s@%s:27017", user, pass, conn)
	log.Println(connString)
	return repository.NewClient(ctx, connString)
}

// rabbitMQ connection & channel declaration
func setupRabbit() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
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

/*
func writeResponse[T any](w http.ResponseWriter, code int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Fatal(err)
	}
}
*/
