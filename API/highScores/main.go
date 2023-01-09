package main

import (
	"api/model"
	"api/repository"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Handler struct {
	userRepo repository.Repository[model.User]
}

func main() {
	ctx := context.Background()
	user := "admin"
	pass := "admin"
	conn := "localhost"
	client, err := repository.NewClient(ctx, fmt.Sprintf("mongodb://%s:%s@%s:27017", user, pass, conn))

	if err != nil {
		log.Fatal(err)
	}

	repo := repository.New[model.User](ctx, client, "SudokuDB", "Users")
	handler := Handler{repo}
	router := mux.NewRouter()

	router.HandleFunc("/users", handler.getHighscores).Methods("GET")

	if err := http.ListenAndServe(":8081", router); err != nil {
		log.Fatal(err)
	}
}

func (self *Handler) getHighscores(writer http.ResponseWriter, request *http.Request) {
	users, err := self.userRepo.GetAll()

	if err != nil {
		log.Fatal(err)
	}

	var score time.Duration
	var scoreArray []time.Duration
	for _, user := range users {
		for _, sudoku := range user.Sudokus {

			score = sudoku.DateStarted.Sub(sudoku.DateSolved)
		}
		scoreArray = append(scoreArray, score)
	}
	writeResponse(writer, http.StatusOK, scoreArray)
}

func writeResponse[T any](w http.ResponseWriter, code int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Fatal(err)
	}
}
