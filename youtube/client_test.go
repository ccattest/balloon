package youtube

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var testClient = &Client{}

func TestMain(m *testing.M) {
	f, err := os.Open("./client_secret.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := google.ConfigFromJSON(b, youtube.YoutubeUploadScope)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	tokenDir := "./token.json"
	tok, err := tokenFromFile(tokenDir)
	if err != nil {
		log.Println("could not get cached token")
		authUrl := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		fmt.Printf("auth url: %s\n", authUrl)
		fifoPath := "/tmp/auth-fifo"
		fmt.Printf("send auth code to %s now\n", fifoPath)
		var code string
		authPipe, err := os.Open(fifoPath)
		if err != nil {
			log.Fatalf("error: failed to open auth pipe %v\n",
				err)
		}
		s := bufio.NewScanner(authPipe)
		for s.Scan() {
			code = s.Text()
		}
		tok, err = cfg.Exchange(ctx, code)
		if err != nil {
			log.Fatalf("failed to exchange token %v\n", err)
		}

		err = saveToken(tokenDir, tok)
		if err != nil {
			log.Printf("failed to save token %v\n", err)
		}
	}

	yts, err := youtube.NewService(ctx,
		option.WithHTTPClient(cfg.Client(ctx, tok)))
	if err != nil {
		log.Fatal(err)
	}

	testClient = &Client{
		Service: yts,
		log:     &log.Logger{},
	}
	os.Exit(m.Run())
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

func saveToken(file string, token *oauth2.Token) error {
	f, err := os.OpenFile(file,
		os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

func TestInsert(t *testing.T) {
	ctx := context.Background()

	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       "test upload",
			Description: "test description",
			CategoryId:  "23",
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: "unlisted",
		},
	}

	f, err := os.Open("./test_data/test_upload.mp4")
	if err != nil {
		t.Fatal(err)
	}

	if err = testClient.Insert(ctx, upload, f); err != nil {
		t.Fatal(err)
	}
}

func TestInput(t *testing.T) {
	f, err := os.Open("/tmp/auth-fifo")
	if err != nil {
		t.Fatal(err)
	}

	s := bufio.NewScanner(f)
	fmt.Println(s.Text())
}
