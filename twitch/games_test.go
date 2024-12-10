package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestTopGames(t *testing.T) {
	ctx := context.Background()
	gr, err := testClient.TopGames(ctx)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%+v\n", gr)
}

func TestGames(t *testing.T) {
	ctx := context.Background()
	gr, err := testClient.Games(ctx, []string{
		"18122", "27471", "488644"},
	)

	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", gr)
}

func TestDownloadGameList(t *testing.T) {
	cachePath := "game-list.json"
	cachedGames := map[string]string{}
	if err := func() error {
		f, err := os.OpenFile(cachePath,
			os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		err = json.NewDecoder(f).Decode(&cachedGames)
		if err != nil {
			return err
		}
		return f.Truncate(0)
	}(); err != nil {
		fmt.Println("failed to read cache")
	}

	cache, err := os.OpenFile(cachePath,
		os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()

	ctx := context.Background()
	batch := []string{}
	for i := 0; i < 600000; i++ {
		id := strconv.Itoa(i)
		if _, ok := cachedGames[id]; ok {
			continue
		}

		batch = append(batch, id)
		if len(batch) == 100 {
			r, err := testClient.Games(ctx, batch)
			if err != nil {
				t.Error(err)
			}

			for _, g := range r.Data {
				fmt.Printf("Got %s %s\n",
					g.ID, g.Title)
				cachedGames[g.ID] = g.Title
			}

			batch = []string{}
		}
	}

	out, err := json.MarshalIndent(cachedGames, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	_, err = cache.Write(out)
	if err != nil {
		t.Fatal(err)
	}
}
