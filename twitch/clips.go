package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"tubr/models"
)

type ContextKey int

const (
	ContextLimit ContextKey = iota
)

func clipRequest(ctx context.Context, url string, p *RequestParams) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if p != nil {
		u := &urlParams{req.URL.Query()}
		u.
			timeRequestParams(p.TimeParams).
			cursorRequestParams(p.Cursor)
		req.URL.RawQuery = u.Encode()
	}
	return req, nil
}

// ClipByID returns a clip from the API with the corresponding clip ID
func (c *Client) ClipByID(ctx context.Context, id string) ([]models.Clip, error) {
	req, err := clipRequest(ctx,
		fmt.Sprintf("%s/clips?id=%s", TwitchAPIURL, id), nil)
	if err != nil {
		return nil, err
	}

	return c.getClips(ctx, req, nil)
}

type TimeParams struct {
	T1, T2 *time.Time
}

type RequestParams struct {
	*TimeParams
	*Cursor
	Limit int
}

type CBGP struct {
	*RequestParams
	GameID string
}

// ClipsByGame returns a slice of clips corresponding to game ID and time range
func (c *Client) ClipsByGame(ctx context.Context, params *CBGP, limit int) ([]models.Clip, error) {
	ctxWithLimit := context.WithValue(ctx, ContextLimit, limit)
	url := fmt.Sprintf("%s/clips?game_id=%s&first=100",
		TwitchAPIURL, params.GameID)

	log.Printf("Getting clips by game %+v", params)

	req, err := clipRequest(ctxWithLimit, url, params.RequestParams)
	if err != nil {
		return nil, err
	}
	return c.getClips(ctxWithLimit, req, nil)
}

type CBBP struct {
	BroadcasterID string
	*RequestParams
	*TimeParams
}

func (c *Client) ClipsByBroadcaster(ctx context.Context, params *CBBP, limit int) ([]models.Clip, error) {
	url := fmt.Sprintf("%s/clips?broadcaster_id=%s&first=20",
		TwitchAPIURL, params.BroadcasterID)

	req, err := clipRequest(ctx, url, params.RequestParams)
	if err != nil {
		return nil, err
	}
	return c.getClips(ctx, req, nil)
}

// mapThumbnailUrlToVideoUrl maps a clip URL to a download URL
// Pre 2024 format
// https://static-cdn.jtvnw.net/twitch-clips-thumbnails-prod/AcceptableCulturedShieldLitFam-YkUH3nysLVrZRf2h/6b146dd5-c33d-44a5-bf63-75badc996804/preview-480x272.jpg
// to
// https://production.assets.clips.twitchcdn.net/v2/media/TastyLazyArmadilloTebowing-td_laoNjyqYo7SrQ/6897dd62-79db-4c0a-b24f-4c255495a575/video.mp4?sig=5f23f0f5d2d8c43da49e8cb88eb48c162af98704&token={"authorization":{"forbidden":false,"reason":""},"clip_uri":"","clip_slug":"TastyLazyArmadilloTebowing-td_laoNjyqYo7SrQ","device_id":null,"expires":1728927003,"user_id":"","version":2}
func mapThumbnailUrlToVideoUrl(uri string) string {
	// Post 2024
	if strings.Contains(uri, "twitch-clips-thumbnails-prod") {
		urlSlice := "-thumbnails-prod"
		start := strings.Index(uri, urlSlice)
		end := strings.Index(uri, "/preview")
		slugCombo := uri[start+len(urlSlice)+1 : end]

		pre := "https://production.assets.clips.twitchcdn.net/v2/media/"

		slug := slugCombo[:strings.Index(slugCombo, "/")]

		token := fmt.Sprintf(`{"authorization":{"forbidden":false,"reason":""},"clip_uri":"","clip_slug":"%s","device_id":null,"expires":1728927003,"user_id":"","version":2}`, slug)
		qp := fmt.Sprintf(`sig=5f23f0f5d2d8c43da49e8cb88eb48c162af98704e&token=%s`, url.QueryEscape(token))
		videoUrl := pre + slugCombo + "/video.mp4" + "?" + qp
		//https://production.assets.clips.twitchcdn.net/v2/media/ImpartialHorribleDogDogFace-7pnHAj9gUuDtdDxr/307d5d75-a0df-4aa8-9db3-015dd864514b/video.mp4%3Fsig%3D5f23f0f5d2d8c43da49e8cb88eb48c162af98704%26token%3D%7B%22authorization%22%3A%7B%22forbidden%22%3Afalse%2C%22reason%22%3A%22%22%7D%2C%22clip_uri%22%3A%22%22%2C%22clip_slug%22%3A%22%2FImpartialHorribleDogDogFace-7pnHAj9gUuDtdDxr%2F307d5d75-a0df-4aa8-9db3-015dd864514b%22%2C%22device_id%22%3Anull%2C%22expires%22%3A1728927003%2C%22user_id%22%3A%22%22%2C%22version%22%3A2%7D
		return videoUrl
	}
	videoUrl := uri[:strings.Index(uri, "-preview")] + ".mp4"
	return videoUrl
}

func (c *Client) getClips(ctx context.Context, req *http.Request, clips []models.Clip) ([]models.Clip, error) {
	if clips == nil {
		clips = []models.Clip{}
	}

	select {
	case <-ctx.Done():
		return clips, nil
	default:
	}

	if len(clips) > ctx.Value(ContextLimit).(int) {
		return clips, nil
	}

	fmt.Printf("requesting: %s\n", req.URL)
	r, err := c.Do(req)
	if err != nil {
		return clips, err
	}

	defer r.Body.Close()

	response := struct {
		Data []struct {
			ID              string    `json:"id"`
			URL             string    `json:"url"`
			EmbedURL        string    `json:"embed_url"`
			BroadcasterID   string    `json:"broadcaster_id"`
			BroadcasterName string    `json:"broadcaster_name"`
			CreatorID       string    `json:"creator_id"`
			CreatorName     string    `json:"creator_name"`
			VideoID         string    `json:"video_id"`
			GameID          string    `json:"game_id"`
			Language        string    `json:"language"`
			Title           string    `json:"title"`
			ViewCount       int       `json:"view_count"`
			CreatedAt       time.Time `json:"created_at"`
			ThumbnailURL    string    `json:"thumbnail_url"`
			Duration        float32   `json:"duration"`
		} `json:"data"`
		Pagination struct {
			Cursor string `json:"cursor"`
		} `json:"pagination"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		return clips, err
	}

	for _, c := range response.Data {
		// deduplicate clips returned based on brodadcaster ID and CreatedAt
		duplicate := false
		for _, o := range clips {
			if c.BroadcasterID == o.Broadcaster.ID {
				var threshold float64 = 180
				if math.Abs(float64(o.ClipDate.Sub(c.CreatedAt))) < threshold {
					fmt.Printf("%s and %s are within %0f of eachother, skipping %s\n", c.ID, o.ID, threshold, o.ID)
					duplicate = true
				}
			}
		}

		if duplicate {
			continue
		}

		v2Thumbnail := strings.Contains(c.ThumbnailURL, "twitch-clips-thumbnails-prod")
		thumbnailVersion := "1"
		if v2Thumbnail {
			thumbnailVersion = "2"
		}
		clips = append(clips, models.Clip{
			ID:  c.ID,
			URL: c.URL,
			Broadcaster: models.User{
				ID:   c.BroadcasterID,
				Name: c.BroadcasterName,
			},
			Clipper: models.User{
				ID:   c.CreatorID,
				Name: c.CreatorName,
			},
			VideoID:             c.VideoID,
			GameID:              c.GameID,
			Lang:                c.Language,
			Title:               c.Title,
			ViewCount:           c.ViewCount,
			ClipDate:            c.CreatedAt,
			ThumbURL:            c.ThumbnailURL,
			VideoUrl:            mapThumbnailUrlToVideoUrl(c.ThumbnailURL),
			Duration:            int(c.Duration),
			ThumbnailUrlVersion: thumbnailVersion,
		})
	}

	if response.Pagination.Cursor != "" {
		query, err := url.ParseQuery(req.URL.RawQuery)
		if err != nil {
			return clips, fmt.Errorf("failed to parse query parameters: %s", err)
		}

		query.Set("after", response.Pagination.Cursor)
		req.URL.RawQuery = query.Encode()
		return c.getClips(ctx, req, clips)
	}

	return clips, nil
}
