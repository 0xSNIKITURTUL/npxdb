package tests

import (
	"context"
	"os"
	"testing"

	"github.com/damishra/NopixelDB/shared"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
)

func TestSearchCharacter(test *testing.T) {
	ctx := context.Background()

	if err := godotenv.Load("../.env"); err != nil {
		test.Errorf("Error occured: %s", err.Error())
	}

	characterName := "Ka chao"
	expected := "kristoferyee"

	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		test.Errorf("Error occured: %s", err.Error())
	}

	character, err := shared.SearchCharacter(ctx, characterName, conn)
	if err != nil {
		test.Errorf("Error occured: %s", err.Error())
	}

	if character.Username != expected {
		test.Errorf(
			"Expected streamer (%s) is not the same as actual streamer (%s)",
			expected,
			character.Username,
		)
	}
}
