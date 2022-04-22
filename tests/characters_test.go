package tests

import (
	"testing"

	"github.com/damishra/streamly/shared"
	"github.com/joho/godotenv"
)

func TestSearchCharacter(test *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		test.Errorf("Error occured: %s", err.Error())
	}

	characterName := "Ka chao"
	expected := "kristoferyee"

	character, err := shared.SearchCharacter(characterName)
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
