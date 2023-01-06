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

func (Sudoku) Generate() Sudoku {
	solution := sudoku.GenerateSolution()
	mask := solution.MinimalMask()
	puzzle := solution.ApplyMask(&mask)

	return Sudoku{primitive.NewObjectID(), puzzle.String(), solution.String()}
}
