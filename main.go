package main

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var ctx = context.Background()

func main() {}

func FetchData() ([][]interface{}, [][]string) {
	var (
		streamers  [][]interface{} = make([][]interface{}, 0)
		characters [][]string      = make([][]string, 0)
	)

	col := colly.NewCollector()
	defer col.Visit("https://nopixel.hasroot.com/characters.php")
	col.OnHTML("#characters", func(h *colly.HTMLElement) {
		h.ForEach("div[data-streamer]", func(_ int, e *colly.HTMLElement) {
			streamerID := strings.Trim(e.Attr("data-streamer"), " ")
			characterName := strings.Trim(e.ChildText(".charName"), " ")
			characterName = strings.ReplaceAll(characterName, "“”", "\"")
			streamerName := strings.Trim(e.ChildText(".profileLink"), " ")
			streamerName = strings.ToLower(streamerName)

			isPresent := false
			for _, streamer := range streamers {
				if streamer[0] == streamerID {
					isPresent = true
					break
				}
			}
			if isPresent {
				streamers = append(streamers, []interface{}{streamerID, streamerName})
			}

			characters = append(characters, []string{characterName, streamerID})
		})
	})

	return streamers, characters
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
	log.Printf("Streamers inserted: %d", count)
	return nil
}

func LoadCharacters(pool *pgxpool.Pool, characters [][]string) error {
	var wg sync.WaitGroup
	count := 0
	defer wg.Wait()
	errorChan := make(chan error)

	for _, character := range characters {
		wg.Add(1)
		go func(character []string, pool *pgxpool.Pool) {
			defer wg.Done()
			conn, err := pool.Acquire(ctx)
			if err != nil {
				errorChan <- err
				return
			}
			defer conn.Release()

			_, err = conn.Exec(
				ctx,
				`INSERT INTO TABLE public.characters
				SELECT $1, to_tsvector($1), s.id
				FROM public.streamers s
				WHERE s.pid = $2;`,
				character[0],
				character[1],
			)
			if err != nil {
				errorChan <- err
				return
			}
			count++
		}(character, pool)
	}

	if err := <-errorChan; err != nil {
		return err
	}

	return nil
}
