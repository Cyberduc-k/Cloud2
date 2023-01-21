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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	userRepo   repository.Repository[model.User]
	sudokuRepo repository.Repository[model.Sudoku]
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

	userRepo := repository.New[model.User](ctx, client, "SudokuDB", "Users")
	sudokuRepo := repository.New[model.Sudoku](ctx, client, "SudokuDB", "Sudokus")
	handler := Handler{userRepo, sudokuRepo}
	router := mux.NewRouter()

	router.HandleFunc("/", handler.Login).Methods("Post")

	router.HandleFunc("/puzzles", handler.getStartSudokuResponse).Methods("Post")

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

func (self *Handler) getStartSudokuResponse(writer http.ResponseWriter, request *http.Request) {
	sudokuIdString := request.FormValue("SudokuId")

	sudokuId, err := primitive.ObjectIDFromHex(sudokuIdString)
	if err != nil {
		log.Fatal(err)
	}

	sudoku, err := self.sudokuRepo.GetById(sudokuId)
	if err != nil {
		log.Fatal(err)
	}

	response := model.StartSudokuResponse{Id: sudokuId, StartState: sudoku.StartState}
	writeResponse(writer, http.StatusOK, response)
}

func writeResponse[T any](w http.ResponseWriter, code int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Fatal(err)
	}
}
