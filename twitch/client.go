package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"tubr/models"
	"tubr/twitch/download"
)

const (
	// TwitchClientID is the api key for Twitch's Helix API
	TwitchClientID = "33eafqtezsbpqdhjoccbk0msft0b5g"
	// TwitchClientSecret is the client secret for OAuth
	TwitchClientSecret = "phubcqun90q5edawwgrmca5hnoeklz"
	// TwitchAPIURL is the url for Twitch's Helix API
	TwitchAPIURL = "https://api.twitch.tv/helix"
	// TwitchDownloadURL is a url used for downloading clips
	TwitchDownloadURL = "https://clips-media-assets2.twitch.tv/"
	// TwitchOauthURL
	TwitchOauthURL = "https://id.twitch.tv/oauth2/token"
	// OauthCachePath
	OauthCachePath = "oauth.json"

	ErrBadResponseCode = "error: bad response %d"
)

// Client is a Twitch Helix Client
type Client struct {
	*http.Client
	AccessToken     string
	RefreshToken    string
	TokenExpiration time.Time
	BadTokenCache   bool
}

type tokenResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	Expiration   int      `json:"expires_in"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}

// NewClient creates a new Client
func NewClient(c *http.Client) *Client {
	var client *http.Client

	if c != nil {
		client = c
	}

	if c == nil {
		client = &http.Client{
			Timeout: time.Second * 10,
		}
	}

	return &Client{
		Client: client,
	}
}

type urlParams struct {
	url.Values
}

type Cursor struct {
	After, Before string
}

type Response struct {
	Pagination struct {
		Cursor string `json:"cursor"`
	} `json:"pagination"`
}

func (u *urlParams) tokenRequestParams(refresh bool, scopes []string) *urlParams {
	u.Add("client_id", TwitchClientID)
	u.Add("client_secret", TwitchClientSecret)
	grantType := "client_credentials"
	if refresh {
		grantType = "refresh_token"
	}
	u.Add("grant_type", grantType)
	if scopes != nil {
		u.Add("scope", strings.Join(scopes, " "))
	}
	return u
}

func (u *urlParams) cursorRequestParams(p *Cursor) *urlParams {
	if p.After != "" {
		u.Add("after", p.After)
	} else if p.Before != "" {
		u.Add("before", p.Before)
	}
	return u
}

func (u *urlParams) timeRequestParams(p *TimeParams) *urlParams {
	// ignore hanging as api will drop
	if (p.T2 != nil && p.T1 == nil) || (p.T1 == nil && p.T2 == nil) {
		return u
	} else if p.T1 != nil && p.T2 == nil {
		newT2 := (*p.T1).Add(time.Hour * 24 * 7)
		p.T2 = &newT2
	}

	u.Add("started_at", p.T1.UTC().Format(time.RFC3339))
	u.Add("ended_at", p.T2.UTC().Format(time.RFC3339))
	return u
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Client-ID", TwitchClientID)
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Accept", "application/json")

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == 401 {
		c.BadTokenCache = true
		fmt.Println("retrying authentication")
		if err = c.Authenticate(req.Context()); err != nil {
			fmt.Printf("error: failed to reauthenticate %s\n", err)
			return res, err
		}

		req.Header.Set("Authorization", "Bearer "+c.AccessToken)

		res, err = c.Client.Do(req)
		if err != nil {
			return res, err
		}
		fmt.Println("status code for retry", res.StatusCode)
	}

	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		res.Body.Close()
		fmt.Println(string(b))
		return res, fmt.Errorf(ErrBadResponseCode, res.StatusCode)
	}

	return res, nil
}

func (c *Client) loadOathFromFs() error {
	f, err := os.Open(OauthCachePath)
	if err != nil {
		return fmt.Errorf("failed to open oauth cache path %s: %s", OauthCachePath, err)
	}

	token := &tokenResponse{}
	if err = json.NewDecoder(f).Decode(token); err != nil {
		return fmt.Errorf("error: failed to decode cached oauth token: %s", err)
	}

	c.loadOauth(token)
	return nil
}

func (c *Client) loadOauth(token *tokenResponse) {
	c.TokenExpiration = time.Now().
		Add(time.Duration(token.Expiration) * time.Second)
	c.AccessToken = token.AccessToken
	c.RefreshToken = token.RefreshToken
}

func (c *Client) RequestAndSave(in, out string) error {
	f, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", in, nil)
	if err != nil {
		return err
	}

	r, err := c.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	_, err = io.Copy(f, r.Body)
	return err
}

// Authenticate gets and sets the access token for the client
func (c *Client) Authenticate(ctx context.Context) error {
	if !c.BadTokenCache {
		if err := c.loadOathFromFs(); err != nil {
		} else {
			return nil
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		TwitchOauthURL, nil)
	if err != nil {
		return err
	}

	params := &urlParams{url.Values{}}
	params.tokenRequestParams(false, nil)
	req.URL.RawQuery = params.Encode()

	res, err := c.Client.Do(req)
	if err != nil {
		return err
	} else if res.StatusCode != 200 {
		return fmt.Errorf("error: receieved response code %d",
			res.StatusCode)
	}
	defer res.Body.Close()

	token := &tokenResponse{}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(body, token); err != nil {
		return err
	}

	_ = cacheToken(token)
	c.loadOauth(token)

	return nil
}

func cacheToken(tokenData *tokenResponse) error {
	oauthFile, err := os.OpenFile(OauthCachePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("error: failed to open oauth file location for caching: %s", err)
	} else {
		defer oauthFile.Close()
		if err = json.NewEncoder(oauthFile).Encode(tokenData); err != nil {
			fmt.Printf("error: failed to cache oauth file: %s", err)
		}
	}
	return nil
}

func (c *Client) Refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "POST", TwitchOauthURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.Client.Do(req)
	if err != nil {
		return err
	} else if res.StatusCode != 200 {
		defer res.Body.Close()
		response, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("error: failed to read %d response body", res.StatusCode)
		}
		return fmt.Errorf("error: failed to refresh token %s code %d %s", string(response), res.StatusCode, err)
	}

	token := &tokenResponse{}
	if err = json.NewDecoder(res.Body).Decode(token); err != nil {
		return err
	}

	_ = cacheToken(token)
	c.loadOauth(token)

	return nil
}

// DownloadFromTwitch downloads a clip and writes it to out, and closes out
func (c *Client) DownloadFromTwitch(
	ctx context.Context, clip *models.Clip, out io.WriteCloser) error {
	defer out.Close()

	var clipData io.ReadCloser
	var err error
	fmt.Println("url Version", clip.ThumbnailUrlVersion)
	switch clip.ThumbnailUrlVersion {
	case "1":
		req, err := http.NewRequestWithContext(ctx, "GET", clip.VideoUrl, nil)
		if err != nil {
			return err
		}
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			return err
		}
		clipData = resp.Body

	case "2":
		cd := &download.ClipseyDownloader{}
		clipResponse, err := cd.DownloadClip(clip.ID)
		if err != nil {
			return err
		}

		clipData = clipResponse
	}

	fmt.Println("out", out)
	_, err = io.Copy(out, clipData)
	return err
}
