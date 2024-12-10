package twitch

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestClipByID(t *testing.T) {
	ctx := context.Background()
	clip, err := testClient.ClipByID(ctx,
		"AwkwardHelplessSalamanderSwiftRage")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(clip)
}

func TestClipsByGame(t *testing.T) {
	ctx := context.Background()
	t1 := time.Now().Add(-time.Hour * 24 * 365)
	params := &CBGP{
		GameID: "32399",
		RequestParams: &RequestParams{
			TimeParams: &TimeParams{&t1, nil},
		},
	}
	clips, err := testClient.ClipsByGame(ctx, params, 100)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(clips)
}

func TestDownloadTopClips(t *testing.T) {
	ctx := context.Background()
	t1 := time.Now().Add(time.Hour * -24 * 30)
	t2 := t1.Add(time.Hour * 24)
	params := &CBGP{
		GameID: "32399",
		RequestParams: &RequestParams{
			TimeParams: &TimeParams{&t1, &t2},
		},
	}
	clips, err := testClient.ClipsByGame(ctx, params, 100)
	if err != nil {
		t.Fatal(err)
	}

	max := 3
	for i, c := range clips {
		if i == max {
			break
		}

		f, err := os.OpenFile("../out/"+c.ID+".mp4",
			os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		if err = f.Truncate(0); err != nil {
			t.Fatal(err)
		}

		fmt.Printf("Downloading clip %d/%d\n", i, max)
		if err := testClient.DownloadFromTwitch(ctx, &c, f); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("Completed download %d/%d\n", i, max)
	}
}
