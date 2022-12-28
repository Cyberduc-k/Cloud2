package main

import (
	"api/model"
	"api/repository"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	userRepo repository.Repository[model.User]
}

func main() {
	ctx := context.Background()
	client, err := repository.NewClient(ctx, "mongodb://localhost:27017")

	if err != nil {
		log.Fatal(err)
	}

	repo := repository.New[model.User](ctx, client, "SudokuDB", "Users")
	handler := Handler{repo}
	router := mux.NewRouter()

	router.HandleFunc("/users", handler.getUsers).Methods("GET")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}

func (self *Handler) getUsers(writer http.ResponseWriter, request *http.Request) {
	users, err := self.userRepo.GetAll()

	if err != nil {
		log.Fatal(err)
	}

	writeResponse(writer, http.StatusOK, users)
}

func writeResponse[T any](w http.ResponseWriter, code int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Fatal(err)
	}
}
