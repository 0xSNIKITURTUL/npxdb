package shared

import (
	"context"
	"os"
	"strings"

	"github.com/jackc/pgx/v4"
)

type Character struct {
	Fullname string
	Username string
}

func SearchCharacter(name string) (*Character, error) {
	var ctx = context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	defer conn.Close(ctx)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(name, " ")
	queryString := strings.Join(parts, " & ")

	var username string
	var fullname string
	result := conn.QueryRow(ctx, `
		SELECT c.fullname, s.username FROM public.characters c
		JOIN public.streamers s ON c.player = s.id
		WHERE c.fullname_token @@ to_tsquery($1) LIMIT 1;
	`, queryString)
	if err := result.Scan(&fullname, &username); err != nil {
		return nil, err
	}

	return &Character{
		Username: username,
		Fullname: fullname,
	}, nil
}
