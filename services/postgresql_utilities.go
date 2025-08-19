package services

import (
	"database/sql"
	"fmt"
)

// CheckDatabaseExists checks if a database exists in PostgreSQL
func (s *PostgreSQLService) CheckDatabaseExists(databaseName string) (bool, error) {
	query := `SELECT 1 FROM pg_database WHERE datname = $1`
	var exists int
	err := s.db.QueryRow(query, databaseName).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check database existence: %w", err)
	}

	return true, nil
}

// CheckTableExists checks if a table exists in a PostgreSQL database
func (s *PostgreSQLService) CheckTableExists(databaseName, tableName string) (bool, error) {
	// Connect to the specific database
	db, err := s.GetDatabaseConnection(databaseName)
	if err != nil {
		return false, fmt.Errorf("failed to connect to database '%s': %w", databaseName, err)
	}
	defer db.Close()

	query := `SELECT 1 FROM information_schema.tables WHERE table_catalog = $1 AND table_name = $2`
	var exists int
	err = db.QueryRow(query, databaseName, tableName).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}

	return true, nil
}

// GetTableList returns a list of all tables in a PostgreSQL database
func (s *PostgreSQLService) GetTableList(databaseName string) ([]string, error) {
	// Connect to the specific database
	db, err := s.GetDatabaseConnection(databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database '%s': %w", databaseName, err)
	}
	defer db.Close()

	query := `SELECT table_name FROM information_schema.tables WHERE table_catalog = $1 AND table_schema = 'public' ORDER BY table_name`
	rows, err := db.Query(query, databaseName)
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
