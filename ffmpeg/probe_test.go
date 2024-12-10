package ffmpeg

import (
	"context"
	"fmt"
	"testing"
)

func TestVideoDuration(t *testing.T) {
	ctx := context.Background()
	for _, slug := range testSlugs {
		dur, err := VideoDuration(ctx, "../out/"+slug+".mp4")
		if err != nil {
			t.Error(err)
		}

		fmt.Printf("Got duration: %d seconds for slug ID %s\n", dur, slug)
	}
}