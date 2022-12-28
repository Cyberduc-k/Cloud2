package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id            primitive.ObjectID `bson:"_id,omitempty"`
	Username      string
	Password      string
	SolvedSudokus []SolvedSudoku
}

func NewUser(username, password string) User {
	return User{Username: username, Password: password}
}
