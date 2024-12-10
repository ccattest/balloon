package youtube

import (
	"context"
	"log"
	"os"

	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
	// "google.golang.org/api/youtube/v3"
)

// Client represents a youtube API client
type Client struct {
	*youtube.Service
	log  *log.Logger
	auth *oauth2.Config
}

func (c *Client) Insert(ctx context.Context, v *youtube.Video, f *os.File) error {
	res, err := c.Videos.Insert([]string{
		"snippet", "status", "id"}, v).
		Context(ctx).
		Media(f).
		Do()

	log.Println(res)
	return err
}
