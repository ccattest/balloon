package store

import (
	"context"

	"tubr/models" // Your models package
)

// InsertClip inserts a new clip into the database
func (db *DB) InsertClip(ctx context.Context, c *models.Clip) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO clips (
			id,
			video_id,
			game_id,
			lang,
			title,
			broadcaster,
			clipper,
			view_count,
			clip_date
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET view_count = $8`,
		c.ID, c.VideoID, c.GameID,
		c.Lang, c.Title, c.Broadcaster.ID,
		c.Clipper.ID, c.ViewCount, c.ClipDate)

	if err != nil {
		return err
	}

	return func(users []models.User) error {
		for _, u := range users {
			_, err = db.ExecContext(ctx, `
				INSERT INTO users (id, name)
				VALUES ($1, $2)
				ON CONFLICT (id) DO UPDATE SET name = $2`,
				u.ID, u.Name)
			if err != nil {
				return err
			}
		}
		return nil
	}([]models.User{c.Broadcaster, c.Clipper})
}

// InsertUser inserts a user model into the database
func (db *DB) InsertUser(ctx context.Context, u *models.User) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, name)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, u.ID, u.Name)

	return err
}

// InsertGame inserts a game model into the database
func (db *DB) InsertGame(ctx context.Context, g *models.Game) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO games (id, title)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, g.ID, g.Title)

	return err
}
