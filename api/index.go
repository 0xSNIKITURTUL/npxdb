package handler

import (
	"context"
	"net/http"
)

type Character struct {
	Fullname string
	Username string
}

var ctx = context.Background()

func Handler(writer http.ResponseWriter, request *http.Request) {

}

func searchCharacter(name string) *Character {

	return nil
}
