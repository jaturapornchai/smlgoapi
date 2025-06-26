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
	log.Printf("🔍 SQL Query: %s", searchQuery)
	log.Printf("🔍 Parameters: %v", searchParams)

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
	return s.SearchProductsByBarcodesWithRelevance(ctx, barcodes, nil, limit, offset)
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
		SELECT COALESCE(CAST(code AS TEXT), 'N/A') as code, 
		       COALESCE(CAST(name AS TEXT), 'N/A') as name,
		       COALESCE(CAST(unit_standard_code AS TEXT), 'N/A') as unit_standard_code,
		       COALESCE(item_type, 0) as item_type,
		       COALESCE(row_order_ref, 0) as row_order_ref,
		       5 as search_priority
		FROM ic_inventory 
		WHERE %s
		%s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, len(params)+1, len(params)+2)

	// Prepare parameters for search query
	searchParams := make([]interface{}, 0)
	searchParams = append(searchParams, params...) // barcode parameters
	searchParams = append(searchParams, limit)     // limit
	searchParams = append(searchParams, offset)    // offset

	// Log the search summary instead of full SQL
	log.Printf("🔍 [PostgreSQL] Searching by %s codes: %d items, limit=%d, offset=%d",
		func() string {
			if strings.Contains(searchQuery, "ORDER BY") && strings.Contains(searchQuery, "CASE") {
				return "IC/Barcode (with relevance)"
			}
			return "IC/Barcode"
		}(), len(params), limit, offset)

	rows, err := s.db.QueryContext(ctx, searchQuery, searchParams...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute barcode search query: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	var icCodes []string // Collect ic_codes for filtered price/balance loading

	for rows.Next() {
		var code, name, unitStandardCode string
		var itemType, rowOrderRef, searchPriority int

		err := rows.Scan(&code, &name, &unitStandardCode, &itemType, &rowOrderRef, &searchPriority)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan barcode search result: %w", err)
		}

		icCodes = append(icCodes, code) // Collect ic_code for later price/balance lookup

		// Default values for pricing and inventory fields
		var salePrice, discountPrice, discountPercent, finalPrice, soldQty, qtyAvailable float64 = 0.0, 0.0, 0.0, 0.0, 0.0, 0.0
		premiumWord := "N/A"
		multiPacking := 0
		multiPackingName := "N/A"
		barcodesField := code // Use the code as default barcodes field

		// Get the actual barcode from the mapping if available
		actualBarcode := code // Default to code
		if barcodeMap != nil {
			if mappedBarcode, exists := barcodeMap[code]; exists {
				actualBarcode = mappedBarcode
			}
		}

		// Get relevance score from map if available
		var relevanceScore float64 = float64(searchPriority) // Default to search priority
		if relevanceMap != nil {
			if score, exists := relevanceMap[code]; exists {
				relevanceScore = score
			}
		}

		result := map[string]interface{}{
			"id":                 code, // Use code as id since there's no separate id field
			"code":               code,
			"name":               name,
			"unit_standard_code": unitStandardCode,
			"item_type":          itemType,
			"row_order_ref":      rowOrderRef,
			"search_priority":    searchPriority,
			"similarity_score":   relevanceScore, // Use relevance score from Weaviate

			// Pricing and inventory fields (will be updated below)
			"sale_price":         salePrice,
			"premium_word":       premiumWord,
			"discount_price":     discountPrice,
			"discount_percent":   discountPercent,
			"final_price":        finalPrice,
			"sold_qty":           soldQty,
			"multi_packing":      multiPacking,
			"multi_packing_name": multiPackingName,
			"barcodes":           barcodesField,
			"barcode":            actualBarcode, // Add the actual barcode from Weaviate
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
	if len(icCodes) > 0 {
		log.Printf("🏷️ [PostgreSQL] Loading price data for %d products...", len(icCodes))
		priceMap, err := s.LoadPriceFormulaFiltered(ctx, icCodes)
		if err != nil {
			log.Printf("⚠️ [PostgreSQL] Failed to load price formula: %v - using default prices", err)
			priceMap = make(map[string]*PriceInfo)
		} else {
			log.Printf("✅ [PostgreSQL] Loaded price data for %d products", len(priceMap))
		}

		log.Printf("📦 [PostgreSQL] Loading balance data for %d products...", len(icCodes))
		balanceMap, err := s.LoadBalanceDataFiltered(ctx, icCodes)
		if err != nil {
			log.Printf("⚠️ [PostgreSQL] Failed to load balance data: %v - using default balance", err)
			balanceMap = make(map[string]*BalanceInfo)
		} else {
			log.Printf("✅ [PostgreSQL] Loaded balance data for %d products", len(balanceMap))
		}

		// Update results with price and balance data
		priceFoundCount := 0
		balanceFoundCount := 0
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
				priceFoundCount++
			}

			// Look up real balance data
			if balanceInfo, exists := balanceMap[code]; exists {
				qtyAvailable := balanceInfo.TotalQty // Use sum of balance_qty as qty_available
				results[i]["qty_available"] = qtyAvailable
				balanceFoundCount++
			}
		}

		log.Printf("� [PostgreSQL] Price data: %d/%d products have pricing", priceFoundCount, len(results))
		log.Printf("📦 [PostgreSQL] Balance data: %d/%d products have stock info", balanceFoundCount, len(results))
	}

	log.Printf("✅ [PostgreSQL] Search completed: found %d results, total count: %d", len(results), totalCount)
	return results, totalCount, nil
}
