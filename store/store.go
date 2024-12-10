package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// postgres dialect
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
)

// DB is the base postgres database type
type DB struct {
	*sql.DB
}

type PostgresConfig struct {
	Host     string `json:"host"`
	Password string `json:"password"`
	User     string `json:"user"`
	DbName   string `json:"db_name"`
}

// Open opens a database connection and initializes the database
func Open(ctx context.Context, config PostgresConfig) (*DB, error) {
	connStr := fmt.Sprintf(`
		host=%s
		dbname=%s
		user=%s
		password=%s
		sslmode=disable`, config.Host, config.DbName, config.User, config.Password)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	db := &DB{conn}
	return db.init(ctx)
}

// init sets up tables in the database
func (db *DB) init(ctx context.Context) (*DB, error) {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS clips (
			id text PRIMARY KEY,
			video_id text,
			game_id text,
			lang text,
			title text,
			view_count numeric,
			duration numeric,
			clip_date timestamp with time zone,
			broadcaster text,
			clipper text
		);

		CREATE TABLE IF NOT EXISTS users (
			id text PRIMARY KEY,
			name text
		);

		CREATE TABLE IF NOT EXISTS games (
			id text PRIMARY KEY,
			title text
		);

		CREATE TABLE IF NOT EXISTS schedules (
			schedule_id serial PRIMARY KEY, 
			label text,
			game_id text,
			broadcaster text,
			platform text,
			language text,
			broadcaster_id text,
			start_date timestamp,
			frequency_days integer,
			clip_time_max_seconds integer,
			repeat_broadcaster boolean,
			target_duration_seconds integer,
			webhook_url text
		);

		CREATE TABLE IF NOT EXISTS videos (
			schedule_id numeric, 
			start_date timestamp without time zone,
			end_date timestamp without time zone,
			release_date timestamp,
			cache_destination text,
			uploaded boolean
		);
	`)

	if err != nil {
		pgErr, ok := err.(*pq.Error)
		if !ok {
			return db, err
		}
		return db, pgErr
	}

	return db, nil
}

// Reset drops relavent tables in the database
func (db *DB) Reset(ctx context.Context) (*DB, error) {
	_, err := db.ExecContext(ctx, `
		DROP TABLE IF EXISTS clips, users, games, schedules, videos;
	`)
	return db, err
}

func (db *DB) UpdateCacheDestination(ctx context.Context, cacheDestination string, startDate time.Time, scheduleID int) error {
	query := `
		UPDATE videos 
		SET cache_destination = $1
		WHERE start_date = $2 AND schedule_id = $3;`

	_, err := db.ExecContext(ctx, query, cacheDestination, startDate, scheduleID)
	return err
}

// IsPrimaryKeyExistsErr returns true if the error is the result
// of a duplicate primary key already existing in the database
func IsPrimaryKeyExistsErr(err error) bool {
	if pgErr, ok := err.(*pq.Error); ok {
		if pgErr.Code == "23505" {
			return true
		}
	}
	return false
}
