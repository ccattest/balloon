package download

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ClipseyDownloader struct {
	*http.Client
}

func NewClipsey() *ClipseyDownloader {
	return &ClipseyDownloader{}
}

func (cd *ClipseyDownloader) DownloadClip(slug string) (io.ReadCloser, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("https://cy49zmt23f.execute-api.us-east-1.amazonaws.com/dev/download_clip?id=%s", slug),
		nil,
	)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	response := struct {
		Data []struct {
			VideoUrl string `json:"video_url"`
		} `json:"data"`
	}{}

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	clipRequest, err := http.NewRequest("GET", response.Data[0].VideoUrl, nil)
	if err != nil {
		return nil, err
	}
	clipResponse, err := client.Do(clipRequest)
	if err != nil {
		return nil, err
	}

	return clipResponse.Body, nil
}
