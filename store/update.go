package store

import (
	"context"
	"tubr/models"
)

// UpdateViewCount updates a clip's amount if views in the db
func (db *DB) UpdateViewCount(ctx context.Context, c *models.Clip) error {
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE clips SET view_count = $1
		WHERE id = $2
	`, c.ViewCount, c.ID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
