package main

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

const schema string = `DROP TABLE IF EXISTS public.characters;
DROP TABLE IF EXISTS public.streamers;
CREATE TABLE public.streamers (
	id BIGSERIAL PRIMARY KEY,
	username VARCHAR(32) UNIQUE NOT NULL,
	pid VARCHAR(32) UNIQUE NOT NULL
);
CREATE TABLE public.characters (
	id BIGSERIAL PRIMARY KEY,
	fullname VARCHAR(255) NOT NULL,
	fullname_token TSVECTOR NOT NULL,
	player BIGINT NOT NULL REFERENCES public.streamers(id),
	UNIQUE (fullname, player)
);`

var ctx = context.Background()

func main() {
	pool, err := pgxpool.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("error: %s", err.Error())
		return
	}
	defer pool.Close()

	streamers, characters := FetchData()

	if err := ClearDB(pool); err != nil {
		log.Fatalf("error: %s", err.Error())
		return
	}

	if err := LoadStreamers(pool, streamers); err != nil {
		log.Fatalf("error: %s", err.Error())
		return
	}

	LoadCharacters(pool, characters)
}

func FetchData() ([][]interface{}, [][]string) {
	var (
		streamers  [][]interface{} = make([][]interface{}, 0)
		characters [][]string      = make([][]string, 0)
	)

	col := colly.NewCollector()
	col.OnHTML("#characters", func(h *colly.HTMLElement) {
		h.ForEach("div[data-streamer]", func(_ int, e *colly.HTMLElement) {
			streamerID := e.Attr("data-streamer")
			characterName := e.ChildText(".charName")
			characterName = strings.ReplaceAll(characterName, "“”", "\"")
			streamerName := e.ChildText(".profileLink")
			streamerName = strings.ToLower(streamerName)

			isUnique := true
			for _, streamer := range streamers {
				if streamer[0] == streamerID {
					isUnique = false
					break
				}
			}
			if isUnique {
				streamers = append(streamers, []interface{}{streamerID, streamerName})
			}

			characters = append(characters, []string{characterName, streamerID})
		})
	})
	col.Visit("https://nopixel.hasroot.com/characters.php")

	defer log.Printf("Characters found: %d", len(characters))
	defer log.Printf("Streamers found: %d", len(streamers))

	return streamers, characters
}

func ClearDB(pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, schema)
	return err
}

func LoadStreamers(pool *pgxpool.Pool, streamers [][]interface{}) error {
	count, err := pool.CopyFrom(
		ctx,
		pgx.Identifier{"streamers"},
		[]string{"pid", "username"},
		pgx.CopyFromRows(streamers),
	)
	if err != nil {
		return err
	}
	defer log.Printf("Streamers inserted: %d", count)
	return nil
}

func LoadCharacters(pool *pgxpool.Pool, characters [][]string) {
	var wg sync.WaitGroup
	count := 0
	defer log.Printf("Characters inserted: %d", count)
	defer wg.Wait()

	for _, character := range characters {
		wg.Add(1)
		go func(character []string, pool *pgxpool.Pool) {
			defer wg.Done()
			conn, err := pool.Acquire(ctx)
			if err != nil {
				log.Printf("error: %s", err.Error())
			}
			defer conn.Release()

			_, err = conn.Exec(
				ctx,
				`INSERT INTO public.characters (fullname, fullname_token, player)
				SELECT $1::VARCHAR, to_tsvector($1::VARCHAR), s.id
				FROM public.streamers s
				WHERE s.pid = $2;`,
				character[0],
				character[1],
			)
			if err != nil {
				log.Printf("error: %s", err.Error())
			}
			count++
		}(character, pool)
	}
}
