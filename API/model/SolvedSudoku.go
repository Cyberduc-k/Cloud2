package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SolvedSudoku struct {
	SudokuId    primitive.ObjectID
	DateStarted time.Time
	DateSolved  time.Time
	TimeToSolve time.Duration
}

func NewSolvedSudoku(sudokuId primitive.ObjectID) SolvedSudoku {
	dateStarted := time.Now()
	return SolvedSudoku{SudokuId: sudokuId, DateStarted: dateStarted}
}

func (self *SolvedSudoku) IsSolved() bool {
	return self.DateSolved != time.Time{}
}

func (self *SolvedSudoku) Solve() {
	self.DateSolved = time.Now()
	self.TimeToSolve = self.DateStarted.Sub(self.DateSolved)
}
