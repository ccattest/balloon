package ffmpeg

const (
	creditOffsetX        = 00
	creditOffsetY        = 000
	creditHeight         = 30
	creditBackForeOffset = 5
)

// CreditBox places the credit box
func CreditBox(text string) []string {
	textWidth := len(text)*15 + 30
	return []string{
		CreditBackground(textWidth),
		CreditForeground(textWidth),
		CreditText(text),
	}
}

// CreditBackground is the background of the credit
func CreditBackground(w int) string {
	return FilterDrawBox(
		creditOffsetX,
		creditOffsetY,
		w,
		creditHeight,
		"SlateGray",
		"0",
	)
}

// CreditForeground is the foreground of the credit
func CreditForeground(w int) string {
	return FilterDrawBox(
		creditOffsetX+creditBackForeOffset,
		creditOffsetY+creditBackForeOffset,
		w,
		creditHeight,
		"Gray",
		"",
	)
}

// CreditText is the text of the credit
func CreditText(text string) string {
	return FilterDrawText(
		creditOffsetX+creditBackForeOffset*3,
		creditOffsetY+creditBackForeOffset*3,
		text)
}
