package ffmpeg

import (
	"bytes"
	"context"
	"testing"
)

var testSlugs = []string{
	"RepleteEnthusiasticMageCeilingCat",
	"CalmImpartialThymeOptimizePrime",
	"SteamyGlamorousGoshawkCopyThis",
	"BoringSoftWaspSuperVinlin",
	"FaintWildSoymilkLitty",
}

func TestCommand(t *testing.T) {
	ctx := context.Background()
	output := bytes.NewBuffer([]byte{})
	cmd := buildCommand(ctx, "../out/RepleteEnthusiasticMageCeilingCat.mp4", "example.mkv", &FFmpegArgs{
		vFilters: []string{"scale=1280x720"},
	}, output)

	t.Log(cmd.String())
}

// Windows
//
//	4.2.1 ok
func TestNormalize(t *testing.T) {
	ctx := context.Background()
	output := bytes.NewBuffer([]byte{})
	for _, slug := range testSlugs {
		ffmpegArgs := NormalizeArgs()
		err := FFmpeg(
			ctx,
			"../out/"+slug+".mp4",
			"./test_results/"+slug+".mkv",
			ffmpegArgs,
			output,
		)

		t.Log(output.String())
		if err != nil {
			t.Error(err)
		}
	}
}

func mapSlugsToConvertedPath() []string {
	c := []string{}
	for _, s := range testSlugs {
		c = append(c, "file test_results/"+s+".mkv")
	}
	return c
}

func TestConcat(t *testing.T) {
	ctx := context.Background()
	slugs := mapSlugsToConvertedPath()
	output := bytes.NewBuffer([]byte{})

	if err := generateFileList("./concat.txt", slugs); err != nil {
		t.Error(err)
	}

	err := FFmpeg(ctx, "./concat.txt", "./test_results/concat-test.mkv", &FFmpegArgs{
		format: "concat",
	}, output)

	t.Log(output.String())
	if err != nil {
		t.Error(err)
	}
}
