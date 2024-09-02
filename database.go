package main

import (
	"database/sql"
	"fmt"
)

func setUploaded(db *sql.DB, id int) error {
	query := `UPDATE files SET uploaded = 1 WHERE id = ?`
	_, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to update uploaded column: %v", err)
	}
	return nil
}
