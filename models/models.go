package models

import (
	"time"
)

// User represents a user on the Twitch platform
type User struct {
	ID   string
	Name string
}

// Clip represents a clip from the API
type Clip struct {
	ID                  string
	URL                 string
	Broadcaster         User
	Clipper             User
	VideoID             string
	GameID              string
	Lang                string
	Title               string
	ViewCount           int
	ThumbURL            string
	ClipDate            time.Time
	Duration            int
	VideoUrl            string
	ThumbnailUrlVersion string
}

// Game represents a game category
type Game struct {
	ID        string `json:"id"`
	Title     string `json:"name"`
	BoxArtURL string `json:"box_art_url"`
}

type Schedule struct {
	ID                    int       `json:"id"`
	Label                 string    `json:"label"`
	GameID                string    `json:"game_id"`
	Broadcaster           string    `json:"broadcaster"`
	Platform              string    `json:"platform"`
	Language              string    `json:"language"`
	BroadcasterID         string    `json:"broadcaster_id"`
	StartDate             time.Time `json:"start_date"`
	FrequencyDays         int       `json:"frequency_days"`
	ClipTimeMaxSeconds    int       `json:"clip_time_max_seconds"`
	RepeatBroadcaster     bool      `json:"repeat_broadcaster"`
	TargetDurationSeconds int       `json:"target_duration_seconds"`
	WebhookURL            string    `json:"webhook_url"`
}

type Video struct {
	ID               string     `json:"id"`
	ScheduleID       int        `json:"schedule_id"`
	StartDate        *time.Time `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	ReleaseDate      *time.Time `json:"release_date"`
	CacheDestination *string    `json:"cache_destination"`
	Uploaded         bool       `json:"bool"`
}
