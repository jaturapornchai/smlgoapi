package services

import (
	"context"
	"database/sql"
	"fmt"

	"smlgoapi/config"
	"smlgoapi/models"

	_ "github.com/lib/pq"
)

type PostgreSQLService struct {
	db     *sql.DB
	config *config.Config
}

func NewPostgreSQLService(config *config.Config) (*PostgreSQLService, error) {
	db, err := sql.Open("postgres", config.GetPostgreSQLDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return &PostgreSQLService{
		db:     db,
		config: config,
	}, nil
}

func (s *PostgreSQLService) Close() error {
	return s.db.Close()
}

// GetDatabaseConnection creates a new connection to a specific database
func (s *PostgreSQLService) GetDatabaseConnection(databaseName string) (*sql.DB, error) {
	// Create DSN for specific database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		s.config.PostgreSQL.Host,
		s.config.PostgreSQL.Port,
		s.config.PostgreSQL.User,
		s.config.PostgreSQL.Password,
		databaseName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection to database '%s': %w", databaseName, err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL database '%s': %w", databaseName, err)
	}

	return db, nil
}

func (s *PostgreSQLService) GetVersion(ctx context.Context) (string, error) {
	var version string
	err := s.db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	return version, err
}

func (s *PostgreSQLService) GetTables(ctx context.Context) ([]models.Table, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []models.Table
	for rows.Next() {
		var table models.Table
		if err := rows.Scan(&table.Name); err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}
		tables = append(tables, table)
	}

	return tables, rows.Err()
}
