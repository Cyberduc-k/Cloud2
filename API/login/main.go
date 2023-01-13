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
		log.Fatal(err)
	}

	repo := repository.New[model.User](ctx, client, "SudokuDB", "Users")
	handler := Handler{repo}
	router := mux.NewRouter()

	router.HandleFunc("/", handler.Login).Methods("Post")

	if err := http.ListenAndServe(":8081", router); err != nil {
		log.Fatal(err)
	}
}

func (self *Handler) Login(writer http.ResponseWriter, request *http.Request) {
	var loginInfo model.User

	loginInfo.Username = request.FormValue("Username")
	loginInfo.Password = request.FormValue("Password")
	user, err := self.userRepo.Login(loginInfo)
	if err != nil {
		log.Fatal(err)
	}

	var loginResponse model.LoginResponse

	loginResponse.Username = user.Username
	loginResponse.Sudokus = user.Sudokus
	loginResponse.Id = user.Id

	writeResponse(writer, http.StatusOK, loginResponse)
}

func writeResponse[T any](w http.ResponseWriter, code int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Fatal(err)
	}
}
