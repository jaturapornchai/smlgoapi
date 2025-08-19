package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"smlgoapi/config"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouseService struct {
	config *config.Config
}

func NewClickHouseService(cfg *config.Config) (*ClickHouseService, error) {
	return &ClickHouseService{
		config: cfg,
	}, nil
}

// GetDatabaseConnection returns a connection to a specific ClickHouse database
func (ch *ClickHouseService) GetDatabaseConnection(databaseName string) (*sql.DB, error) {
	// Open connection using ClickHouse driver
	db := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", ch.config.ClickHouse.Host, ch.config.ClickHouse.Port)},
		Auth: clickhouse.Auth{
			Database: databaseName,
			Username: ch.config.ClickHouse.User,
			Password: ch.config.ClickHouse.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 5 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse database '%s': %w", databaseName, err)
	}

	log.Printf("🎯 Connected to ClickHouse database: %s", databaseName)
	return db, nil
}

// CheckDatabaseExists checks if a database exists in ClickHouse
func (ch *ClickHouseService) CheckDatabaseExists(databaseName string) (bool, error) {
	// Connect to system database to check if target database exists
	db, err := ch.GetDatabaseConnection("system")
	if err != nil {
		return false, fmt.Errorf("failed to connect to ClickHouse system database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var count int
	query := "SELECT count() FROM system.databases WHERE name = ?"
	err = db.QueryRowContext(ctx, query, databaseName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check database existence: %w", err)
	}

	return count > 0, nil
}

// CheckTableExists checks if a table exists in a ClickHouse database
func (ch *ClickHouseService) CheckTableExists(databaseName, tableName string) (bool, error) {
	db, err := ch.GetDatabaseConnection(databaseName)
	if err != nil {
		return false, fmt.Errorf("failed to connect to ClickHouse database '%s': %w", databaseName, err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var count int
	query := "SELECT count() FROM system.tables WHERE database = ? AND name = ?"
	err = db.QueryRowContext(ctx, query, databaseName, tableName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}

	return count > 0, nil
}

// GetDatabaseList returns a list of all databases in ClickHouse
func (ch *ClickHouseService) GetDatabaseList() ([]string, error) {
	db, err := ch.GetDatabaseConnection("system")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse system database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := "SELECT name FROM system.databases ORDER BY name"
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get database list: %w", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, fmt.Errorf("failed to scan database name: %w", err)
		}
		databases = append(databases, dbName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating database rows: %w", err)
	}

	return databases, nil
}

// GetTableList returns a list of all tables in a ClickHouse database
func (ch *ClickHouseService) GetTableList(databaseName string) ([]string, error) {
	db, err := ch.GetDatabaseConnection(databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse database '%s': %w", databaseName, err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := "SELECT name FROM system.tables WHERE database = ? ORDER BY name"
	rows, err := db.QueryContext(ctx, query, databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get table list: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating table rows: %w", err)
	}

	return tables, nil
}
