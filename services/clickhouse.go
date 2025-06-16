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
