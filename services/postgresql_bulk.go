package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

// BulkInsertWithCopy inserts multiple rows into a table using PostgreSQL COPY protocol
func BulkInsertWithCopy(ctx context.Context, db *sql.DB, tableName string, columns []string, rows [][]any) error {
	if len(rows) == 0 {
		return nil
	}

	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer txn.Rollback()

	stmt, err := txn.Prepare(pq.CopyIn(tableName, columns...))
	if err != nil {
		return fmt.Errorf("failed to prepare COPY statement: %w", err)
	}
	defer stmt.Close()

	for _, row := range rows {
		if _, err := stmt.Exec(row...); err != nil {
			return fmt.Errorf("failed to exec COPY: %w", err)
		}
	}

	if _, err := stmt.Exec(); err != nil {
		return fmt.Errorf("failed to flush COPY: %w", err)
	}

	if err := stmt.Close(); err != nil {
		return fmt.Errorf("failed to close COPY statement: %w", err)
	}

	if err := txn.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
