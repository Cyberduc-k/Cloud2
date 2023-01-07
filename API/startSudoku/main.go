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

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	userRepo   repository.Repository[model.User]
	sudokuRepo repository.Repository[model.Sudoku]
	channel    *amqp.Channel
}

func main() {
	ctx := context.Background()
	client, err := setupMongo(ctx)
	if err != nil {
		log.Fatal(err)
	}

	rabbitConn, chl, err := setupRabbit("StartPuzzle")
	if err != nil {
		log.Fatal(err)
	}

	defer rabbitConn.Close()
	userRepo := repository.New[model.User](ctx, client, "SudokuDB", "Users")
	sudokuRepo := repository.New[model.Sudoku](ctx, client, "SudokuDB", "Sudokus")
	handler := Handler{userRepo, sudokuRepo, chl}
	router := mux.NewRouter()

	router.HandleFunc("/sudokus/start", handler.startSudoku).Methods("POST")

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

	writeResponse(writer, http.StatusOK, model.StartSudokuResponse{Id: sudokuId})
}

func setupMongo(ctx context.Context) (*mongo.Client, error) {
	user := os.Getenv("MONGODB_USER")
	pass := os.Getenv("MONGODB_PASSWORD")
	conn := os.Getenv("MONGODB_CONNECTION")
	connString := fmt.Sprintf("mongodb://%s:%s@%s:27017", user, pass, conn)
	log.Println(connString)
	return repository.NewClient(ctx, connString)
}

func setupRabbit(queueName string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		return nil, nil, err
	}

	chl, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	_, err = chl.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)

	return conn, chl, err
}

func writeResponse[T any](w http.ResponseWriter, code int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Fatal(err)
	}
}