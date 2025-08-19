package services

import (
	"context"
	"fmt"
	"log"
)

// SearchProductsByExactBarcode searches for products with exact barcode match
func (s *PostgreSQLService) SearchProductsByExactBarcode(ctx context.Context, query string, limit, offset int) ([]map[string]interface{}, int, error) {
	log.Printf("🔍 [EXACT-BARCODE-SEARCH] Starting search for barcode: '%s'", query)

	searchQuery := `
		SELECT DISTINCT 
			ic.code,
			ic.name_th,
			ic.name_en,
			ic.brand,
			ic.category,
			ic.unit,
			ic.image_url,
			ib.barcode
		FROM ic_inventory ic
		INNER JOIN ic_inventory_barcode ib ON ic.code = ib.ic_code
		WHERE LOWER(ib.barcode) = LOWER($1)
		ORDER BY ic.code
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, searchQuery, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute exact barcode search: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	var icCodes []string

	for rows.Next() {
		var code, nameTh, nameEn, brand, category, unit, imageURL, barcode string
		err := rows.Scan(&code, &nameTh, &nameEn, &brand, &category, &unit, &imageURL, &barcode)
		if err != nil {
			log.Printf("⚠️ Failed to scan row: %v", err)
			continue
		}

		result := map[string]interface{}{
			"code":      code,
			"name_th":   nameTh,
			"name_en":   nameEn,
			"brand":     brand,
			"category":  category,
			"unit":      unit,
			"image_url": imageURL,
			"barcode":   barcode,
		}

		results = append(results, result)
		icCodes = append(icCodes, code)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	// Get total count
	countQuery := `
		SELECT COUNT(DISTINCT ic.code)
		FROM ic_inventory ic
		INNER JOIN ic_inventory_barcode ib ON ic.code = ib.ic_code
		WHERE LOWER(ib.barcode) = LOWER($1)
	`

	var totalCount int
	err = s.db.QueryRowContext(ctx, countQuery, query).Scan(&totalCount)
	if err != nil {
		log.Printf("⚠️ Failed to get total count: %v", err)
		totalCount = len(results)
	}

	// Load price data
	if len(icCodes) > 0 {
		s.enrichResultsWithPrice(ctx, results, icCodes)
	}

	log.Printf("✅ [EXACT-BARCODE-SEARCH] Found %d results for barcode '%s'", len(results), query)
	return results, totalCount, nil
}

// SearchProductsByExactCode searches for products with exact code match
func (s *PostgreSQLService) SearchProductsByExactCode(ctx context.Context, query string, limit, offset int) ([]map[string]interface{}, int, error) {
	log.Printf("🔍 [EXACT-CODE-SEARCH] Starting search for code: '%s'", query)

	searchQuery := `
		SELECT DISTINCT 
			ic.code,
			ic.name_th,
			ic.name_en,
			ic.brand,
			ic.category,
			ic.unit,
			ic.image_url
		FROM ic_inventory ic
		WHERE LOWER(ic.code) = LOWER($1)
		ORDER BY ic.code
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, searchQuery, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute exact code search: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	var icCodes []string

	for rows.Next() {
		var code, nameTh, nameEn, brand, category, unit, imageURL string
		err := rows.Scan(&code, &nameTh, &nameEn, &brand, &category, &unit, &imageURL)
		if err != nil {
			log.Printf("⚠️ Failed to scan row: %v", err)
			continue
		}

		result := map[string]interface{}{
			"code":      code,
			"name_th":   nameTh,
			"name_en":   nameEn,
			"brand":     brand,
			"category":  category,
			"unit":      unit,
			"image_url": imageURL,
		}

		results = append(results, result)
		icCodes = append(icCodes, code)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	// Get total count
	countQuery := `
		SELECT COUNT(DISTINCT ic.code)
		FROM ic_inventory ic
		WHERE LOWER(ic.code) = LOWER($1)
	`

	var totalCount int
	err = s.db.QueryRowContext(ctx, countQuery, query).Scan(&totalCount)
	if err != nil {
		log.Printf("⚠️ Failed to get total count: %v", err)
		totalCount = len(results)
	}

	// Load price data
	if len(icCodes) > 0 {
		s.enrichResultsWithPrice(ctx, results, icCodes)
	}

	log.Printf("✅ [EXACT-CODE-SEARCH] Found %d results for code '%s'", len(results), query)
	return results, totalCount, nil
}
