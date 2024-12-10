package ffmpeg

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// VideoDuration gets a videos duration in seconds 
func VideoDuration(ctx context.Context, path string) (int, error) {
	var binPath string
	if runtime.GOOS == "windows" {
		binPath = BinPathWin + "ffprobe.exe"
	} else if runtime.GOOS == "linux" {
		binPath = BinPathArch + "ffprobe"
	} else {
		return 0, fmt.Errorf("error: unknown platform")
	}
			
	args := []string{"-v", "error", "-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1", path}

	output, err := exec.Command(binPath, args...).Output()
	if err != nil {
		return -1, err
	}

	pieces := strings.Split(string(output[:len(output)-1]), ".")
	seconds, err := strconv.Atoi(pieces[0])
	if err != nil {
		return -1, errors.Wrap(err, "error: failed to parse duration")
	}
	return seconds, nil
}