package ffmpeg

import (
	"context"
	"fmt"
	"io"
	"os"
)

// generateFileList generates a file list for the ffmpeg -concat operator
func generateFileList(path string, vidPaths []string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	if err = f.Truncate(0); err != nil {
		return err
	}

	defer f.Close()
	for _, p := range vidPaths {
		_, err = f.Write([]byte(fmt.Sprintf("file '%s'\n", p)))
		if err != nil {
			return err
		}
	}
	return nil
}

// Concat xxx
func Concat(ctx context.Context, vidPaths []string, outPath string, output io.Writer) error {
	fileListPath := outPath + ".txt"
	if err := generateFileList(fileListPath, vidPaths); err != nil {
		return err
	}

	return FFmpeg(
		ctx,
		fileListPath,
		outPath,
		&FFmpegArgs{
			format: "concat",
		},
		output,
	)
}
