package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"tubr/models"
)

var (
	GAME_LIST = map[string]string{}
)

type gr struct {
	Response
	Data []models.Game
}

// TopGames returns a slice of the top games
func (c *Client) TopGames(ctx context.Context) ([]models.Game, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/games/top?first=100", TwitchAPIURL), nil)
	if err != nil {
		return nil, err
	}

	return c.getGames(ctx, req, nil)
}

func (c *Client) Games(ctx context.Context, ids []string) ([]models.Game, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/games?first=100", TwitchAPIURL), nil)
	if err != nil {
		return nil, err
	}

	urlVals := &urlParams{req.URL.Query()}
	for _, v := range ids {
		urlVals.Add("id", v)
	}
	req.URL.RawQuery = urlVals.Encode()

	return c.getGames(ctx, req, nil)
}

func (c *Client) getGames(ctx context.Context, req *http.Request, games []models.Game) ([]models.Game, error) {
	if games == nil {
		games = []models.Game{}
	}

	select {
	case <-ctx.Done():
		return games, nil
	default:
	}

	fmt.Printf("requesting: %s\n", req.URL)
	r, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	response := &gr{}
	err = json.NewDecoder(r.Body).Decode(response)
	if err != nil {
		return nil, err
	}

	for _, g := range response.Data {
		games = append(games, models.Game{
			ID:        g.ID,
			Title:     g.Title,
			BoxArtURL: g.BoxArtURL,
		})
	}

	if response.Pagination.Cursor != "" {
		query, err := url.ParseQuery(req.URL.RawQuery)
		if err != nil {
			return games, fmt.Errorf("error: failed to parse query parameters: %s", err)
		}

		query.Set("after", response.Pagination.Cursor)
		req.URL.RawQuery = query.Encode()
		return c.getGames(ctx, req, games)
	}

	return games, nil
}

// ExportGames writes a game slice to "../game-list.json" in the form of JSON
func ExportGames(games []models.Game) error {
	existingGames, err := LoadGames()
	if err != nil {
		return err
	}
	storedGames := map[string]string{}

	for _, g := range existingGames {
		storedGames[g.ID] = g.Title
	}
	for _, g := range games {
		storedGames[g.ID] = g.Title
	}

	b, err := json.MarshalIndent(storedGames, "", "\t")
	if err != nil {
		return err
	}

	f, err := os.OpenFile("../game-list.json",
		os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	defer f.Close()
	if err := f.Truncate(0); err != nil {
		return err
	}

	_, err = f.Write(b)
	return err
}

// LoadGames loads game categories from JSON file
func LoadGames() ([]models.Game, error) {
	f, err := os.Open("../game-list.json")
	if err != nil {
		return nil, err
	}

	gameMap := map[string]string{}
	games := []models.Game{}
	if err := json.NewDecoder(f).Decode(&gameMap); err != nil {
		return nil, err
	}

	for id, title := range gameMap {
		games = append(games, models.Game{
			ID:    id,
			Title: title,
		})
	}

	return games, nil
}

func ImportGameList() {
	f, err := os.Open("./game-list.json")
	if err != nil {
		panic(err)
	}

	gameList := map[string]string{}
	if err = json.NewDecoder(f).Decode(&gameList); err != nil {
		panic(err)
	}

	fmt.Printf("Imported %d games from filesystem\n", len(gameList))
}
