package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/damishra/streamly/shared"
)

func CharacterHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Header().Add("Content-Type", "text/plain")

	if request.Method != "GET" {
		if _, err := writer.Write([]byte("Method Not Allowed")); err != nil {
			shared.HandleServerError(&writer, err)
		}
		return
	}

	fullname := request.Form.Get("name")
	if strings.Compare(fullname, "") != 0 {
		if _, err := writer.Write([]byte("Bad Request: field `name` is empty")); err != nil {
			shared.HandleServerError(&writer, err)
		}
		return
	}

	character, err := shared.SearchCharacter(fullname)
	if err != nil {
		if _, err := writer.Write([]byte("Character Not Found")); err != nil {
			shared.HandleServerError(&writer, err)
		}
	}

	responseStr := fmt.Sprintf("%s is played by twitch.tv/%s", character.Fullname, character.Username)
	if _, err := writer.Write([]byte(responseStr)); err != nil {
		shared.HandleServerError(&writer, err)
	}
}
