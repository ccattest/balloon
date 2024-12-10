package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"tubr/models"
)

var selectClips = `
SELECT DISTINCT
	clips.id,
	clips.video_id,
	clips.game_id,
	clips.lang,
	clips.title,
	clips.view_count,
	clips.clip_date,
	broadcaster.id,
	broadcaster.name,
	clipper.id,
	clipper.name 
FROM clips
	LEFT JOIN users AS broadcaster ON (clips.broadcaster = broadcaster.id)
	LEFT JOIN users AS clipper ON (clips.clipper = clipper.id)
`

var selectGames = `
SELECT id, title
FROM games
`

// ClipQueryArgs is a struct for passing arguments to db queries on the clips table
type ClipQueryArgs struct {
	Limit         int
	BroadcasterID string
	GameID        string
	ClipperID     string
	T1, T2        *time.Time
	Offset        int
	ID            string
}

// DefaultClipQueryArgs returns a default struct of args
func DefaultClipQueryArgs() *ClipQueryArgs {
	return &ClipQueryArgs{
		Limit:         100,
		BroadcasterID: "",
		ClipperID:     "",
		T1:            nil, T2: nil,
		Offset: 0,
	}
}

func (args *ClipQueryArgs) buildQuery(query string) (string, []interface{}) {
	if args == nil {
		return selectClips, []interface{}{}
	}
	psqlArgs := []interface{}{}
	where := false

	// newArg adds an arg to the list of conditional parameters
	// with an associated comparison
	newArg := func(arg interface{}, comp string) {
		if arg == nil {
			return
		}
		switch arg.(type) {
		case string:
			x, ok := arg.(string)
			if !ok || x == "" {
				return
			}
		case *time.Time:
			x, ok := arg.(*time.Time)
			if !ok || x == nil || x.Equal(time.Time{}) {
				return
			}
			arg = x.Format(time.RFC3339)
		}

		if !where {
			query += "WHERE "
			where = true
		}

		psqlArgs = append(psqlArgs, arg)
		if len(psqlArgs) > 1 {
			query += " AND \n\t"
		}
		query += fmt.Sprintf("%s $%d", comp, len(psqlArgs))
	}

	newArg(args.BroadcasterID, "broadcaster =")
	newArg(args.ClipperID, "clipper =")
	newArg(args.GameID, "game_id =")
	newArg(args.ID, "id =")
	newArg(args.T1, "clip_date >=")
	newArg(args.T2, "clip_date <=")
	query += fmt.Sprintf("\nORDER BY clip_date ASC")
	query += fmt.Sprintf("\nLIMIT %d OFFSET %d", args.Limit, args.Offset)

	return query, psqlArgs
}

// SelectClips .
func (db *DB) SelectClips(ctx context.Context, args *ClipQueryArgs) ([]models.Clip, error) {
	query, qArgs := args.buildQuery(selectClips)

	log.Printf("running query\n%s\n", query)

	r, err := db.QueryContext(ctx, query, qArgs...)
	if err != nil {
		return nil, err
	}

	return scanClips(r)
}

func scanClips(r *sql.Rows) ([]models.Clip, error) {
	clips := []models.Clip{}
	for r.Next() {
		c := models.Clip{
			Broadcaster: models.User{},
			Clipper:     models.User{},
		}
		if err := r.Scan(
			&c.ID,
			&c.VideoID,
			&c.GameID,
			&c.Lang,
			&c.Title,
			&c.ViewCount,
			&c.ClipDate,
			&c.Broadcaster.ID,
			&c.Broadcaster.Name,
			&c.Clipper.ID,
			&c.Clipper.Name,
		); err != nil {
			return clips, err
		}
		clips = append(clips, c)
	}
	return clips, nil
}
