package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/damishra/NopixelDB/shared"
	"github.com/jackc/pgx/v4"
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
		return
	}

	fullname := request.Form.Get("name")
	if fullname == "" {
		writer.Write([]byte("Bad Request: field `name` is empty"))
		return
	}

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		shared.HandleServerError(&writer, err)
		return
	}
	defer conn.Close(ctx)

	character, err := shared.SearchCharacter(ctx, fullname, conn)
	if err != nil {
		writer.Write([]byte("Character Not Found"))
		return
	}

	responseStr := fmt.Sprintf("%s is played by twitch.tv/%s", character.Fullname, character.Username)
	writer.Write([]byte(responseStr))
}
