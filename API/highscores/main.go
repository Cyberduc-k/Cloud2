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
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler struct {
	userRepo repository.Repository[model.User]
}

func main() {
	ctx := context.Background()
	user := os.Getenv("MONGODB_USER")
	pass := os.Getenv("MONGODB_PASSWORD")
	conn := os.Getenv("MONGODB_CONNECTION")
	client, err := repository.NewClient(ctx, fmt.Sprintf("mongodb://%s:%s@%s:27017", user, pass, conn))

	if err != nil {
		log.Printf("error: %v", err)
	}

	repo := repository.New[model.User](ctx, client, "SudokuDB", "Users")
	handler := Handler{repo}
	router := mux.NewRouter()

	router.HandleFunc("/", handler.getHighscores).Methods("GET")
	router.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)

	if err := http.ListenAndServe(":8081", router); err != nil {
		log.Printf("error: %v", err)
		return
	}
}

func (self *Handler) getHighscores(writer http.ResponseWriter, request *http.Request) {
	users, err := self.userRepo.GetAll()

	if err != nil {
		log.Printf("error: %v", err)
		writer.WriteHeader(http.StatusNotFound)
		return
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
		log.Printf("eror: %v", err)
	}
}
