package ffmpeg

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
)

const (
	// BinPathWin is the absolute path of ffmpeg binaries on windows
	BinPathWin = `.\bin\ffmpeg-4.2.1-win64-static\bin\`
	// BinPathArch is the absolute path of the binary on unix
	BinPathArch = `/usr/bin/`
)

// FFmpegArgs represents arguments for ffmpeg
type FFmpegArgs struct {
	format    string
	vFilters  []string
	otherArgs []string
}

func LoadAgenda(path string) ([]string, error) {
	f, err := os.Open(path + "/agenda.txt")
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)
	agenda := []string{}
	for scanner.Scan() {
		agenda = append(agenda, scanner.Text())
	}
	return agenda, nil
}

func buildCommand(ctx context.Context, in, out string, opts *FFmpegArgs, logger io.Writer) *exec.Cmd {
	args := []string{}
	var command string
	if runtime.GOOS == "windows" {
		command = "powershell"
		args = append(args, BinPathWin+`ffmpeg.exe`)
	} else {
		command = "ffmpeg"
	}

	if opts.format != "" {
		args = append(args, "-f", opts.format)
	}

	args = append(args, "-y", "-i", in)

	if len(opts.vFilters) > 0 {
		vf := ``
		for i, f := range opts.vFilters {
			if i != len(opts.vFilters) && vf != `` {
				vf += `,`
			}
			vf += f
		}
		args = append(args, "-vf", vf)
	}

	args = append(args, out)

	cmd := exec.CommandContext(ctx, command, args...)
	fmt.Println(cmd)
	cmd.Stdout = logger
	cmd.Stderr = logger
	return cmd
}

// FFmpeg calls ffmpeg
func FFmpeg(ctx context.Context, in, out string, opts *FFmpegArgs, logger io.Writer) error {
	return buildCommand(ctx, in, out, opts, logger).Run()
}

// NormalizeArgs returns an object with normalize filters
func NormalizeArgs() *FFmpegArgs {
	return &FFmpegArgs{
		vFilters: []string{
			FilterFormat("yuv420p"),
			FilterScale(1280, 720),
			FilterFrameRate(30),
		},
		otherArgs: []string{
			"-vcodec", "libx264",
			"-acodec", "aac",
		},
	}
}

// AddVFilter adds a video filter to the vFilters slice
func (args *FFmpegArgs) AddVFilter(f interface{}) *FFmpegArgs {
	switch fil := f.(type) {
	case string:
		args.vFilters = append(args.vFilters, fil)
	case []string:
		args.vFilters = append(args.vFilters, fil...)
	default:
		panic(errors.New("error: invalid input to AddVFilter, must be a string or []string"))
	}

	return args
}
