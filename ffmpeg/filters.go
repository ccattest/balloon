package ffmpeg

import (
	"fmt"
)

// FilterScale maps w, h to an ffmpeg scale filter
func FilterScale(w, h int) string {
	return fmt.Sprintf("scale=%dx%d", w, h)
}

// FilterFormat maps format to anj ffmpeg format filter
func FilterFormat(format string) string {
	return "format=pix_fmts=" + format
}

// FilterFrameRate maps fps to ffmpeg fps filter
func FilterFrameRate(fps int) string {
	return fmt.Sprintf("fps=fps=%d", fps)
}

// FilterDrawBox draws a box over the stream
func FilterDrawBox(x, y, w, h int, c, o string) string {
	if o == "" {
		o = "fill"
	}

	return fmt.Sprintf(
		"drawbox=x=%d:y=%d:w=%d:h=%d:color=%s@1:t=%s",
		x, y, w, h, c, o)
}

// FilterDrawText ...
func FilterDrawText(x, y int, t string) string {
	return fmt.Sprintf(
		`drawtext=x=%d:y=%d:text="%s":font=Inconsolata:fontsize=22`,
		x, y, t)
}
