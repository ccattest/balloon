package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"tubr/ffmpeg"
	"tubr/models"
	"tubr/twitch"

	"github.com/google/uuid"
)

type Processor struct{}
type ProcessParameters struct {
	GameID                string
	Duration              *int
	ClipLengthMax         *int
	DontRepeatBroadCaster bool
	TargetDuration        *int
	t1, t2                *time.Time
}

func (p *Processor) ParseParams(r *http.Request) (*ProcessParameters, error) {
	query := r.URL.Query()

	var duration *int
	if len(query["duration"]) > 0 {
		parsed, err := strconv.Atoi(query["duration"][0])
		if err != nil {
			return nil, err
		}
		duration = &parsed
	}

	t1, t2, err := parseTimeParams(query, duration)
	if err != nil {
		return nil, err
	}

	var clipTimeMax *int
	if len(query["clip_time_max"]) > 0 {
		parsed, err := strconv.Atoi(query["clip_time_max"][0])
		if err != nil {
			return nil, err
		}

		clipTimeMax = &parsed
	}

	var targetDuration *int
	if len(query["target_duration"]) > 0 {
		parsed, err := strconv.Atoi(query["target_duration"][0])
		if err != nil {
			return nil, err
		}
		targetDuration = &parsed
	} else {
		defaultDuration := 30
		targetDuration = &defaultDuration
	}

	gameIdQuery := query["game_id"]
	if len(gameIdQuery) != 1 {
		return nil, errors.New("failed to find game_id in query")
	}
	gameId := gameIdQuery[0]

	return &ProcessParameters{
		t1:             t1,
		t2:             t2,
		Duration:       duration,
		TargetDuration: targetDuration,
		ClipLengthMax:  clipTimeMax,
		GameID:         gameId,
	}, nil
}

func parseTimeParams(query url.Values, duration *int) (*time.Time, *time.Time, error) {
	var t1, t2 *time.Time
	if len(query["t1"]) > 0 {
		var err error
		t1, err = parseTime(query["t1"][0])
		if err != nil {
			return nil, nil, err
		}
	} else {
		return nil, nil, errors.New("error: missing t1 parameter")
	}

	if len(query["t2"]) > 0 {
		var err error
		t2, err = parseTime(query["t2"][0])
		if err != nil {
			return t1, nil, err
		}
	} else {
		log.Println("no t2 query parameter")
	}

	if t2 == nil && duration == nil {
		return t1, t2, errors.New("t2 and duration cannot both be empty")
	}

	if (t2 == nil) && (duration != nil) {
		add := time.Duration(*duration) * 24 * time.Hour
		t1Plus := t1.Add(add)
		t2 = &t1Plus
	}

	return t1, t2, nil
}

func GenerateTitle(t1, t2 *time.Time, gameId string) string {
	formatDate := func(t *time.Time) string {
		return t.Format("01/02/06")
	}

	gameName := twitch.GAME_LIST[gameId]

	return fmt.Sprintf("Top %s twitch Clips %s - %s", gameName, formatDate(t1), formatDate(t2))
}

type ProcessResult struct {
	ProcessId        string  `json:"process_id"`
	Errors           []error `json:"errors"`
	Duration         int     `json:"duration"`
	Length           int     `json:"compilation_length"`
	Description      string  `json:"description"`
	Title            string  `json:"title"`
	EncodedVideo     string  `json:"encoded_video"`
	EncodedThumbnail string  `json:"encoded_thumbnail"`
}

func process(ctx context.Context, params *ProcessParameters) (*ProcessResult, error) {
	processId := uuid.New().String()
	errors := []error{}

	apiParams := &twitch.CBGP{
		RequestParams: &twitch.RequestParams{
			TimeParams: &twitch.TimeParams{T1: params.t1, T2: params.t2},
			Cursor:     &twitch.Cursor{},
		},
		GameID: params.GameID,
	}

	clips, err := tc.ClipsByGame(ctx, apiParams, 1000)
	if err != nil {
		return nil, fmt.Errorf("error: failed to get clip data: %s", err)
	}
	fmt.Printf("Got %d clips\n", len(clips))

	compilation := []models.Clip{}

	if err = os.Mkdir(processId, 0644); err != nil {
		return nil, fmt.Errorf("error: failed to create compilation directory %s", err)
	}
	saveClipResponse(clips, processId)

	blacklist := append([]string{}, GLOBAL_BLACKLIST...)
	var duration = 0

PickClip:
	for _, clip := range clips {
		// check clip broadcaster name against blacklist
		for _, b := range blacklist {
			if b == clip.Broadcaster.Name {
				continue PickClip
			}
		}

		// don't repeat broadcaster in compilation
		if params.DontRepeatBroadCaster {
			blacklist = append(blacklist, clip.Broadcaster.Name)
		}

		// save clip
		path := fmt.Sprintf("%s/%s.mp4", processId, clip.ID)
		if err = downloadAndSave(ctx, path, &clip); err != nil {
			errors = append(errors, fmt.Errorf("error: failed to download clip: %s", err))
			continue
		}

		// tally
		duration += clip.Duration
		compilation = append(compilation, clip)

		if duration > *params.TargetDuration {
			break
		}
	}

	// generate path list
	paths := []string{}
	for _, c := range compilation {
		paths = append(paths, fmt.Sprintf("%s/%s.mp4", processId, c.ID))
	}

	// normalize and attribute clips
	for i, path := range paths {
		sp := strings.Split(path, ".mp4")
		sp[0] += "_norm.mp4"
		newPath := strings.Join(sp, "")
		credit := ffmpeg.CreditBox(compilation[i].Broadcaster.Name)
		args := ffmpeg.NormalizeArgs()
		args.AddVFilter(credit)
		if err := ffmpeg.FFmpeg(ctx, path, newPath, args, os.Stdout); err != nil {
			return nil, err
		}
		paths[i] = newPath
	}

	// generate normlized path list
	ids := []string{}
	for _, c := range compilation {
		ids = append(ids, c.ID+"_norm.mp4")
	}

	outFilePath := processId + "/compilation.mp4"

	// splice videoes together
	if err := ffmpeg.Concat(ctx, ids, outFilePath, os.Stdout); err != nil {
		return nil, fmt.Errorf("error: failed to concantenate: %s", err)
	}

	// read and encode compilation for response
	outFile, err := os.Open(outFilePath)
	if err != nil {
		return nil, fmt.Errorf("error: failed to open compilation path (%s): %s", outFilePath, err)
	}

	outFileData, err := io.ReadAll(outFile)
	if err != nil {
		return nil, fmt.Errorf("error: failed to read file into memory: %s", err)
	}

	encodedCompilation := base64.StdEncoding.EncodeToString(outFileData)

	log.Printf("Encoded file data: %d\n", len(encodedCompilation))

	thumbUrl := compilation[0].ThumbURL

	encodedThumbnail, err := downloadThumbnail(thumbUrl)
	if err != nil {
		errors = append(errors, fmt.Errorf("error: failed to download and encode thumbnail: %s", err))
	}

	return &ProcessResult{
		ProcessId:        processId,
		Errors:           errors,
		Duration:         duration,
		Length:           len(compilation),
		Description:      generateDescription(compilation),
		Title:            GenerateTitle(params.t1, params.t2, apiParams.GameID),
		EncodedVideo:     encodedCompilation,
		EncodedThumbnail: encodedThumbnail,
	}, nil
}

func downloadAndSave(ctx context.Context, path string, clip *models.Clip) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	if err = tc.DownloadFromTwitch(ctx, clip, file); err != nil {
		return err
	}

	return nil
}

func generateDescription(clips []models.Clip) string {
	description := ""
	pointInTime := 0

	normalize := func(d int) string {
		if d < 10 {
			return "0" + strconv.Itoa(d)
		}
		return strconv.Itoa(d)
	}

	for _, clip := range clips {
		description += fmt.Sprintf("%s %s (%s:%s)\n",
			clip.Title,
			clip.Broadcaster.Name,
			normalize(pointInTime/60),
			normalize(pointInTime%60))

		pointInTime += clip.Duration
	}
	return description
}

func parseTime(str string) (*time.Time, error) {
	t, err := time.Parse(time.DateOnly, str)
	return &t, err
}

func saveClipResponse(clips []models.Clip, processId string) {
	f, err := os.Create(processId + "/clips.json")
	if err != nil {
		panic(err)
	}

	if err = json.NewEncoder(f).Encode(clips); err != nil {
		panic(err)
	}
}

func downloadThumbnail(url string) (string, error) {
	thumbReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error: failed to create thumbnail request (%s): %s", url, err)
	}

	resp, err := tc.Do(thumbReq)
	if err != nil {
		return "", fmt.Errorf("error: failed to make thumbnail request: %s", err)
	}

	defer resp.Body.Close()
	thumbnailData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error: failed to read thumbnail data: %s", err)
	}
	return base64.StdEncoding.EncodeToString(thumbnailData), nil
}
