//go:build windows
// +build windows

package ffmpeg

// import (
// 	"fmt"
// 	"io"
// 	"os/exec"
// )

// func callFFMpeg(in, out string, args *FFmpegArgs, stdout io.Writer) error {
// 	a := []string{BinPathWin + `ffmpeg.exe`, "-y"}
// 	if args.format != "" {
// 		a = append(a, []string{"-f", args.format}...)
// 	}
// 	a = append(a, []string{"-i", in}...)
// 	fmt.Println(args.otherArgs, args.vFilters)
// 	if len(args.vFilters) > 0 {
// 		vFilters := "-vf "
// 		for _, f := range args.vFilters {
// 			if vFilters == "-vf " {
// 				vFilters += f
// 			} else {
// 				vFilters += "," + f
// 			}
// 		}
// 		a = append(a, vFilters)
// 	}
// 	a = append(a, args.otherArgs...)
// 	a = append(a, out)

// 	cmd := exec.Command("powershell", a...)
// 	cmd.Stderr = stdout
// 	if err := cmd.Run(); err != nil {
// 		return err
// 	}

// 	return nil
// }
