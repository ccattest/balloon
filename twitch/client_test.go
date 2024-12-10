package twitch

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var mockMap = map[string]string{
	"/games": GamesMock,
	"/clips": ClipsMock,
}

type Interceptor struct{}

func (i Interceptor) RoundTrip(r *http.Request) (*http.Response, error) {
	respBody := mockMap[r.URL.Path]
	fmt.Println(respBody)

	body := bytes.NewReader([]byte(respBody))

	return &http.Response{
		Status:        "200 OK",
		StatusCode:    http.StatusOK,
		Body:          ioutil.NopCloser(body),
		ContentLength: int64(body.Len()),
		Close:         true,
		Request:       r,
	}, nil
}

var testClient = NewClient(
	&http.Client{
		Transport: Interceptor{},
	},
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	err := testClient.Authenticate(ctx)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}

func TestAuthenticated(t *testing.T) {
	assert.NotEmpty(t, testClient.AccessToken)
	assert.Greater(t,
		testClient.TokenExpiration.Unix(),
		time.Now().Unix())
}
