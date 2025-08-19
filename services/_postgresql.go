package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

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

// PriceInfo holds price information from ic_inventory_price_formula
type PriceInfo struct {
	ICCode string
	Price0 float64 // Parsed from string
	Price1 float64
	Price2 float64
	Price3 float64
	Price4 float64
}

// BalanceInfo holds balance information from ic_balance
type BalanceInfo struct {
	ICCode   string
	TotalQty float64 // Sum of balance_qty across all wh_code
}

// LoadPriceFormula loads all price data from ic_inventory_price_formula into memory
func (s *PostgreSQLService) LoadPriceFormula(ctx context.Context) (map[string]*PriceInfo, error) {
	// Check if the price formula table exists
	checkTableQuery := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'ic_inventory_price_formula'`

	var tableExists int
	err := s.db.QueryRowContext(ctx, checkTableQuery).Scan(&tableExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check price formula table existence: %w", err)
	}

	if tableExists == 0 {
		log.Printf("⚠️ Table 'ic_inventory_price_formula' not found - using default prices")
		return make(map[string]*PriceInfo), nil
	}
	// Load all price data
	query := `
		SELECT COALESCE(CAST(ic_code AS TEXT), '') as ic_code,
		       COALESCE(CAST(price_0 AS TEXT), '0') as price_0,
		       COALESCE(CAST(price_1 AS TEXT), '0') as price_1,
		       COALESCE(CAST(price_2 AS TEXT), '0') as price_2,
		       COALESCE(CAST(price_3 AS TEXT), '0') as price_3,
		       COALESCE(CAST(price_4 AS TEXT), '0') as price_4
		FROM ic_inventory_price_formula
		WHERE ic_code IS NOT NULL AND ic_code != ''`

	log.Printf("🏷️ Loading price formula data...")

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to load price formula: %w", err)
	}
	defer rows.Close()

	priceMap := make(map[string]*PriceInfo)
	for rows.Next() {
		var icCode, price0Str, price1Str, price2Str, price3Str, price4Str string

		err := rows.Scan(&icCode, &price0Str, &price1Str, &price2Str, &price3Str, &price4Str)
		if err != nil {
			log.Printf("❌ Failed to scan price row: %v", err)
			continue
		}

		// Parse all prices from string to float64
		price0, err := strconv.ParseFloat(strings.TrimSpace(price0Str), 64)
		if err != nil {
			log.Printf("⚠️ Failed to parse price_0 '%s' for ic_code '%s': %v", price0Str, icCode, err)
			price0 = 0.0
		}

		price1, err := strconv.ParseFloat(strings.TrimSpace(price1Str), 64)
		if err != nil {
			price1 = 0.0
		}

		price2, err := strconv.ParseFloat(strings.TrimSpace(price2Str), 64)
		if err != nil {
			price2 = 0.0
		}

		price3, err := strconv.ParseFloat(strings.TrimSpace(price3Str), 64)
		if err != nil {
			price3 = 0.0
		}

		price4, err := strconv.ParseFloat(strings.TrimSpace(price4Str), 64)
		if err != nil {
			price4 = 0.0
		}

		priceMap[icCode] = &PriceInfo{
			ICCode: icCode,
			Price0: price0,
			Price1: price1,
			Price2: price2,
			Price3: price3,
			Price4: price4,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	log.Printf("✅ Loaded %d price records", len(priceMap))
	return priceMap, nil
}

// LoadPriceFormulaFiltered loads price data for specific ic_codes only
func (s *PostgreSQLService) LoadPriceFormulaFiltered(ctx context.Context, icCodes []string) (map[string]*PriceInfo, error) {
	if len(icCodes) == 0 {
		return make(map[string]*PriceInfo), nil
	}

	// Check if the price formula table exists
	checkTableQuery := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'ic_inventory_price_formula'`

	var tableExists int
	err := s.db.QueryRowContext(ctx, checkTableQuery).Scan(&tableExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check price formula table existence: %w", err)
	}

	if tableExists == 0 {
		log.Printf("⚠️ Table 'ic_inventory_price_formula' not found - using default prices")
		return make(map[string]*PriceInfo), nil
	}

	// Build IN clause for filtering
	placeholders := make([]string, len(icCodes))
	params := make([]interface{}, len(icCodes))
	for i, code := range icCodes {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		params[i] = code
	}

	// Load filtered price data
	query := fmt.Sprintf(`
		SELECT COALESCE(CAST(ic_code AS TEXT), '') as ic_code,
		       COALESCE(CAST(price_0 AS TEXT), '0') as price_0,
		       COALESCE(CAST(price_1 AS TEXT), '0') as price_1,
		       COALESCE(CAST(price_2 AS TEXT), '0') as price_2,
		       COALESCE(CAST(price_3 AS TEXT), '0') as price_3,
		       COALESCE(CAST(price_4 AS TEXT), '0') as price_4
		FROM ic_inventory_price_formula
		WHERE ic_code IN (%s)`, strings.Join(placeholders, ","))

	log.Printf("🏷️ Loading price formula data for %d specific items...", len(icCodes))

	rows, err := s.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to load filtered price formula: %w", err)
	}
	defer rows.Close()

	priceMap := make(map[string]*PriceInfo)

	for rows.Next() {
		var icCode, price0Str, price1Str, price2Str, price3Str, price4Str string

		err := rows.Scan(&icCode, &price0Str, &price1Str, &price2Str, &price3Str, &price4Str)
		if err != nil {
			log.Printf("❌ Failed to scan price row: %v", err)
			continue
		}

		// Parse all prices from string to float64
		price0, err := strconv.ParseFloat(strings.TrimSpace(price0Str), 64)
		if err != nil {
			log.Printf("⚠️ Failed to parse price_0 '%s' for ic_code '%s': %v", price0Str, icCode, err)
			price0 = 0.0
		}

		price1, err := strconv.ParseFloat(strings.TrimSpace(price1Str), 64)
		if err != nil {
			price1 = 0.0
		}

		price2, err := strconv.ParseFloat(strings.TrimSpace(price2Str), 64)
		if err != nil {
			price2 = 0.0
		}

		price3, err := strconv.ParseFloat(strings.TrimSpace(price3Str), 64)
		if err != nil {
			price3 = 0.0
		}

		price4, err := strconv.ParseFloat(strings.TrimSpace(price4Str), 64)
		if err != nil {
			price4 = 0.0
		}

		priceMap[icCode] = &PriceInfo{
			ICCode: icCode,
			Price0: price0,
			Price1: price1,
			Price2: price2,
			Price3: price3,
			Price4: price4,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	log.Printf("✅ Loaded %d filtered price records", len(priceMap))
	return priceMap, nil
}

// LoadBalanceData loads all balance data from ic_balance into memory, grouped by ic_code
func (s *PostgreSQLService) LoadBalanceData(ctx context.Context) (map[string]*BalanceInfo, error) {
	// Check if the balance table exists
	checkTableQuery := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'ic_balance'`

	var tableExists int
	err := s.db.QueryRowContext(ctx, checkTableQuery).Scan(&tableExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check balance table existence: %w", err)
	}

	if tableExists == 0 {
		log.Printf("⚠️ Table 'ic_balance' not found - using default balance")
		return make(map[string]*BalanceInfo), nil
	}

	// Load balance data grouped by ic_code (sum balance_qty by wh_code)
	query := `
		SELECT COALESCE(CAST(ic_code AS TEXT), '') as ic_code,
		       COALESCE(SUM(balance_qty), 0) as total_qty
		FROM ic_balance
		WHERE ic_code IS NOT NULL AND ic_code != ''
		GROUP BY ic_code`

	log.Printf("📦 Loading balance data...")

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to load balance data: %w", err)
	}
	defer rows.Close()

	balanceMap := make(map[string]*BalanceInfo)

	for rows.Next() {
		var icCode string
		var totalQty float64

		err := rows.Scan(&icCode, &totalQty)
		if err != nil {
			log.Printf("❌ Failed to scan balance row: %v", err)
			continue
		}

		balanceMap[icCode] = &BalanceInfo{
			ICCode:   icCode,
			TotalQty: totalQty,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("balance rows iteration error: %w", err)
	}

	log.Printf("✅ Loaded %d balance records", len(balanceMap))
	return balanceMap, nil
}

// LoadBalanceDataFiltered loads balance data for specific ic_codes only
func (s *PostgreSQLService) LoadBalanceDataFiltered(ctx context.Context, icCodes []string) (map[string]*BalanceInfo, error) {
	if len(icCodes) == 0 {
		return make(map[string]*BalanceInfo), nil
	}

	// Check if the balance table exists
	checkTableQuery := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'ic_balance'`

	var tableExists int
	err := s.db.QueryRowContext(ctx, checkTableQuery).Scan(&tableExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check balance table existence: %w", err)
	}

	if tableExists == 0 {
		log.Printf("⚠️ Table 'ic_balance' not found - using default balance")
		return make(map[string]*BalanceInfo), nil
	}

	// Build IN clause for filtering
	placeholders := make([]string, len(icCodes))
	params := make([]interface{}, len(icCodes))
	for i, code := range icCodes {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		params[i] = code
	}

	// Load filtered balance data grouped by ic_code
	query := fmt.Sprintf(`
		SELECT COALESCE(CAST(ic_code AS TEXT), '') as ic_code,
		       COALESCE(SUM(balance_qty), 0) as total_qty
		FROM ic_balance
		WHERE ic_code IN (%s)
		GROUP BY ic_code`, strings.Join(placeholders, ","))

	log.Printf("📦 Loading balance data for %d specific items...", len(icCodes))

	rows, err := s.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to load filtered balance data: %w", err)
	}
	defer rows.Close()

	balanceMap := make(map[string]*BalanceInfo)

	for rows.Next() {
		var icCode string
		var totalQty float64

		err := rows.Scan(&icCode, &totalQty)
		if err != nil {
			log.Printf("❌ Failed to scan balance row: %v", err)
			continue
		}

		balanceMap[icCode] = &BalanceInfo{
			ICCode:   icCode,
			TotalQty: totalQty,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("balance rows iteration error: %w", err)
	}

	log.Printf("✅ Loaded %d filtered balance records", len(balanceMap))
	return balanceMap, nil
}

// SearchProducts performs a full text search on the ic_inventory table in PostgreSQL
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
	// If ic_inventory table doesn't exist, return error instead of mock data
	if tableExists == 0 {
		return nil, 0, fmt.Errorf("table 'ic_inventory' not found in database - please create the table or contact system administrator")
	}

	// Split query into words for OR search
	words := strings.Fields(strings.TrimSpace(query))
	if len(words) == 0 {
		words = []string{query} // If no spaces, use the whole query
	}
	// Build OR conditions for full text search - using ILIKE for better Unicode support
	// Search only in 'code' and 'name' fields as requested
	var orConditions []string
	for range words {
		orConditions = append(orConditions, "CAST(name AS TEXT) ILIKE ?")
		orConditions = append(orConditions, "CAST(code AS TEXT) ILIKE ?")
	}

	// Convert PostgreSQL placeholder format
	whereClause := strings.Join(orConditions, " OR ")
	paramIndex := 1
	for range orConditions {
		whereClause = strings.Replace(whereClause, "?", fmt.Sprintf("$%d", paramIndex), 1)
		paramIndex++
	}

	// Prepare parameters for count query
	var countParams []interface{}
	for _, word := range words {
		if strings.TrimSpace(word) != "" {
			countParams = append(countParams, "%"+word+"%") // name search
			countParams = append(countParams, "%"+word+"%") // code search
		}
	}

	// Get count of matching records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) as total_count
		FROM ic_inventory
		WHERE %s`, whereClause)

	countRows, err := s.db.QueryContext(ctx, countQuery, countParams...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute count query: %w", err)
	}
	defer countRows.Close()

	var totalCount int
	if countRows.Next() {
		if err := countRows.Scan(&totalCount); err != nil {
			return nil, 0, fmt.Errorf("failed to scan count result: %w", err)
		}
	}

	// Build search query with priority scoring
	searchQuery := fmt.Sprintf(`
		SELECT COALESCE(CAST(code AS TEXT), 'N/A') as code,
		       COALESCE(CAST(name AS TEXT), 'N/A') as name,
		       COALESCE(CAST(unit_standard_code AS TEXT), 'N/A') as unit_standard_code,
		       COALESCE(item_type, 0) as item_type,
		       COALESCE(row_order_ref, 0) as row_order_ref,
		       CASE
		           WHEN CAST(code AS TEXT) ILIKE $%d THEN 5
		           WHEN CAST(code AS TEXT) ILIKE $%d THEN 3
		           WHEN CAST(name AS TEXT) ILIKE $%d THEN 2
		           ELSE 1
		       END as search_priority
		FROM ic_inventory
		WHERE %s
		ORDER BY search_priority DESC, LENGTH(name) ASC, name ASC
		LIMIT $%d OFFSET $%d`,
		len(countParams)+1, len(countParams)+2, len(countParams)+3, whereClause, len(countParams)+4, len(countParams)+5)

	// Prepare parameters for search query
	searchParams := make([]interface{}, 0)
	searchParams = append(searchParams, countParams...) // word parameters
	searchParams = append(searchParams, query)          // exact match for code
	searchParams = append(searchParams, "%"+query+"%")  // like match for code
	searchParams = append(searchParams, "%"+query+"%")  // like match for name
	searchParams = append(searchParams, limit)          // limit
	searchParams = append(searchParams, offset)         // offset

	// Log the actual SQL query for debugging
	// log.Printf("🔍 SQL Query: %s", searchQuery)
	// log.Printf("🔍 Parameters: %v", searchParams)

	rows, err := s.db.QueryContext(ctx, searchQuery, searchParams...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	var icCodes []string // Collect ic_codes for filtered price/balance loading

	for rows.Next() {
		var code, name, unitStandardCode string
		var itemType, rowOrderRef, searchPriority int

		err := rows.Scan(&code, &name, &unitStandardCode, &itemType, &rowOrderRef, &searchPriority)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan search result: %w", err)
		}

		icCodes = append(icCodes, code) // Collect ic_code for later price/balance lookup

		// Default values for pricing and inventory fields
		var salePrice, discountPrice, discountPercent, finalPrice, soldQty, qtyAvailable float64 = 0.0, 0.0, 0.0, 0.0, 0.0, 0.0
		premiumWord := "N/A"
		multiPacking := 0
		multiPackingName := "N/A"
		barcodes := "N/A"

		result := map[string]interface{}{
			"id":                 code, // Use code as id since there's no separate id field
			"code":               code,
			"name":               name,
			"unit_standard_code": unitStandardCode,
			"item_type":          itemType,
			"row_order_ref":      rowOrderRef,
			"search_priority":    searchPriority,
			"similarity_score":   float64(searchPriority), // Use search priority as similarity score

			// Pricing and inventory fields (will be updated below)
			"sale_price":         salePrice,
			"premium_word":       premiumWord,
			"discount_price":     discountPrice,
			"discount_percent":   discountPercent,
			"final_price":        finalPrice,
			"sold_qty":           soldQty,
			"multi_packing":      multiPacking,
			"multi_packing_name": multiPackingName,
			"barcodes":           barcodes,
			"qty_available":      qtyAvailable,

			// Legacy fields for backward compatibility
			"description":   "",        // Not available in ic_inventory
			"price":         salePrice, // Map to sale_price for compatibility
			"balance_qty":   0.0,       // Not available in ic_inventory
			"unit":          unitStandardCode,
			"supplier_code": "N/A", // Not available in ic_inventory
			"img_url":       "",    // Not available in ic_inventory
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	// Now load price and balance data only for the found products
	log.Printf("🏷️ Loading price formula data for %d found items...", len(icCodes))
	priceMap, err := s.LoadPriceFormulaFiltered(ctx, icCodes)
	if err != nil {
		log.Printf("⚠️ Failed to load price formula: %v - using default prices", err)
		priceMap = make(map[string]*PriceInfo)
	}

	log.Printf("📦 Loading balance data for %d found items...", len(icCodes))
	balanceMap, err := s.LoadBalanceDataFiltered(ctx, icCodes)
	if err != nil {
		log.Printf("⚠️ Failed to load balance data: %v - using default balance", err)
		balanceMap = make(map[string]*BalanceInfo)
	}

	// Update results with price and balance data
	for i, result := range results {
		code := result["code"].(string)

		// Look up real price data
		if priceInfo, exists := priceMap[code]; exists {
			salePrice := priceInfo.Price0     // Use price_0 as sale_price
			finalPrice := priceInfo.Price0    // Use price_0 as final_price too
			discountPrice := priceInfo.Price1 // Use price_1 as discount_price if available

			results[i]["sale_price"] = salePrice
			results[i]["final_price"] = finalPrice
			results[i]["discount_price"] = discountPrice
			results[i]["price"] = salePrice // Update legacy field too

			log.Printf("💰 Found price for %s: sale_price=%.2f, final_price=%.2f, discount_price=%.2f",
				code, salePrice, finalPrice, discountPrice)
		} else {
			log.Printf("⚠️ No price found for ic_code: %s - using defaults", code)
		}

		// Look up real balance data
		if balanceInfo, exists := balanceMap[code]; exists {
			qtyAvailable := balanceInfo.TotalQty // Use sum of balance_qty as qty_available
			results[i]["qty_available"] = qtyAvailable
			log.Printf("📦 Found balance for %s: qty_available=%.2f", code, qtyAvailable)
		} else {
			log.Printf("⚠️ No balance found for ic_code: %s - using default (0.0)", code)
		}
	}

	log.Printf("✅ Search completed: found %d results, total count: %d", len(results), totalCount)
	return results, totalCount, nil
}

// SearchProductsByBarcodes performs search on the ic_inventory table using specific barcodes
func (s *PostgreSQLService) SearchProductsByBarcodes(ctx context.Context, barcodes []string, limit, offset int) ([]map[string]interface{}, int, error) {
	// For now, treat barcodes as search terms
	query := strings.Join(barcodes, " ")
	return s.SearchProducts(ctx, query, limit, offset)
}

// Helper method to enrich results with price and balance data
func (s *PostgreSQLService) enrichResultsWithPriceAndBalance(ctx context.Context, results []map[string]interface{}, icCodes []string) {
	// Load price and balance data
	log.Printf("🏷️ Loading price formula data for %d found items...", len(icCodes))
	priceMap, err := s.LoadPriceFormulaFiltered(ctx, icCodes)
	if err != nil {
		log.Printf("⚠️ Failed to load price formula: %v - using default prices", err)
		priceMap = make(map[string]*PriceInfo)
	}

	log.Printf("📦 Loading balance data for %d found items...", len(icCodes))
	balanceMap, err := s.LoadBalanceDataFiltered(ctx, icCodes)
	if err != nil {
		log.Printf("⚠️ Failed to load balance data: %v - using default balance", err)
		balanceMap = make(map[string]*BalanceInfo)
	}

	// Update results with price and balance data
	for i, result := range results {
		code := result["code"].(string)

		// Look up real price data
		if priceInfo, exists := priceMap[code]; exists {
			salePrice := priceInfo.Price0
			finalPrice := priceInfo.Price0
			discountPrice := priceInfo.Price1

			results[i]["sale_price"] = salePrice
			results[i]["final_price"] = finalPrice
			results[i]["discount_price"] = discountPrice
			results[i]["price"] = salePrice

			log.Printf("💰 Found price for %s: sale_price=%.2f, final_price=%.2f, discount_price=%.2f",
				code, salePrice, finalPrice, discountPrice)
		}

		// Look up real balance data
		if balanceInfo, exists := balanceMap[code]; exists {
			qtyAvailable := balanceInfo.TotalQty
			results[i]["qty_available"] = qtyAvailable
			log.Printf("📦 Found balance for %s: qty_available=%.2f", code, qtyAvailable)
		}
	}
}

// SearchProductsByExactBarcode searches specifically in ic_inventory_barcode.barcode field
func (s *PostgreSQLService) SearchProductsByExactBarcode(ctx context.Context, query string, limit, offset int) ([]map[string]interface{}, int, error) {
	// First check if the ic_inventory_barcode table exists
	checkTableQuery := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'ic_inventory_barcode'`

	var tableExists int
	err := s.db.QueryRowContext(ctx, checkTableQuery).Scan(&tableExists)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check ic_inventory_barcode table existence: %w", err)
	}

	if tableExists == 0 {
		log.Printf("⚠️ Table 'ic_inventory_barcode' not found, skipping barcode search")
		return []map[string]interface{}{}, 0, nil
	}

	// Search for exact barcode match
	whereClause := "CAST(ib.barcode AS TEXT) = $1"

	// Get count of matching records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) as total_count
		FROM ic_inventory_barcode ib
		INNER JOIN ic_inventory i ON CAST(ib.ic_code AS TEXT) = CAST(i.code AS TEXT)
		WHERE %s`, whereClause)

	var totalCount int
	err = s.db.QueryRowContext(ctx, countQuery, query).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute barcode count query: %w", err)
	}

	if totalCount == 0 {
		return []map[string]interface{}{}, 0, nil
	}

	// Build search query
	searchQuery := fmt.Sprintf(`
		SELECT COALESCE(CAST(i.code AS TEXT), 'N/A') as code,
		       COALESCE(CAST(i.name AS TEXT), 'N/A') as name,
		       COALESCE(CAST(i.unit_standard_code AS TEXT), 'N/A') as unit_standard_code,
		       COALESCE(i.item_type, 0) as item_type,
		       COALESCE(i.row_order_ref, 0) as row_order_ref,
		       COALESCE(CAST(ib.barcode AS TEXT), 'N/A') as matched_barcode,
			   COALESCE(CAST(ib.image_url AS TEXT), '') as image_url,
		       10 as search_priority
		FROM ic_inventory_barcode ib
		INNER JOIN ic_inventory i ON CAST(ib.ic_code AS TEXT) = CAST(i.code AS TEXT)
		WHERE %s
		ORDER BY i.name ASC
		LIMIT $2 OFFSET $3`, whereClause)

	// log.Printf("🔍 [BARCODE-SEARCH] SQL Query: %s", searchQuery)
	// log.Printf("🔍 [BARCODE-SEARCH] Parameters: [%s, %d, %d]", query, limit, offset)

	rows, err := s.db.QueryContext(ctx, searchQuery, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute barcode search query: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	var icCodes []string

	for rows.Next() {
		var code, name, unitStandardCode, matchedBarcode, imageURL string
		var itemType, rowOrderRef, searchPriority int

		err := rows.Scan(&code, &name, &unitStandardCode, &itemType, &rowOrderRef, &matchedBarcode, &imageURL, &searchPriority)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan barcode search result: %w", err)
		}

		icCodes = append(icCodes, code)

		result := map[string]interface{}{
			"id":                 code,
			"code":               code,
			"name":               name,
			"unit_standard_code": unitStandardCode,
			"item_type":          itemType,
			"row_order_ref":      rowOrderRef,
			"search_priority":    searchPriority,
			"similarity_score":   float64(searchPriority),
			"matched_barcode":    matchedBarcode,
			"search_method":      "barcode_exact",

			// Default values for pricing and inventory fields
			"sale_price":         0.0,
			"premium_word":       "N/A",
			"discount_price":     0.0,
			"discount_percent":   0.0,
			"final_price":        0.0,
			"sold_qty":           0.0,
			"multi_packing":      0,
			"multi_packing_name": "N/A",
			"barcodes":           matchedBarcode,
			"qty_available":      0.0,
			"description":        "",
			"price":              0.0,
			"balance_qty":        0.0,
			"unit":               unitStandardCode,
			"supplier_code":      "N/A",
			"img_url":            imageURL,
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("barcode search rows iteration error: %w", err)
	}

	// Load price and balance data
	if len(icCodes) > 0 {
		s.enrichResultsWithPriceAndBalance(ctx, results, icCodes)
	}

	log.Printf("✅ [BARCODE-SEARCH] Found %d results for barcode '%s'", len(results), query)
	return results, totalCount, nil
}

// SearchProductsByExactCode searches specifically in ic_inventory.code field
func (s *PostgreSQLService) SearchProductsByExactCode(ctx context.Context, query string, limit, offset int) ([]map[string]interface{}, int, error) {
	// First check if the ic_inventory table exists
	checkTableQuery := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'ic_inventory'`

	var tableExists int
	err := s.db.QueryRowContext(ctx, checkTableQuery).Scan(&tableExists)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check ic_inventory table existence: %w", err)
	}

	if tableExists == 0 {
		return nil, 0, fmt.Errorf("table 'ic_inventory' not found in database")
	}

	// Search for exact code match
	whereClause := "CAST(code AS TEXT) = $1"

	// Get count of matching records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) as total_count
		FROM ic_inventory
		WHERE %s`, whereClause)

	var totalCount int
	err = s.db.QueryRowContext(ctx, countQuery, query).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute code count query: %w", err)
	}

	if totalCount == 0 {
		return []map[string]interface{}{}, 0, nil
	}

	// Build search query
	searchQuery := fmt.Sprintf(`
		SELECT COALESCE(CAST(code AS TEXT), 'N/A') as code,
		       COALESCE(CAST(name AS TEXT), 'N/A') as name,
		       COALESCE(CAST(unit_standard_code AS TEXT), 'N/A') as unit_standard_code,
		       COALESCE(item_type, 0) as item_type,
		       COALESCE(row_order_ref, 0) as row_order_ref,
		       8 as search_priority
		FROM ic_inventory
		WHERE %s
		ORDER BY name ASC
		LIMIT $2 OFFSET $3`, whereClause)

	// log.Printf("🔍 [CODE-SEARCH] SQL Query: %s", searchQuery)
	// log.Printf("🔍 [CODE-SEARCH] Parameters: [%s, %d, %d]", query, limit, offset)

	rows, err := s.db.QueryContext(ctx, searchQuery, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute code search query: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	var icCodes []string

	for rows.Next() {
		var code, name, unitStandardCode string
		var itemType, rowOrderRef, searchPriority int

		err := rows.Scan(&code, &name, &unitStandardCode, &itemType, &rowOrderRef, &searchPriority)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan code search result: %w", err)
		}

		icCodes = append(icCodes, code)

		result := map[string]interface{}{
			"id":                 code,
			"code":               code,
			"name":               name,
			"unit_standard_code": unitStandardCode,
			"item_type":          itemType,
			"row_order_ref":      rowOrderRef,
			"search_priority":    searchPriority,
			"similarity_score":   float64(searchPriority),
			"search_method":      "code_exact",

			// Default values for pricing and inventory fields
			"sale_price":         0.0,
			"premium_word":       "N/A",
			"discount_price":     0.0,
			"discount_percent":   0.0,
			"final_price":        0.0,
			"sold_qty":           0.0,
			"multi_packing":      0,
			"multi_packing_name": "N/A",
			"barcodes":           "N/A",
			"qty_available":      0.0,
			"description":        "",
			"price":              0.0,
			"balance_qty":        0.0,
			"unit":               unitStandardCode,
			"supplier_code":      "N/A",
			"img_url":            "",
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("code search rows iteration error: %w", err)
	}

	// Load price and balance data
	if len(icCodes) > 0 {
		s.enrichResultsWithPriceAndBalance(ctx, results, icCodes)
	}

	log.Printf("✅ [CODE-SEARCH] Found %d results for code '%s'", len(results), query)
	return results, totalCount, nil
}

// SearchProductsByBarcodesWithRelevance performs search on the ic_inventory table using specific barcodes with relevance scores
func (s *PostgreSQLService) SearchProductsByBarcodesWithRelevance(ctx context.Context, barcodes []string, relevanceMap map[string]float64, limit, offset int) ([]map[string]interface{}, int, error) {
	return s.SearchProductsByBarcodesWithRelevanceAndBarcodeMap(ctx, barcodes, relevanceMap, nil, limit, offset)
}

// SearchProductsByBarcodesWithRelevanceAndBarcodeMap performs search with barcode mapping
func (s *PostgreSQLService) SearchProductsByBarcodesWithRelevanceAndBarcodeMap(ctx context.Context, barcodes []string, relevanceMap map[string]float64, barcodeMap map[string]string, limit, offset int) ([]map[string]interface{}, int, error) {
	if len(barcodes) == 0 {
		return []map[string]interface{}{}, 0, nil
	}

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

	if tableExists == 0 {
		return nil, 0, fmt.Errorf("table 'ic_inventory' not found in database - please create the table or contact system administrator")
	}

	// Build IN clause for barcode filtering
	placeholders := make([]string, len(barcodes))
	params := make([]interface{}, len(barcodes))
	for i, barcode := range barcodes {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		params[i] = barcode
	}

	whereClause := fmt.Sprintf("CAST(code AS TEXT) IN (%s)", strings.Join(placeholders, ","))

	// Get count of matching records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) as total_count
		FROM ic_inventory
		WHERE %s`, whereClause)

	countRows, err := s.db.QueryContext(ctx, countQuery, params...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute count query: %w", err)
	}
	defer countRows.Close()

	var totalCount int
	if countRows.Next() {
		if err := countRows.Scan(&totalCount); err != nil {
			return nil, 0, fmt.Errorf("failed to scan count result: %w", err)
		}
	}

	// Build search query with barcode filtering and ordering by relevance (if available) then by name
	var orderClause string
	if relevanceMap != nil && len(relevanceMap) > 0 {
		// Create CASE statement for relevance-based ordering
		var caseClauses []string
		for code, relevance := range relevanceMap {
			caseClauses = append(caseClauses, fmt.Sprintf("WHEN CAST(code AS TEXT) = '%s' THEN %f",
				strings.Replace(code, "'", "''", -1), relevance)) // Escape single quotes
		}
		orderClause = fmt.Sprintf(`ORDER BY
			CASE %s ELSE 0 END DESC,
			name ASC`, strings.Join(caseClauses, " "))
	} else {
		orderClause = "ORDER BY name ASC"
	}

	searchQuery := fmt.Sprintf(`
		SELECT COALESCE(CAST(i.code AS TEXT), 'N/A') as code,
			COALESCE(CAST(i.name AS TEXT), 'N/A') as name,
			COALESCE(CAST(i.unit_standard_code AS TEXT), 'N/A') as unit_standard_code,
			COALESCE(i.item_type, 0) as item_type,
			COALESCE(i.row_order_ref, 0) as row_order_ref,
			COALESCE(ib.image_url, '') as image_url,
			6 as search_priority
		FROM ic_inventory i
		LEFT JOIN (
			SELECT ic_code, MAX(image_url) as image_url
			FROM ic_inventory_barcode
			WHERE image_url IS NOT NULL AND image_url != ''
			GROUP BY ic_code
		) ib ON ib.ic_code = i.code
		WHERE %s
		%s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, len(params)+1, len(params)+2)

	// Add limit and offset to parameters
	params = append(params, limit, offset)

	// log.Printf("🔍 [BARCODE-MAP-SEARCH] SQL Query: %s", searchQuery)
	// log.Printf("🔍 [BARCODE-MAP-SEARCH] Parameters: %v", params)

	rows, err := s.db.QueryContext(ctx, searchQuery, params...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	var icCodes []string

	for rows.Next() {
		var code, name, unitStandardCode, imageURL string
		var itemType, rowOrderRef, searchPriority int

		err := rows.Scan(&code, &name, &unitStandardCode, &itemType, &rowOrderRef, &imageURL, &searchPriority)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan search result: %w", err)
		}

		icCodes = append(icCodes, code)

		// Get relevance score if available
		relevanceScore := 0.0
		if relevanceMap != nil {
			if score, exists := relevanceMap[code]; exists {
				relevanceScore = score
			}
		}

		// Get barcode mapping if available
		mappedBarcode := "N/A"
		if barcodeMap != nil {
			if barcode, exists := barcodeMap[code]; exists {
				mappedBarcode = barcode
			}
		}

		result := map[string]interface{}{
			"id":                 code,
			"code":               code,
			"name":               name,
			"unit_standard_code": unitStandardCode,
			"item_type":          itemType,
			"row_order_ref":      rowOrderRef,
			"search_priority":    searchPriority,
			"similarity_score":   relevanceScore,
			"barcode":            mappedBarcode,
			"search_method":      "barcode_mapping",

			// Default values for pricing and inventory fields
			"sale_price":         0.0,
			"premium_word":       "N/A",
			"discount_price":     0.0,
			"discount_percent":   0.0,
			"final_price":        0.0,
			"sold_qty":           0.0,
			"multi_packing":      0,
			"multi_packing_name": "N/A",
			"barcodes":           mappedBarcode,
			"qty_available":      0.0,
			"description":        "",
			"price":              0.0,
			"balance_qty":        0.0,
			"unit":               unitStandardCode,
			"supplier_code":      "N/A",
			"img_url":            imageURL,
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	// Load price and balance data
	if len(icCodes) > 0 {
		s.enrichResultsWithPriceAndBalance(ctx, results, icCodes)
	}

	log.Printf("✅ [BARCODE-MAP-SEARCH] Found %d results", len(results))
	return results, totalCount, nil
}

// SearchProductsByLikeBarcode performs LIKE search in ic_inventory_barcode.barcode field
func (s *PostgreSQLService) SearchProductsByLikeBarcode(ctx context.Context, query string, limit, offset int) ([]map[string]interface{}, int, error) {
	// First check if the ic_inventory_barcode table exists
	checkTableQuery := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'ic_inventory_barcode'`

	var tableExists int
	err := s.db.QueryRowContext(ctx, checkTableQuery).Scan(&tableExists)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check ic_inventory_barcode table existence: %w", err)
	}

	if tableExists == 0 {
		log.Printf("⚠️ Table 'ic_inventory_barcode' not found, skipping barcode LIKE search")
		return []map[string]interface{}{}, 0, nil
	}

	// Simple LIKE search in barcode field
	whereClause := "CAST(ib.barcode AS TEXT) LIKE $1"

	// Get count of matching records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) as total_count
		FROM ic_inventory_barcode ib
		INNER JOIN ic_inventory i ON CAST(ib.ic_code AS TEXT) = CAST(i.code AS TEXT)
		WHERE %s`, whereClause)

	var totalCount int
	queryWithWildcards := "%" + query + "%"
	err = s.db.QueryRowContext(ctx, countQuery, queryWithWildcards).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute barcode LIKE count query: %w", err)
	}

	if totalCount == 0 {
		return []map[string]interface{}{}, 0, nil
	}

	// Build search query
	searchQuery := fmt.Sprintf(`
		SELECT COALESCE(CAST(i.code AS TEXT), 'N/A') as code,
			COALESCE(CAST(i.name AS TEXT), 'N/A') as name,
			COALESCE(CAST(i.unit_standard_code AS TEXT), 'N/A') as unit_standard_code,
			COALESCE(i.item_type, 0) as item_type,
			COALESCE(i.row_order_ref, 0) as row_order_ref,
			COALESCE(CAST(ib.barcode AS TEXT), 'N/A') as matched_barcode,
			COALESCE(CAST(ib.image_url AS TEXT), '') as image_url,   // <== เพิ่มตรงนี้
			7 as search_priority
		FROM ic_inventory_barcode ib
		INNER JOIN ic_inventory i ON CAST(ib.ic_code AS TEXT) = CAST(i.code AS TEXT)
		WHERE %s
		ORDER BY i.name ASC
		LIMIT $2 OFFSET $3`, whereClause)

	// log.Printf("🔍 [BARCODE-LIKE-SEARCH] SQL Query: %s", searchQuery)
	// log.Printf("🔍 [BARCODE-LIKE-SEARCH] Parameters: [%s, %d, %d]", queryWithWildcards, limit, offset)

	rows, err := s.db.QueryContext(ctx, searchQuery, queryWithWildcards, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute barcode LIKE search query: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	var icCodes []string

	for rows.Next() {
		var code, name, unitStandardCode, matchedBarcode, imageURL string
		var itemType, rowOrderRef, searchPriority int

		err := rows.Scan(&code, &name, &unitStandardCode, &itemType, &rowOrderRef, &matchedBarcode, &imageURL, &searchPriority)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan barcode LIKE search result: %w", err)
		}

		icCodes = append(icCodes, code)

		result := map[string]interface{}{
			"id":                 code,
			"code":               code,
			"name":               name,
			"unit_standard_code": unitStandardCode,
			"item_type":          itemType,
			"row_order_ref":      rowOrderRef,
			"search_priority":    searchPriority,
			"similarity_score":   float64(searchPriority),
			"matched_barcode":    matchedBarcode,
			"search_method":      "barcode_like",

			// Default values for pricing and inventory fields
			"sale_price":         0.0,
			"premium_word":       "N/A",
			"discount_price":     0.0,
			"discount_percent":   0.0,
			"final_price":        0.0,
			"sold_qty":           0.0,
			"multi_packing":      0,
			"multi_packing_name": "N/A",
			"barcodes":           matchedBarcode,
			"qty_available":      0.0,
			"description":        "",
			"price":              0.0,
			"balance_qty":        0.0,
			"unit":               unitStandardCode,
			"supplier_code":      "N/A",
			"img_url":            imageURL,
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("barcode LIKE search rows iteration error: %w", err)
	}

	// Load price and balance data
	if len(icCodes) > 0 {
		s.enrichResultsWithPriceAndBalance(ctx, results, icCodes)
	}

	log.Printf("✅ [BARCODE-LIKE-SEARCH] Found %d results for barcode pattern '%s'", len(results), query)
	return results, totalCount, nil
}

// SearchProductsByLikeCode performs LIKE search in ic_inventory.code field
func (s *PostgreSQLService) SearchProductsByLikeCode(ctx context.Context, query string, limit, offset int) ([]map[string]interface{}, int, error) {
	// First check if the ic_inventory table exists
	checkTableQuery := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'ic_inventory'`

	var tableExists int
	err := s.db.QueryRowContext(ctx, checkTableQuery).Scan(&tableExists)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check ic_inventory table existence: %w", err)
	}

	if tableExists == 0 {
		return nil, 0, fmt.Errorf("table 'ic_inventory' not found in database")
	}

	// Simple LIKE search in code field
	whereClause := "CAST(code AS TEXT) LIKE $1"

	// Get count of matching records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) as total_count
		FROM ic_inventory
		WHERE %s`, whereClause)

	var totalCount int
	queryWithWildcards := "%" + query + "%"
	err = s.db.QueryRowContext(ctx, countQuery, queryWithWildcards).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute code LIKE count query: %w", err)
	}

	if totalCount == 0 {
		return []map[string]interface{}{}, 0, nil
	}

	// Build search query
	searchQuery := fmt.Sprintf(`
		SELECT COALESCE(CAST(code AS TEXT), 'N/A') as code,
		       COALESCE(CAST(name AS TEXT), 'N/A') as name,
		       COALESCE(CAST(unit_standard_code AS TEXT), 'N/A') as unit_standard_code,
		       COALESCE(item_type, 0) as item_type,
		       COALESCE(row_order_ref, 0) as row_order_ref,
		       5 as search_priority
		FROM ic_inventory
		WHERE %s
		ORDER BY name ASC
		LIMIT $2 OFFSET $3`, whereClause)

	// log.Printf("🔍 [CODE-LIKE-SEARCH] SQL Query: %s", searchQuery)
	// log.Printf("🔍 [CODE-LIKE-SEARCH] Parameters: [%s, %d, %d]", queryWithWildcards, limit, offset)

	rows, err := s.db.QueryContext(ctx, searchQuery, queryWithWildcards, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute code LIKE search query: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	var icCodes []string

	for rows.Next() {
		var code, name, unitStandardCode string
		var itemType, rowOrderRef, searchPriority int

		err := rows.Scan(&code, &name, &unitStandardCode, &itemType, &rowOrderRef, &searchPriority)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan code LIKE search result: %w", err)
		}

		icCodes = append(icCodes, code)

		result := map[string]interface{}{
			"id":                 code,
			"code":               code,
			"name":               name,
			"unit_standard_code": unitStandardCode,
			"item_type":          itemType,
			"row_order_ref":      rowOrderRef,
			"search_priority":    searchPriority,
			"similarity_score":   float64(searchPriority),
			"search_method":      "code_like",

			// Default values for pricing and inventory fields
			"sale_price":         0.0,
			"premium_word":       "N/A",
			"discount_price":     0.0,
			"discount_percent":   0.0,
			"final_price":        0.0,
			"sold_qty":           0.0,
			"multi_packing":      0,
			"multi_packing_name": "N/A",
			"barcodes":           "N/A",
			"qty_available":      0.0,
			"description":        "",
			"price":              0.0,
			"balance_qty":        0.0,
			"unit":               unitStandardCode,
			"supplier_code":      "N/A",
			"img_url":            "",
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("code LIKE search rows iteration error: %w", err)
	}

	// Load price and balance data
	if len(icCodes) > 0 {
		s.enrichResultsWithPriceAndBalance(ctx, results, icCodes)
	}

	log.Printf("✅ [CODE-LIKE-SEARCH] Found %d results for code pattern '%s'", len(results), query)
	return results, totalCount, nil
}

// SearchProductsSimpleLike performs simple LIKE search in both barcode and code fields
func (s *PostgreSQLService) SearchProductsSimpleLike(ctx context.Context, query string, limit, offset int) ([]map[string]interface{}, int, error) {
	if strings.TrimSpace(query) == "" {
		return []map[string]interface{}{}, 0, nil
	}

	// Add wildcards to query
	queryWithWildcards := "%" + query + "%"

	log.Printf("🔍 [SIMPLE-LIKE-SEARCH] Searching for: '%s' in both barcode and code fields", query)

	// Check if tables exist
	checkBarcodeTableQuery := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'ic_inventory_barcode'`

	checkInventoryTableQuery := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'ic_inventory'`

	var barcodeTableExists, inventoryTableExists int

	err := s.db.QueryRowContext(ctx, checkBarcodeTableQuery).Scan(&barcodeTableExists)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check barcode table existence: %w", err)
	}

	err = s.db.QueryRowContext(ctx, checkInventoryTableQuery).Scan(&inventoryTableExists)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check inventory table existence: %w", err)
	}

	if inventoryTableExists == 0 {
		return nil, 0, fmt.Errorf("table 'ic_inventory' not found in database")
	}

	var unionQuery string
	var params []interface{}

	if barcodeTableExists > 0 {
		// Union query to search both barcode and code fields
		unionQuery = `
		SELECT * FROM (
			-- Search in barcode table
			SELECT DISTINCT
				COALESCE(CAST(i.code AS TEXT), 'N/A') as code,
				COALESCE(CAST(i.name AS TEXT), 'N/A') as name,
				COALESCE(CAST(i.unit_standard_code AS TEXT), 'N/A') as unit_standard_code,
				COALESCE(i.item_type, 0) as item_type,
				COALESCE(i.row_order_ref, 0) as row_order_ref,
				COALESCE(CAST(ib.barcode AS TEXT), 'N/A') as matched_barcode,
				COALESCE(CAST(ib.image_url AS TEXT), '') as image_url,
				'barcode' as search_source,
				9 as search_priority
			FROM ic_inventory_barcode ib
			INNER JOIN ic_inventory i ON CAST(ib.ic_code AS TEXT) = CAST(i.code AS TEXT)
			WHERE CAST(ib.barcode AS TEXT) LIKE $1

			UNION

			-- Search in code field
			SELECT DISTINCT
				COALESCE(CAST(code AS TEXT), 'N/A') as code,
				COALESCE(CAST(name AS TEXT), 'N/A') as name,
				COALESCE(CAST(unit_standard_code AS TEXT), 'N/A') as unit_standard_code,
				COALESCE(item_type, 0) as item_type,
				COALESCE(row_order_ref, 0) as row_order_ref,
				'N/A' as matched_barcode,
				'code' as search_source,
				7 as search_priority
			FROM ic_inventory
			WHERE CAST(code AS TEXT) LIKE $2
		) combined_results
		ORDER BY search_priority DESC, name ASC
		LIMIT $3 OFFSET $4`

		params = []interface{}{queryWithWildcards, queryWithWildcards, limit, offset}
	} else {
		// Only search in code field if barcode table doesn't exist
		log.Printf("⚠️ Table 'ic_inventory_barcode' not found, searching only in code field")
		unionQuery = `
		SELECT DISTINCT
			COALESCE(CAST(code AS TEXT), 'N/A') as code,
			COALESCE(CAST(name AS TEXT), 'N/A') as name,
			COALESCE(CAST(unit_standard_code AS TEXT), 'N/A') as unit_standard_code,
			COALESCE(item_type, 0) as item_type,
			COALESCE(row_order_ref, 0) as row_order_ref,
			'N/A' as matched_barcode,
			'code' as search_source,
			7 as search_priority
		FROM ic_inventory
		WHERE CAST(code AS TEXT) LIKE $1
		ORDER BY name ASC
		LIMIT $2 OFFSET $3`

		params = []interface{}{queryWithWildcards, limit, offset}
	}

	log.Printf("🔍 [SIMPLE-LIKE-SEARCH] SQL Query: %s", unionQuery)
	log.Printf("🔍 [SIMPLE-LIKE-SEARCH] Parameters: %v", params)

	rows, err := s.db.QueryContext(ctx, unionQuery, params...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute simple LIKE search query: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	var icCodes []string

	for rows.Next() {
		var code, name, unitStandardCode, matchedBarcode, imageURL, searchSource string
		var itemType, rowOrderRef, searchPriority int

		err := rows.Scan(&code, &name, &unitStandardCode, &itemType, &rowOrderRef, &matchedBarcode, &imageURL, &searchSource, &searchPriority)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan simple LIKE search result: %w", err)
		}

		icCodes = append(icCodes, code)

		result := map[string]interface{}{
			"id":                 code,
			"code":               code,
			"name":               name,
			"unit_standard_code": unitStandardCode,
			"item_type":          itemType,
			"row_order_ref":      rowOrderRef,
			"search_priority":    searchPriority,
			"similarity_score":   float64(searchPriority),
			"matched_barcode":    matchedBarcode,
			"search_method":      "simple_like_" + searchSource,
			"search_source":      searchSource,

			// Default values for pricing and inventory fields
			"sale_price":         0.0,
			"premium_word":       "N/A",
			"discount_price":     0.0,
			"discount_percent":   0.0,
			"final_price":        0.0,
			"sold_qty":           0.0,
			"multi_packing":      0,
			"multi_packing_name": "N/A",
			"barcodes":           matchedBarcode,
			"qty_available":      0.0,
			"description":        "",
			"price":              0.0,
			"balance_qty":        0.0,
			"unit":               unitStandardCode,
			"supplier_code":      "N/A",
			"img_url":            imageURL,
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("simple LIKE search rows iteration error: %w", err)
	}

	// Get total count for pagination
	var countQuery string
	var countParams []interface {
	}

	if barcodeTableExists > 0 {
		countQuery = `
		SELECT COUNT(*) FROM (
			SELECT DISTINCT i.code
			FROM ic_inventory_barcode ib
			INNER JOIN ic_inventory i ON CAST(ib.ic_code AS TEXT) = CAST(i.code AS TEXT)
			WHERE CAST(ib.barcode AS TEXT) LIKE $1

			UNION

			SELECT DISTINCT code
			FROM ic_inventory
			WHERE CAST(code AS TEXT) LIKE $2
		) combined_count`
		countParams = []interface{}{queryWithWildcards, queryWithWildcards}
	} else {
		countQuery = `
		SELECT COUNT(DISTINCT code)
		FROM ic_inventory
		WHERE CAST(code AS TEXT) LIKE $1`
		countParams = []interface{}{queryWithWildcards}
	}

	var totalCount int
	err = s.db.QueryRowContext(ctx, countQuery, countParams...).Scan(&totalCount)
	if err != nil {
		log.Printf("⚠️ Failed to get total count: %v", err)
		totalCount = len(results) // Fallback to result count
	}

	// Load price and balance data
	if len(icCodes) > 0 {
		s.enrichResultsWithPriceAndBalance(ctx, results, icCodes)
	}

	log.Printf("✅ [SIMPLE-LIKE-SEARCH] Found %d results for pattern '%s'", len(results), query)
	return results, totalCount, nil
}

// SearchProductsByBarcodesWithRelevanceAndBarcodeMapInCollection performs search with barcode mapping in specific database/collection
func (s *PostgreSQLService) SearchProductsByBarcodesWithRelevanceAndBarcodeMapInCollection(ctx context.Context, barcodes []string, relevanceMap map[string]float64, barcodeMap map[string]string, limit, offset int, collection string) ([]map[string]interface{}, int, error) {
	if len(barcodes) == 0 {
		return []map[string]interface{}{}, 0, nil
	}

	// Get database name from collection (convert to lowercase)
	databaseName := strings.ToLower(collection)
	if databaseName == "" {
		// Fallback to default database from config
		databaseName = s.config.PostgreSQL.Database
	}

	// Create new connection for the specific database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		s.config.PostgreSQL.Host,
		s.config.PostgreSQL.Port,
		s.config.PostgreSQL.User,
		s.config.PostgreSQL.Password,
		databaseName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to connect to database '%s': %w", databaseName, err)
	}
	defer db.Close()

	// Test connection to the specific database
	if err := db.Ping(); err != nil {
		return nil, 0, fmt.Errorf("failed to ping database '%s': %w", databaseName, err)
	}

	log.Printf("🔗 [COLLECTION-SEARCH] Connected to database: %s", databaseName)

	// First check if the ic_inventory table exists
	checkTableQuery := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'ic_inventory'`

	var tableExists int
	err = db.QueryRowContext(ctx, checkTableQuery).Scan(&tableExists)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check table existence in database '%s': %w", databaseName, err)
	}

	if tableExists == 0 {
		return nil, 0, fmt.Errorf("table 'ic_inventory' not found in database '%s' - please create the table or contact system administrator", databaseName)
	}

	// Build IN clause for barcode filtering
	placeholders := make([]string, len(barcodes))
	params := make([]interface{}, len(barcodes))
	for i, barcode := range barcodes {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		params[i] = barcode
	}

	whereClause := fmt.Sprintf("CAST(code AS TEXT) IN (%s)", strings.Join(placeholders, ","))

	// Get count of matching records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) as total_count
		FROM ic_inventory
		WHERE %s`, whereClause)

	countRows, err := db.QueryContext(ctx, countQuery, params...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute count query in database '%s': %w", databaseName, err)
	}
	defer countRows.Close()

	var totalCount int
	if countRows.Next() {
		if err := countRows.Scan(&totalCount); err != nil {
			return nil, 0, fmt.Errorf("failed to scan count result: %w", err)
		}
	}

	// Build search query with barcode filtering and ordering by relevance (if available) then by name
	var orderClause string
	if relevanceMap != nil && len(relevanceMap) > 0 {
		// Create CASE statement for relevance-based ordering
		var caseClauses []string
		for code, relevance := range relevanceMap {
			caseClauses = append(caseClauses, fmt.Sprintf("WHEN CAST(code AS TEXT) = '%s' THEN %f",
				strings.Replace(code, "'", "''", -1), relevance)) // Escape single quotes
		}
		orderClause = fmt.Sprintf(`ORDER BY
			CASE %s ELSE 0 END DESC,
			name ASC`, strings.Join(caseClauses, " "))
	} else {
		orderClause = "ORDER BY name ASC"
	}

	searchQuery := fmt.Sprintf(`
		SELECT COALESCE(CAST(i.code AS TEXT), 'N/A') as code,
			COALESCE(CAST(i.name AS TEXT), 'N/A') as name,
			COALESCE(CAST(i.unit_standard_code AS TEXT), 'N/A') as unit_standard_code,
			COALESCE(i.item_type, 0) as item_type,
			COALESCE(i.row_order_ref, 0) as row_order_ref,
			COALESCE(ib.image_url, '') as image_url,
			6 as search_priority
		FROM ic_inventory i
		LEFT JOIN (
			SELECT ic_code, MAX(image_url) as image_url
			FROM ic_inventory_barcode
			WHERE image_url IS NOT NULL AND image_url != ''
			GROUP BY ic_code
		) ib ON ib.ic_code = i.code
		WHERE %s
		%s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, len(params)+1, len(params)+2)

	// Add limit and offset to parameters
	params = append(params, limit, offset)

	// log.Printf("🔍 [COLLECTION-SEARCH] Database: %s, SQL Query: %s", databaseName, searchQuery)
	// log.Printf("🔍 [COLLECTION-SEARCH] Parameters: %v", params)

	rows, err := db.QueryContext(ctx, searchQuery, params...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search query in database '%s': %w", databaseName, err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	var icCodes []string

	for rows.Next() {
		var code, name, unitStandardCode, imageURL string
		var itemType, rowOrderRef, searchPriority int

		err := rows.Scan(&code, &name, &unitStandardCode, &itemType, &rowOrderRef, &imageURL, &searchPriority)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan search result: %w", err)
		}

		icCodes = append(icCodes, code)

		// Get relevance score if available
		relevanceScore := 0.0
		if relevanceMap != nil {
			if score, exists := relevanceMap[code]; exists {
				relevanceScore = score
			}
		}

		// Get barcode mapping if available
		mappedBarcode := "N/A"
		if barcodeMap != nil {
			if barcode, exists := barcodeMap[code]; exists {
				mappedBarcode = barcode
			}
		}

		result := map[string]interface{}{
			"code":             code,
			"name":             name,
			"unit":             unitStandardCode,
			"item_type":        itemType,
			"row_order_ref":    rowOrderRef,
			"img_url":          imageURL,
			"similarity_score": relevanceScore,
			"search_priority":  searchPriority,
			"barcode":          mappedBarcode,
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error reading search results: %w", err)
	}

	log.Printf("🔍 [COLLECTION-SEARCH] Found %d IC codes from database '%s', fetching additional data...", len(icCodes), databaseName)

	// Enhanced data loading with proper error handling
	if len(icCodes) > 0 {
		// Load prices and balance using the same database connection
		priceResults, priceErr := s.loadPricesForICCodesInDatabase(ctx, db, icCodes, databaseName)
		if priceErr != nil {
			log.Printf("⚠️ [COLLECTION-SEARCH] Price loading failed for database '%s': %v", databaseName, priceErr)
		}

		balanceResults, balanceErr := s.loadBalanceForICCodesInDatabase(ctx, db, icCodes, databaseName)
		if balanceErr != nil {
			log.Printf("⚠️ [COLLECTION-SEARCH] Balance loading failed for database '%s': %v", databaseName, balanceErr)
		}

		// Merge additional data into results
		for i := range results {
			code := results[i]["code"].(string)

			// Add price data
			if priceData, exists := priceResults[code]; exists {
				for key, value := range priceData {
					results[i][key] = value
				}
			}

			// Add balance data
			if balanceData, exists := balanceResults[code]; exists {
				for key, value := range balanceData {
					results[i][key] = value
				}
			}
		}

		log.Printf("💰 [COLLECTION-SEARCH] Price data: %d/%d products have pricing in database '%s'", len(priceResults), len(icCodes), databaseName)
		log.Printf("📦 [COLLECTION-SEARCH] Balance data: %d/%d products have stock info in database '%s'", len(balanceResults), len(icCodes), databaseName)
	}

	log.Printf("✅ [COLLECTION-SEARCH] Search completed in database '%s': found %d results, total count: %d", databaseName, len(results), totalCount)

	return results, totalCount, nil
}

// Helper methods for loading additional data in specific database
func (s *PostgreSQLService) loadPricesForICCodesInDatabase(ctx context.Context, db *sql.DB, icCodes []string, databaseName string) (map[string]map[string]interface{}, error) {
	results := make(map[string]map[string]interface{})

	if len(icCodes) == 0 {
		return results, nil
	}

	// Build IN clause
	placeholders := make([]string, len(icCodes))
	params := make([]interface{}, len(icCodes))
	for i, code := range icCodes {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		params[i] = code
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(CAST(ic_code AS TEXT), 'N/A') as ic_code,
			COALESCE(price, 0) as price,
			COALESCE(CAST(supplier_code AS TEXT), 'N/A') as supplier_code
		FROM ic_price
		WHERE CAST(ic_code AS TEXT) IN (%s)`, strings.Join(placeholders, ","))

	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		return results, fmt.Errorf("failed to load prices from database '%s': %w", databaseName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var icCode, supplierCode string
		var price float64

		if err := rows.Scan(&icCode, &price, &supplierCode); err != nil {
			continue // Skip problematic rows
		}

		results[icCode] = map[string]interface{}{
			"price":         price,
			"supplier_code": supplierCode,
		}
	}

	return results, nil
}

func (s *PostgreSQLService) loadBalanceForICCodesInDatabase(ctx context.Context, db *sql.DB, icCodes []string, databaseName string) (map[string]map[string]interface{}, error) {
	results := make(map[string]map[string]interface{})

	if len(icCodes) == 0 {
		return results, nil
	}

	// Build IN clause
	placeholders := make([]string, len(icCodes))
	params := make([]interface{}, len(icCodes))
	for i, code := range icCodes {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		params[i] = code
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(CAST(ic_code AS TEXT), 'N/A') as ic_code,
			COALESCE(balance_qty, 0) as balance_qty
		FROM ic_balance
		WHERE CAST(ic_code AS TEXT) IN (%s)`, strings.Join(placeholders, ","))

	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		return results, fmt.Errorf("failed to load balance from database '%s': %w", databaseName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var icCode string
		var balanceQty float64

		if err := rows.Scan(&icCode, &balanceQty); err != nil {
			continue // Skip problematic rows
		}

		results[icCode] = map[string]interface{}{
			"balance_qty": balanceQty,
		}
	}

	return results, nil
}

// SearchProductsByExactBarcodeInCollection searches specifically in ic_inventory_barcode.barcode field in specific database
func (s *PostgreSQLService) SearchProductsByExactBarcodeInCollection(ctx context.Context, query string, limit, offset int, collection string) ([]map[string]interface{}, int, error) {
	// Get database name from collection (convert to lowercase)
	databaseName := strings.ToLower(collection)
	if databaseName == "" {
		// Fallback to default database from config
		databaseName = s.config.PostgreSQL.Database
	}

	// Create new connection for the specific database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		s.config.PostgreSQL.Host,
		s.config.PostgreSQL.Port,
		s.config.PostgreSQL.User,
		s.config.PostgreSQL.Password,
		databaseName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to connect to database '%s': %w", databaseName, err)
	}
	defer db.Close()

	// Test connection to the specific database
	if err := db.Ping(); err != nil {
		return nil, 0, fmt.Errorf("failed to ping database '%s': %w", databaseName, err)
	}

	log.Printf("🔗 [EXACT-BARCODE-SEARCH] Connected to database: %s", databaseName)

	// First check if the ic_inventory_barcode table exists
	checkTableQuery := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'ic_inventory_barcode'`

	var tableExists int
	err = db.QueryRowContext(ctx, checkTableQuery).Scan(&tableExists)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check ic_inventory_barcode table existence in database '%s': %w", databaseName, err)
	}

	if tableExists == 0 {
		log.Printf("⚠️ Table 'ic_inventory_barcode' not found in database '%s', skipping barcode search", databaseName)
		return []map[string]interface{}{}, 0, nil
	}

	// Search for exact barcode match
	whereClause := "CAST(ib.barcode AS TEXT) = $1"

	// Get count of matching records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) as total_count
		FROM ic_inventory_barcode ib
		INNER JOIN ic_inventory i ON CAST(ib.ic_code AS TEXT) = CAST(i.code AS TEXT)
		WHERE %s`, whereClause)

	var totalCount int
	err = db.QueryRowContext(ctx, countQuery, query).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute barcode count query in database '%s': %w", databaseName, err)
	}

	if totalCount == 0 {
		return []map[string]interface{}{}, 0, nil
	}

	// Build search query
	searchQuery := fmt.Sprintf(`
		SELECT COALESCE(CAST(i.code AS TEXT), 'N/A') as code,
			COALESCE(CAST(i.name AS TEXT), 'N/A') as name,
			COALESCE(CAST(ib.barcode AS TEXT), 'N/A') as barcode,
			COALESCE(CAST(i.unit_standard_code AS TEXT), 'N/A') as unit_standard_code,
			COALESCE(ib.image_url, '') as image_url,
			100.0 as similarity_score,
			1 as search_priority
		FROM ic_inventory_barcode ib
		INNER JOIN ic_inventory i ON CAST(ib.ic_code AS TEXT) = CAST(i.code AS TEXT)
		WHERE %s
		ORDER BY i.name ASC
		LIMIT $2 OFFSET $3`, whereClause)

	log.Printf("🔍 [EXACT-BARCODE-SEARCH] Database: %s, SQL Query: %s", databaseName, searchQuery)
	log.Printf("🔍 [EXACT-BARCODE-SEARCH] Parameters: [%s, %d, %d]", query, limit, offset)

	rows, err := db.QueryContext(ctx, searchQuery, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute barcode search query in database '%s': %w", databaseName, err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var code, name, barcode, unitStandardCode, imageURL string
		var similarityScore float64
		var searchPriority int

		err := rows.Scan(&code, &name, &barcode, &unitStandardCode, &imageURL, &similarityScore, &searchPriority)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan barcode search result: %w", err)
		}

		result := map[string]interface{}{
			"code":             code,
			"name":             name,
			"barcode":          barcode,
			"unit":             unitStandardCode,
			"img_url":          imageURL,
			"similarity_score": similarityScore,
			"search_priority":  searchPriority,
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error reading barcode search results: %w", err)
	}

	log.Printf("✅ [EXACT-BARCODE-SEARCH] Search completed in database '%s': found %d results, total count: %d", databaseName, len(results), totalCount)

	return results, totalCount, nil
}

// SearchProductsByExactCodeInCollection searches specifically for exact ic_code match in specific database
func (s *PostgreSQLService) SearchProductsByExactCodeInCollection(ctx context.Context, query string, limit, offset int, collection string) ([]map[string]interface{}, int, error) {
	// Get database name from collection (convert to lowercase)
	databaseName := strings.ToLower(collection)
	if databaseName == "" {
		// Fallback to default database from config
		databaseName = s.config.PostgreSQL.Database
	}

	// Create new connection for the specific database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		s.config.PostgreSQL.Host,
		s.config.PostgreSQL.Port,
		s.config.PostgreSQL.User,
		s.config.PostgreSQL.Password,
		databaseName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to connect to database '%s': %w", databaseName, err)
	}
	defer db.Close()

	// Test connection to the specific database
	if err := db.Ping(); err != nil {
		return nil, 0, fmt.Errorf("failed to ping database '%s': %w", databaseName, err)
	}

	log.Printf("🔗 [EXACT-CODE-SEARCH] Connected to database: %s", databaseName)

	// First check if the ic_inventory table exists
	checkTableQuery := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'ic_inventory'`

	var tableExists int
	err = db.QueryRowContext(ctx, checkTableQuery).Scan(&tableExists)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check table existence in database '%s': %w", databaseName, err)
	}

	if tableExists == 0 {
		return nil, 0, fmt.Errorf("table 'ic_inventory' not found in database '%s' - please create the table or contact system administrator", databaseName)
	}

	// Search for exact code match
	whereClause := "CAST(code AS TEXT) = $1"

	// Get count of matching records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) as total_count
		FROM ic_inventory
		WHERE %s`, whereClause)

	var totalCount int
	err = db.QueryRowContext(ctx, countQuery, query).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute code count query in database '%s': %w", databaseName, err)
	}

	if totalCount == 0 {
		return []map[string]interface{}{}, 0, nil
	}

	// Build search query
	searchQuery := fmt.Sprintf(`
		SELECT COALESCE(CAST(i.code AS TEXT), 'N/A') as code,
			COALESCE(CAST(i.name AS TEXT), 'N/A') as name,
			COALESCE(CAST(i.unit_standard_code AS TEXT), 'N/A') as unit_standard_code,
			COALESCE(i.item_type, 0) as item_type,
			COALESCE(i.row_order_ref, 0) as row_order_ref,
			COALESCE(ib.image_url, '') as image_url,
			100.0 as similarity_score,
			1 as search_priority
		FROM ic_inventory i
		LEFT JOIN (
			SELECT ic_code, MAX(image_url) as image_url
			FROM ic_inventory_barcode
			WHERE image_url IS NOT NULL AND image_url != ''
			GROUP BY ic_code
		) ib ON ib.ic_code = i.code
		WHERE %s
		ORDER BY i.name ASC
		LIMIT $2 OFFSET $3`, whereClause)

	log.Printf("🔍 [EXACT-CODE-SEARCH] Database: %s, SQL Query: %s", databaseName, searchQuery)
	log.Printf("🔍 [EXACT-CODE-SEARCH] Parameters: [%s, %d, %d]", query, limit, offset)

	rows, err := db.QueryContext(ctx, searchQuery, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute code search query in database '%s': %w", databaseName, err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var code, name, unitStandardCode, imageURL string
		var itemType, rowOrderRef, searchPriority int
		var similarityScore float64

		err := rows.Scan(&code, &name, &unitStandardCode, &itemType, &rowOrderRef, &imageURL, &similarityScore, &searchPriority)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan code search result: %w", err)
		}

		result := map[string]interface{}{
			"code":             code,
			"name":             name,
			"unit":             unitStandardCode,
			"item_type":        itemType,
			"row_order_ref":    rowOrderRef,
			"img_url":          imageURL,
			"similarity_score": similarityScore,
			"search_priority":  searchPriority,
			"barcode":          "N/A", // Add default barcode field
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error reading code search results: %w", err)
	}

	log.Printf("✅ [EXACT-CODE-SEARCH] Search completed in database '%s': found %d results, total count: %d", databaseName, len(results), totalCount)

	return results, totalCount, nil
}

// ExecuteCommandInCollection executes a SQL command in the specified database/collection
func (s *PostgreSQLService) ExecuteCommandInCollection(ctx context.Context, query string, collection string) (interface{}, error) {
	// Get database name from collection (convert to lowercase)
	databaseName := strings.ToLower(collection)
	if databaseName == "" {
		// Fallback to default database from config
		databaseName = s.config.PostgreSQL.Database
	}

	// Create new connection for the specific database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		s.config.PostgreSQL.Host,
		s.config.PostgreSQL.Port,
		s.config.PostgreSQL.User,
		s.config.PostgreSQL.Password,
		databaseName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database '%s': %w", databaseName, err)
	}
	defer db.Close()

	// Test connection to the specific database
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database '%s': %w", databaseName, err)
	}

	log.Printf("🔗 [PG-COMMAND-COLLECTION] Connected to database: %s", databaseName)

	// Execute the command
	result, err := db.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("command execution failed in database '%s': %w", databaseName, err)
	}

	// Get rows affected if possible
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// Some commands might not return rows affected, return basic success
		return map[string]interface{}{
			"status":   "success",
			"query":    query,
			"database": databaseName,
		}, nil
	}

	return map[string]interface{}{
		"status":        "success",
		"rows_affected": rowsAffected,
		"query":         query,
		"database":      databaseName,
	}, nil
}

// ExecuteSelectInCollection executes a SELECT query in the specified database/collection
func (s *PostgreSQLService) ExecuteSelectInCollection(ctx context.Context, query string, collection string) ([]interface{}, error) {
	// Get database name from collection (convert to lowercase)
	databaseName := strings.ToLower(collection)
	if databaseName == "" {
		// Fallback to default database from config
		databaseName = s.config.PostgreSQL.Database
	}

	// Create new connection for the specific database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		s.config.PostgreSQL.Host,
		s.config.PostgreSQL.Port,
		s.config.PostgreSQL.User,
		s.config.PostgreSQL.Password,
		databaseName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database '%s': %w", databaseName, err)
	}
	defer db.Close()

	// Test connection to the specific database
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database '%s': %w", databaseName, err)
	}

	log.Printf("🔗 [PG-SELECT-COLLECTION] Connected to database: %s", databaseName)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed in database '%s': %w", databaseName, err)
	}
	defer rows.Close()

	// Get column information
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns in database '%s': %w", databaseName, err)
	}

	var results []interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row in database '%s': %w", databaseName, err)
		}

		// Create a map for this row
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				// Convert []byte to string
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		results = append(results, rowMap)
	}

	return results, nil
}
