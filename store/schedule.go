package store

import (
	"context"
	"database/sql"

	"log"

	"tubr/models"
)

var selectSchedules = `
 SELECT 
	schedule_id, 
	label,
	game_id,
	broadcaster,
	platform,
	language,
	broadcaster_id,
	start_date,
	frequency_days,
	clip_time_max_secondsr,
	repeat_broadcaster,
	target_duration_seconds,
	webhook_url 
FROM SCHEDULES
`

func (db *DB) InsertSchedule(ctx context.Context, s *models.Schedule) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO schedules (
			schedule_id, 
			label,
			game_id,
			broadcaster,
			platform,
			language,
			broadcaster_id,
			start_date,
			frequency_days,
			clip_time_max_seconds,
			repeat_broadcaster,
			target_duration_seconds,
			webhook_url 
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (schedule_id) DO UPDATE SET schedule_id = $1
		`, s.ID, s.Label, s.GameID, s.Broadcaster, s.Platform, s.Language, s.BroadcasterID, s.StartDate, s.FrequencyDays, s.ClipTimeMaxSeconds, s.RepeatBroadcaster, s.TargetDurationSeconds, s.WebhookURL)
	return err
}

// ClipQueryArgs is a struct for passing arguments to db queries on the clips table
type ScheduleQueryArgs struct {
	ScheduleID int
}

// DefaultClipQueryArgs returns a default struct of args
func DefaultScheduleQueryArgs() *ScheduleQueryArgs {
	return &ScheduleQueryArgs{
		ScheduleID: 1,
	}
}

// SelectSchedule .
func (db *DB) GetSchedules(ctx context.Context, args *ScheduleQueryArgs) ([]models.Schedule, error) {
	query := `
	SELECT
		schedule_id,
		label,
		game_id,
		broadcaster,
		platform,
		language,
		broadcaster_id,
		start_date,
		frequency_days,
		clip_time_max_seconds,
		repeat_broadcaster,
		target_duration_seconds,
		webhook_url	
	FROM schedules`
	query += " ORDER BY schedule_id"
	r, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Println("Error fetching schedules:", err)
		return nil, err
	}

	return scanSchedules(r)
}

func scanSchedules(r *sql.Rows) ([]models.Schedule, error) {
	schedules := []models.Schedule{}
	for r.Next() {
		c := models.Schedule{}
		if err := r.Scan(
			&c.ID,
			&c.Label,
			&c.GameID,
			&c.Broadcaster,
			&c.Platform,
			&c.Language,
			&c.BroadcasterID,
			&c.StartDate,
			&c.FrequencyDays,
			&c.ClipTimeMaxSeconds,
			&c.RepeatBroadcaster,
			&c.TargetDurationSeconds,
			&c.WebhookURL,
		); err != nil {
			return schedules, err
		}
		schedules = append(schedules, c)
	}
	return schedules, nil
}
