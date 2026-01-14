package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"

	"smlgoapi/config"
	"smlgoapi/models"

	_ "github.com/lib/pq"
)

type PostgreSQLService struct {
	db                  *sql.DB
	config              *config.Config
	resultTableEnsured  map[string]bool
	resultTableEnsuredM sync.Mutex
}

func NewPostgreSQLService(config *config.Config) (*PostgreSQLService, error) {
	// Use System DSN for the main service connection (for internal admin tasks like InitSystemTables, UpsertUser)
	db, err := sql.Open("postgres", config.GetPostgreSQLSystemDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	service := &PostgreSQLService{
		db:                 db,
		config:             config,
		resultTableEnsured: make(map[string]bool),
	}

	if err := service.ensureResultTable(service.db, config.PostgreSQL.Database); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensure result table: %w", err)
	}

	if err := service.InitSystemTables(); err != nil {
		log.Printf("⚠️  Warning: Failed to initialize system tables: %v", err)
		// We don't necessarily want to fail here if the db is read-only or similar,
		// but for this migration it's important.
	}

	return service, nil
}

func (s *PostgreSQLService) Close() error {
	return s.db.Close()
}

// GetDefaultDatabaseName returns the database name configured via POSTGRESQL_DATABASE environment variable
func (s *PostgreSQLService) GetDefaultDatabaseName() string {
	return s.config.PostgreSQL.Database
}

// GetDefaultDatabaseConnection returns a connection to the default database configured via POSTGRESQL_DATABASE
func (s *PostgreSQLService) GetDefaultDatabaseConnection() (*sql.DB, error) {
	return s.GetDatabaseConnection(s.config.PostgreSQL.Database)
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

	if err := s.ensureResultTable(db, databaseName); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensure result table: %w", err)
	}

	return db, nil
}

// GetSystemDatabaseConnection creates a new connection to a specific database using System (Admin) credentials
func (s *PostgreSQLService) GetSystemDatabaseConnection(databaseName string) (*sql.DB, error) {
	// Create DSN for specific database using System credentials
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		s.config.PostgreSQLSystem.Host,
		s.config.PostgreSQLSystem.Port,
		s.config.PostgreSQLSystem.User,
		s.config.PostgreSQLSystem.Password,
		databaseName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL system connection to database '%s': %w", databaseName, err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL system database '%s': %w", databaseName, err)
	}

	// Ensure result table exists (using system connection is safe)
	if err := s.ensureResultTable(db, databaseName); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensure result table: %w", err)
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

func (s *PostgreSQLService) ensureResultTable(db *sql.DB, databaseName string) error {
	if db == nil {
		return fmt.Errorf("nil database connection")
	}

	s.resultTableEnsuredM.Lock()
	if s.resultTableEnsured == nil {
		s.resultTableEnsured = make(map[string]bool)
	}
	if s.resultTableEnsured[databaseName] {
		s.resultTableEnsuredM.Unlock()
		return nil
	}
	s.resultTableEnsuredM.Unlock()

	if err := TableResultCreate(db); err != nil {
		return err
	}

	s.resultTableEnsuredM.Lock()
	s.resultTableEnsured[databaseName] = true
	s.resultTableEnsuredM.Unlock()

	return nil
}
