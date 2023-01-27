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
	userRepo   repository.Repository[model.User]
	sudokuRepo repository.Repository[model.Sudoku]
}

type MyObjectID string

func main() {
	ctx := context.Background()
	client, err := setupMongo(ctx)
	if err != nil {
		log.Fatal(err)
	}

	userRepo := repository.New[model.User](ctx, client, "SudokuDB", "Users")
	sudokuRepo := repository.New[model.Sudoku](ctx, client, "SudokuDB", "Sudokus")
	handler := Handler{userRepo, sudokuRepo}
	router := mux.NewRouter()

	router.HandleFunc("/", handler.stopSudoku).Methods("POST")

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

	log.Println(user)
	userSolution := request.FormValue("Solution")
	sudoku, err := self.sudokuRepo.GetById(user.CurrentSudokuId)
	if err != nil {
		log.Fatal(err)
	}

	//check if the puzzle is incorrect (if so, then return)
	if userSolution != sudoku.Solution {
		response := "Incorrect sudoku solution, please try again"
		writeResponse(writer, http.StatusOK, response)
		return
	}

	for i := range user.Sudokus {
		progress := &user.Sudokus[i]
		if !progress.IsSolved() && progress.SudokuId == user.CurrentSudokuId {
			progress.Solve()

			update := bson.M{"$set": bson.M{"sudokus": user.Sudokus}}
			if err := self.userRepo.Update(user.Id, update); err != nil {
				log.Fatal(err)
			}
			response := "Success!"
			writeResponse(writer, http.StatusOK, response)

			//set currentsudokuId for the user back to a nil primitive object ID
			user.CurrentSudokuId = primitive.NilObjectID
			setup := bson.M{"$set": bson.M{"currentsudoku": user.CurrentSudokuId}}
			if err := self.userRepo.Update(user.Id, setup); err != nil {
				log.Fatal(err)
			}

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
