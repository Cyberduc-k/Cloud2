package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Sudoku struct {
	Id         primitive.ObjectID `bson:"_id,omitempty"`
	StartState string
	Solution   string
}

func Generate() Sudoku {
	// TODO: generate sudoku
	return Sudoku{}
}
