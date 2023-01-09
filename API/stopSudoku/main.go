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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	userRepo repository.Repository[model.User]
}

type MyObjectID string

func main() {
	ctx := context.Background()
	client, err := setupMongo(ctx)
	if err != nil {
		log.Fatal(err)
	}

	userRepo := repository.New[model.User](ctx, client, "SudokuDB", "Users")
	handler := Handler{userRepo}
	router := mux.NewRouter()

	router.HandleFunc("/sudokus/finish", handler.stopSudoku).Methods("POST")

	if err := http.ListenAndServe(":8081", router); err != nil {
		log.Fatal(err)
	}
}

func (self *Handler) stopSudoku(writer http.ResponseWriter, request *http.Request) {
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

	for i := range user.Sudokus {
		progress := &user.Sudokus[i]
		if !progress.IsSolved() {
			progress.Solve()

			update := bson.M{"$set": bson.M{"sudokus": user.Sudokus}}
			if err := self.userRepo.Update(user.Id, update); err != nil {
				log.Fatal(err)
			}
			response := ""
			writeResponse(writer, http.StatusOK, response)
			return
		}
	}
}

func setupMongo(ctx context.Context) (*mongo.Client, error) {
	user := os.Getenv("MONGODB_USER")
	pass := os.Getenv("MONGODB_PASSWORD")
	conn := os.Getenv("MONGODB_CONNECTION")
	connString := fmt.Sprintf("mongodb://%s:%s@%s:27017", user, pass, conn)
	log.Println(connString)
	return repository.NewClient(ctx, connString)
}

func writeResponse[T any](w http.ResponseWriter, code int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Fatal(err)
	}
}
