package services

import (
	"context"
	"database/sql"
	"fmt"
	"os"
)

// CheckAndCreateDatabase checks if a specific database exists and creates it if not
func (s *PostgreSQLService) CheckAndCreateDatabase(ctx context.Context, databaseName string) (map[string]interface{}, error) {
	// First, check if database exists
	checkQuery := `SELECT 1 FROM pg_database WHERE datname = $1`
	var exists int
	err := s.db.QueryRowContext(ctx, checkQuery, databaseName).Scan(&exists)

	if err == sql.ErrNoRows {
		// Database doesn't exist, create it
		createQuery := fmt.Sprintf(`CREATE DATABASE "%s"`, databaseName)
		_, err := s.db.ExecContext(ctx, createQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to create database '%s': %w", databaseName, err)
		}

		return map[string]interface{}{
			"status":   "created",
			"database": databaseName,
			"message":  fmt.Sprintf("Database '%s' created successfully", databaseName),
		}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to check database existence: %w", err)
	}

	// Database exists
	return map[string]interface{}{
		"status":   "exists",
		"database": databaseName,
		"message":  fmt.Sprintf("Database '%s' already exists", databaseName),
	}, nil
}

// CheckAndCreateTable checks if a table exists and creates it from SQL script if not
func (s *PostgreSQLService) CheckAndCreateTable(ctx context.Context, databaseName, tableName string) (map[string]interface{}, error) {
	// First, check if table exists
	checkQuery := `SELECT 1 FROM information_schema.tables WHERE table_catalog = $1 AND table_name = $2`
	var exists int
	err := s.db.QueryRowContext(ctx, checkQuery, databaseName, tableName).Scan(&exists)

	if err == sql.ErrNoRows {
		// Table doesn't exist, look for SQL script and create it
		scriptPath := fmt.Sprintf("sql/postgresql/%s.sql", tableName)

		// Check if SQL file exists
		sqlContent, err := os.ReadFile(scriptPath)
		if err != nil {
			return map[string]interface{}{
				"status":      "script_not_found",
				"table_name":  tableName,
				"database":    databaseName,
				"script_path": scriptPath,
				"message":     fmt.Sprintf("SQL script not found: %s", scriptPath),
			}, nil
		}

		// Execute the SQL script
		_, err = s.db.ExecContext(ctx, string(sqlContent))
		if err != nil {
			return nil, fmt.Errorf("failed to create table '%s' from script '%s': %w", tableName, scriptPath, err)
		}

		return map[string]interface{}{
			"status":      "created",
			"table_name":  tableName,
			"database":    databaseName,
			"script_path": scriptPath,
			"message":     fmt.Sprintf("Table '%s' created successfully from script '%s'", tableName, scriptPath),
		}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to check table existence: %w", err)
	}

	// Table exists
	return map[string]interface{}{
		"status":     "exists",
		"table_name": tableName,
		"database":   databaseName,
		"message":    fmt.Sprintf("Table '%s' already exists in database '%s'", tableName, databaseName),
	}, nil
}

// CheckAndCreateAllTables scans all SQL files in sql/postgresql/ folder and creates missing tables
func (s *PostgreSQLService) CheckAndCreateAllTables(ctx context.Context, databaseName string) (map[string]interface{}, error) {
	// Get connection to specific database (Use System Connection for creating tables)
	db, err := s.GetSystemDatabaseConnection(databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database '%s': %w", databaseName, err)
	}
	defer db.Close()

	// Read SQL folder to get all .sql files
	files, err := os.ReadDir("sql/postgresql")
	if err != nil {
		return nil, fmt.Errorf("failed to read sql/postgresql directory: %w", err)
	}

	var allTables []string
	var createdTables []string
	var existingTables []string
	var totalCreated int

	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 4 && file.Name()[len(file.Name())-4:] == ".sql" {
			// Extract table name from filename (remove .sql extension)
			tableName := file.Name()[:len(file.Name())-4]
			allTables = append(allTables, tableName)

			// Check if table exists
			checkQuery := `SELECT 1 FROM information_schema.tables WHERE table_catalog = $1 AND table_name = $2`
			var exists int
			err := db.QueryRowContext(ctx, checkQuery, databaseName, tableName).Scan(&exists)

			if err == sql.ErrNoRows {
				// Table doesn't exist, create it from SQL script
				scriptPath := fmt.Sprintf("sql/postgresql/%s", file.Name())
				sqlContent, err := os.ReadFile(scriptPath)
				if err != nil {
					return nil, fmt.Errorf("failed to read SQL script '%s': %w", scriptPath, err)
				}

				_, err = db.ExecContext(ctx, string(sqlContent))
				if err != nil {
					return nil, fmt.Errorf("failed to execute SQL script '%s': %w", scriptPath, err)
				}

				createdTables = append(createdTables, tableName)
				totalCreated++
			} else if err != nil {
				return nil, fmt.Errorf("failed to check table '%s' existence: %w", tableName, err)
			} else {
				// Table exists
				existingTables = append(existingTables, tableName)
			}
		}
	}

	var status string
	var message string

	if totalCreated == 0 {
		status = "all_exist"
		message = fmt.Sprintf("All %d tables already exist in database '%s'", len(allTables), databaseName)
	} else if len(existingTables) == 0 {
		status = "all_created"
		message = fmt.Sprintf("Created %d tables in database '%s'", totalCreated, databaseName)
	} else {
		status = "partial_created"
		message = fmt.Sprintf("Created %d new tables, %d already existed in database '%s'", totalCreated, len(existingTables), databaseName)
	}

	return map[string]interface{}{
		"status":   status,
		"database": databaseName,
		"message":  message,
		"tables":   allTables,
		"created":  createdTables,
		"existing": existingTables,
	}, nil
}
