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
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		tables = append(tables, table)
	}

	return tables, rows.Err()
}

// ExecuteCommand executes a SQL command (INSERT, UPDATE, DELETE, CREATE, etc.)
func (s *PostgreSQLService) ExecuteCommand(ctx context.Context, query string) (interface{}, error) {
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
func (s *PostgreSQLService) ExecuteSelect(ctx context.Context, query string) ([]interface{}, error) {
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

// SearchProducts performs a text-based search on the ic_inventory table in PostgreSQL
func (s *PostgreSQLService) SearchProducts(ctx context.Context, query string, limit, offset int) ([]map[string]interface{}, int, error) {
	// First check if the ic_inventory table exists
	checkTableQuery := `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = 'ic_inventory'`

	var tableExists int
	err := s.db.QueryRowContext(ctx, checkTableQuery).Scan(&tableExists)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check table existence: %w", err)
	}

	// If ic_inventory table doesn't exist, return helpful message
	if tableExists == 0 {
		return []map[string]interface{}{
			{
				"id":               "error",
				"name":             "ไม่พบตาราง ic_inventory ในฐานข้อมูล",
				"code":             "NO_IC_INVENTORY",
				"description":      "กรุณาสร้างตาราง ic_inventory หรือติดต่อผู้ดูแลระบบ",
				"price":            0.0,
				"balance_qty":      0.0,
				"unit":             "N/A",
				"supplier_code":    "N/A",
				"img_url":          "",
				"search_priority":  1,
				"similarity_score": 0.0,
			},
		}, 1, nil
	}

	// Get count of matching records
	countQuery := `
		SELECT COUNT(*) as total_count
		FROM ic_inventory 
		WHERE LOWER(CAST(name AS TEXT)) LIKE LOWER($1) 
		   OR LOWER(CAST(code AS TEXT)) LIKE LOWER($1)`
	countRows, err := s.db.QueryContext(ctx, countQuery, "%"+query+"%")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute count query: %w", err)
	}
	defer countRows.Close()

	var totalCount int
	if countRows.Next() {
		if err := countRows.Scan(&totalCount); err != nil {
			return nil, 0, fmt.Errorf("failed to scan count result: %w", err)
		}
	} // Now get the actual search results with pagination from ic_inventory
	searchQuery := `
		SELECT COALESCE(CAST(code AS TEXT), 'N/A') as code, 
		       COALESCE(CAST(name AS TEXT), 'N/A') as name,
		       COALESCE(CAST(unit_standard_code AS TEXT), 'N/A') as unit_standard_code,
		       COALESCE(item_type, 0) as item_type,
		       COALESCE(row_order_ref, 0) as row_order_ref,
		       CASE 
		           WHEN LOWER(CAST(code AS TEXT)) LIKE LOWER($1) THEN 3
		           WHEN LOWER(CAST(name AS TEXT)) LIKE LOWER($1) THEN 2
		           ELSE 1
		       END as search_priority
		FROM ic_inventory 
		WHERE LOWER(CAST(name AS TEXT)) LIKE LOWER($1) 
		   OR LOWER(CAST(code AS TEXT)) LIKE LOWER($1)
		ORDER BY search_priority DESC, name ASC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.QueryContext(ctx, searchQuery, "%"+query+"%", limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}

	for rows.Next() {
		var code, name, unitStandardCode string
		var itemType, rowOrderRef, searchPriority int

		err := rows.Scan(&code, &name, &unitStandardCode, &itemType, &rowOrderRef, &searchPriority)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan search result: %w", err)
		}

		result := map[string]interface{}{
			"id":                 code, // Use code as id since there's no separate id field
			"code":               code,
			"name":               name,
			"description":        "",  // Not available in ic_inventory
			"price":              0.0, // Not available in ic_inventory
			"balance_qty":        0.0, // Not available in ic_inventory
			"unit":               unitStandardCode,
			"supplier_code":      "N/A", // Not available in ic_inventory
			"img_url":            "",    // Not available in ic_inventory
			"search_priority":    searchPriority,
			"similarity_score":   float64(searchPriority), // Use search priority as similarity score
			"item_type":          itemType,
			"row_order_ref":      rowOrderRef,
			"unit_standard_code": unitStandardCode,
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, totalCount, nil
}
