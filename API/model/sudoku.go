package model

import (
	"github.com/direvus/sudoku"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Sudoku struct {
	Id         primitive.ObjectID `bson:"_id,omitempty"`
	StartState string
	Solution   string
}

func GenerateSudoku() Sudoku {
	solution := sudoku.GenerateSolution()
	mask := solution.MinimalMask()
	puzzle := solution.ApplyMask(&mask)

	return Sudoku{Id: primitive.NilObjectID, StartState: puzzle.String(), Solution: solution.String()}
}
