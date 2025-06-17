package services

import (
	"context"
	"database/sql"
	"fmt"

	"smlgoapi/config"
	"smlgoapi/models"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouseService struct {
	db     *sql.DB
	config *config.Config
}

func NewClickHouseService(config *config.Config) (*ClickHouseService, error) {
	db, err := sql.Open("clickhouse", config.GetClickHouseDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open ClickHouse connection: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return &ClickHouseService{
		db:     db,
		config: config,
	}, nil
}

func (s *ClickHouseService) Close() error {
	return s.db.Close()
}

func (s *ClickHouseService) GetVersion(ctx context.Context) (string, error) {
	var version string
	err := s.db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	return version, err
}

func (s *ClickHouseService) GetTables(ctx context.Context) ([]models.Table, error) {
	rows, err := s.db.QueryContext(ctx, "SHOW TABLES")
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []models.Table
	for rows.Next() {
		var table models.Table
		if err := rows.Scan(&table.Name); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		tables = append(tables, table)
	}

	return tables, rows.Err()
}

// ExecuteCommand executes a SQL command (INSERT, UPDATE, DELETE, CREATE, etc.)
func (s *ClickHouseService) ExecuteCommand(ctx context.Context, query string) (interface{}, error) {
	// Execute the command
	result, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	// Get rows affected if possible
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// Some commands might not return rows affected, return basic success
		return map[string]interface{}{
			"status": "success",
			"query":  query,
		}, nil
	}

	return map[string]interface{}{
		"status":        "success",
		"rows_affected": rowsAffected,
		"query":         query,
	}, nil
}

// ExecuteSelect executes a SELECT query and returns the result data
func (s *ClickHouseService) ExecuteSelect(ctx context.Context, query string) ([]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute select query: %w", err)
	}
	defer rows.Close()

	// Get column information
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for this row
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Convert []uint8 to string if needed
			if b, ok := val.([]uint8); ok {
				val = string(b)
			}

			rowMap[col] = val
		}

		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, nil
}
