package shared

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
)

type Character struct {
	Fullname string
	Username string
}

func SearchCharacter(ctx context.Context, name string, conn *pgx.Conn) (*Character, error) {
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

	if &fullname == nil || &username == nil {
		return nil, fmt.Errorf("nothing found in database")
	}

	return &Character{
		Username: username,
		Fullname: fullname,
	}, nil
}
