package store

import (
	"context"
	"log"
	"tubr/models"
)

func (db *DB) InsertVideos(ctx context.Context, videos []models.Video) error {
	for _, v := range videos {
		var insertVideos = `
			INSERT INTO videos (
				schedule_id,
				start_date,
				end_date,
				release_date,
				cache_destination,
				uploaded
			) VALUES (
				$1, $2, $3, $4, $5, $6
			)
		`

		_, err := db.ExecContext(ctx, insertVideos,
			v.ScheduleID,
			v.StartDate,
			v.EndDate,
			v.ReleaseDate,
			v.CacheDestination,
			v.Uploaded,
		)
		if err != nil {
			log.Printf("Failed to insert video %+v %s", v, err)
			continue
		}
	}
	return nil
}

func (db *DB) GetVideosByScheduleId(ctx context.Context, id int) ([]models.Video, error) {
	var selectVideosByScheduleId = `
		SELECT
			schedule_id,
			start_date,
			end_date,
			release_date,
			cache_destination,
			uploaded
		FROM
			videos
		WHERE schedule_id = $1
		ORDER BY start_date DESC
	`

	r, err := db.QueryContext(ctx, selectVideosByScheduleId, id)
	if err != nil {
		return nil, err
	}

	videos := []models.Video{}
	for r.Next() {
		video := models.Video{}
		if err = r.Scan(
			&video.ScheduleID,
			&video.StartDate,
			&video.EndDate,
			&video.ReleaseDate,
			&video.CacheDestination,
			&video.Uploaded,
		); err != nil {
			return videos, err
		}

		videos = append(videos, video)
	}

	return videos, nil
}

func (db *DB) CountVideosByScheduleID(ctx context.Context, id int) (int, error) {
	r := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM videos
		WHERE schedule_id = $1
	`, id)

	var count int
	err := r.Scan(&count)

	return count, err
}
