package handler

import (
	"net/http"

	"github.com/damishra/streamly/shared"
)

func CharacterHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Header().Add("Content-Type", "text/plain")

	if request.Method != "GET" {
		writer.Write([]byte("Method Not Allowed"))
		return
	}

	if err := request.ParseForm(); err != nil {
		shared.HandleServerError(&writer, err)
	}

	fullname := request.Form.Get("name")
	if fullname == "" {
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
