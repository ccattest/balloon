package main

import (
	"context"
	"flag"
	"log"
	"time"
	"tubr/config"
	"tubr/store"
	"tubr/twitch"
)

var (
	GAME_ID = flag.String("game", "25416", "game ID")
	T1      = flag.String("t1", "", "t1")
	T2      = flag.String("t2", "", "t2")
	LIMIT   = flag.Int("limit", 1000, "limit")
)

func main() {
	flag.Parse()
	cfg := config.FromENV()
	ctx := context.Background()
	db, err := store.Open(ctx, cfg.PostgresConfig)
	if err != nil {
		log.Fatal("failed to open database")
	}

	var t1, t2 time.Time
	if len(*T1) > 0 {
		t1, err = time.Parse(time.DateOnly, *T1)
		if err != nil {
			log.Fatalf("Failed to parse t1: %s %s", *T1, err)
		}

		t2, err = time.Parse(time.DateOnly, *T2)
		if err != nil {
			log.Fatalf("Failed to parse t1: %s %s", *T2, err)
		}
	}

	tc := twitch.NewClient(nil)
	clips, err := tc.ClipsByGame(ctx, &twitch.CBGP{
		GameID: *GAME_ID,
		RequestParams: &twitch.RequestParams{
			Cursor: &twitch.Cursor{},
			TimeParams: &twitch.TimeParams{
				T1: &t1,
				T2: &t2,
			},
		},
	}, *LIMIT)

	if err != nil {
		log.Fatal("failed to pull clips for game id")
	}

	log.Printf("Inserting %d clips\n", len(clips))
	for _, clip := range clips {
		if err = db.InsertClip(ctx, &clip); err != nil {
			log.Println("failed to insert clip")
		}
	}
	log.Println("Clips inserted")
}
