package store

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"

	"tubr/models"
	"tubr/twitch"
)

var (
	testDatabase *DB
	testClient   *twitch.Client
	testClips    = []models.Clip{}
	testGames    = []models.Game{}
)

func TestMain(m *testing.M) {
	var err error
	testClient = twitch.NewClient(&http.Client{})
	if err != nil {
		log.Fatal(err)
	}

	testGames, err = twitch.LoadGames()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	testDatabase, err = Open(ctx, PostgresConfig{
		Host:     "localhost",
		User:     "postgres",
		Password: "postgres",
		DbName:   "postgres",
	})
	if err != nil {
		log.Fatal(err)
	}

	defer testDatabase.Close()
	os.Exit(m.Run())
}

// func TestInsertClips(t *testing.T) {
// 	ctx := context.Background()
// 	for i, g := range testGames {
// 		if i > 5 {
// 			break
// 		}

// 		params := &twitch.CBGP{
// 			GameID: g.ID,
// 			RequestParams: &twitch.RequestParams{
// 				TimeParams: &twitch.TimeParams{T1: nil, T2: nil},
// 			},
// 		}
// 		data, err := testClient.ClipsByGame(ctx, params, 100)
// 		if err != nil {
// 			t.Error(err)
// 		}

// 		testClips = append(testClips, data...)
// 	}

// 	for _, c := range testClips {
// 		if err := testDatabase.InsertClip(ctx, &c); err != nil {
// 			t.Error(err)
// 		}
// 	}
// }
