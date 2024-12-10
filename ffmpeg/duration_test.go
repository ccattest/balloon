package ffmpeg

import (
	"fmt"
	"testing"
	"time"
)

func TestAddition(t *testing.T) {
	curDuration := time.Duration(100) * time.Second
	fmt.Println(curDuration.Seconds())
}