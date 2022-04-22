package handler

import (
	"net/http"
	"strings"
)

func CharacterHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Header().Add("Content-Type", "text/plain")

	if request.Method != "GET" {
		writer.Write([]byte("Method Not Allowed"))
		return
	}

	fullname := request.Form.Get("name")
	if strings.Compare(fullname, "") != 0 {
		writer.Write([]byte("Bad Request: field `name` is empty"))
		return
	}

	writer.Write([]byte(fullname))
	/*
		character, err := shared.SearchCharacter(fullname)
		if err != nil {
			writer.Write([]byte("Character Not Found"))
		}
		responseStr := fmt.Sprintf("%s is played by twitch.tv/%s", character.Fullname, character.Username)
		writer.Write([]byte(responseStr))
	*/
}
