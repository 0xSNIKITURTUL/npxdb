package main

import (
	"context"
	"flag"
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
	username VARCHAR(32) NOT NULL,
	pid VARCHAR(32) NOT NULL,
	UNIQUE(username, pid)
);
CREATE TABLE public.characters (
	id BIGSERIAL PRIMARY KEY,
	fullname VARCHAR(255) NOT NULL,
	fullname_token TSVECTOR NOT NULL,
	player BIGINT NOT NULL REFERENCES public.streamers(id)
);`

var ctx = context.Background()

var fullreset = flag.Bool("reset", false, "reset the database and insert fresh values")

func main() {
	flag.Parse()

	pool, err := pgxpool.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("error: %s", err.Error())
		return
	}
	defer pool.Close()

	if *fullreset {
		if err := ClearDB(pool); err != nil {
			log.Fatalf("error: %s", err.Error())
			return
		}
	}

	streamers, characters := FetchData(pool)
	if err != nil {
		log.Fatalf("error: %s", err.Error())
		return
	}

	if err := LoadStreamers(pool, streamers); err != nil {
		log.Fatalf("error: %s", err.Error())
		return
	}

	LoadCharacters(pool, characters)
}

func FetchData(pool *pgxpool.Pool) ([][]interface{}, [][]string) {
	var (
		streamers  [][]interface{} = make([][]interface{}, 0)
		characters [][]string      = make([][]string, 0)
		tempChars  [][]string      = make([][]string, 0)
		tempStrem  [][]string      = make([][]string, 0)
		wg         sync.WaitGroup
	)

	col := colly.NewCollector()
	col.OnHTML("#characters", func(h *colly.HTMLElement) {
		wg.Add(1)
		go func() {
			resChars, _ := pool.Query(ctx, "SELECT s.pid, c.fullname FROM characters c JOIN streamers s ON c.player = s.id")
			for resChars.Next() {
				var pid, fullname string
				resChars.Scan(&pid, &fullname)
				tempChars = append(tempChars, []string{fullname, pid})
			}
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			resStrem, _ := pool.Query(ctx, "SELECT username, pid FROM streamers")
			for resStrem.Next() {
				var pid, username string
				resStrem.Scan(&username, &pid)
				tempStrem = append(tempStrem, []string{pid, username})
			}
			wg.Done()
		}()

		wg.Wait()

		h.ForEach("div[data-streamer]", func(_ int, e *colly.HTMLElement) {
			streamerID := e.Attr("data-streamer")
			characterName := e.ChildText(".charName")
			characterName = strings.ReplaceAll(characterName, "“”", "\"")
			streamerName := e.ChildText(".profileLink")
			streamerName = strings.ToLower(streamerName)

			isUniqueStreamer := true
			for _, streamer := range tempStrem {
				if streamer[0] == streamerID {
					isUniqueStreamer = false
				}
			}
			for _, streamer := range streamers {
				if streamer[0] == streamerID {
					isUniqueStreamer = false
					break
				}
			}
			if isUniqueStreamer {
				streamers = append(streamers, []interface{}{streamerID, streamerName})
			}

			isUniqueCharacter := true
			for _, character := range tempChars {
				if character[0] == characterName && character[1] == streamerID {
					isUniqueCharacter = false
				}
			}
			if isUniqueCharacter {
				characters = append(characters, []string{characterName, streamerID})
			}
		})
	})
	col.Visit("https://nopixel.hasroot.com/characters.php")

	defer log.Printf("Characters found: %d", len(characters))
	defer log.Printf("Streamers found: %d", len(streamers))

	return streamers, characters
}

// ClearDB drops the existing schema in the database and recreates it.
// To be used only in case of a fresh database or a data corruption scenario.
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
		}(character, pool)
	}
}
