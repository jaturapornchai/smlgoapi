package services

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// SearchProductsByBarcodesWithRelevanceAndBarcodeMap searches products by multiple barcodes with relevance and barcode mapping
// This is specifically used for vector search integration with Weaviate
func (s *PostgreSQLService) SearchProductsByBarcodesWithRelevanceAndBarcodeMap(ctx context.Context, barcodes []string, relevanceMap map[string]float64, barcodeMap map[string]string, limit, offset int) ([]map[string]interface{}, int, error) {
	if len(barcodes) == 0 {
		return []map[string]interface{}{}, 0, nil
	}

	log.Printf("🔍 [VECTOR-SEARCH] Starting database search for %d barcodes with relevance and mapping", len(barcodes))

	// Create placeholders for barcodes
	placeholders := make([]string, len(barcodes))
	args := make([]interface{}, len(barcodes))
	for i, barcode := range barcodes {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = strings.ToLower(barcode)
	}

	searchQuery := fmt.Sprintf(`
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
		WHERE LOWER(ib.barcode) IN (%s)
		ORDER BY ic.code
	`, strings.Join(placeholders, ","))

	rows, err := s.db.QueryContext(ctx, searchQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute vector search database query: %w", err)
	}
	defer rows.Close()

	var allResults []map[string]interface{}
	var icCodes []string

	for rows.Next() {
		var code, nameTh, nameEn, brand, category, unit, imageURL, barcode string
		err := rows.Scan(&code, &nameTh, &nameEn, &brand, &category, &unit, &imageURL, &barcode)
		if err != nil {
			log.Printf("⚠️ Failed to scan vector search row: %v", err)
			continue
		}

		// Get relevance score from map (default to 1.0 if not found)
		relevanceScore := 1.0
		if score, exists := relevanceMap[barcode]; exists {
			relevanceScore = score
		}

		// Get mapped barcode if available
		mappedBarcode := barcode
		if mapped, exists := barcodeMap[barcode]; exists {
			mappedBarcode = mapped
		}

		result := map[string]interface{}{
			"code":             code,
			"name_th":          nameTh,
			"name_en":          nameEn,
			"brand":            brand,
			"category":         category,
			"unit":             unit,
			"image_url":        imageURL,
			"barcode":          barcode,
			"mapped_barcode":   mappedBarcode,
			"similarity_score": relevanceScore,
		}

		allResults = append(allResults, result)
		icCodes = append(icCodes, code)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("vector search rows iteration error: %w", err)
	}

	// Sort by relevance score (highest first)
	for i := 0; i < len(allResults)-1; i++ {
		for j := i + 1; j < len(allResults); j++ {
			scoreI := allResults[i]["similarity_score"].(float64)
			scoreJ := allResults[j]["similarity_score"].(float64)
			if scoreJ > scoreI {
				allResults[i], allResults[j] = allResults[j], allResults[i]
			}
		}
	}

	totalCount := len(allResults)

	// Apply pagination
	start := offset
	end := offset + limit
	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	results := allResults[start:end]

	// Load price data for the paginated results
	if len(results) > 0 {
		paginatedICCodes := make([]string, len(results))
		for i, result := range results {
			paginatedICCodes[i] = result["code"].(string)
		}
		s.enrichResultsWithPrice(ctx, results, paginatedICCodes)
	}

	log.Printf("✅ [VECTOR-SEARCH] Found %d total results, returning %d results with relevance scores", totalCount, len(results))
	return results, totalCount, nil
}

// SearchProductsByICCodesWithRelevance searches products by IC codes with relevance mapping
// Alternative method for vector search when using IC codes instead of barcodes
func (s *PostgreSQLService) SearchProductsByICCodesWithRelevance(ctx context.Context, icCodes []string, relevanceMap map[string]float64, limit, offset int) ([]map[string]interface{}, int, error) {
	if len(icCodes) == 0 {
		return []map[string]interface{}{}, 0, nil
	}

	log.Printf("🔍 [VECTOR-SEARCH-IC] Starting database search for %d IC codes with relevance", len(icCodes))

	// Create placeholders for IC codes
	placeholders := make([]string, len(icCodes))
	args := make([]interface{}, len(icCodes))
	for i, code := range icCodes {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = code
	}

	searchQuery := fmt.Sprintf(`
		SELECT DISTINCT 
			ic.code,
			ic.name_th,
			ic.name_en,
			ic.brand,
			ic.category,
			ic.unit,
			ic.image_url
		FROM ic_inventory ic
		WHERE ic.code IN (%s)
		ORDER BY ic.code
	`, strings.Join(placeholders, ","))

	rows, err := s.db.QueryContext(ctx, searchQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute vector search by IC codes: %w", err)
	}
	defer rows.Close()

	var allResults []map[string]interface{}
	var foundICCodes []string

	for rows.Next() {
		var code, nameTh, nameEn, brand, category, unit, imageURL string
		err := rows.Scan(&code, &nameTh, &nameEn, &brand, &category, &unit, &imageURL)
		if err != nil {
			log.Printf("⚠️ Failed to scan vector search IC row: %v", err)
			continue
		}

		// Get relevance score from map (default to 1.0 if not found)
		relevanceScore := 1.0
		if score, exists := relevanceMap[code]; exists {
			relevanceScore = score
		}

		result := map[string]interface{}{
			"code":             code,
			"name_th":          nameTh,
			"name_en":          nameEn,
			"brand":            brand,
			"category":         category,
			"unit":             unit,
			"image_url":        imageURL,
			"similarity_score": relevanceScore,
		}

		allResults = append(allResults, result)
		foundICCodes = append(foundICCodes, code)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("vector search IC rows iteration error: %w", err)
	}

	// Sort by relevance score (highest first)
	for i := 0; i < len(allResults)-1; i++ {
		for j := i + 1; j < len(allResults); j++ {
			scoreI := allResults[i]["similarity_score"].(float64)
			scoreJ := allResults[j]["similarity_score"].(float64)
			if scoreJ > scoreI {
				allResults[i], allResults[j] = allResults[j], allResults[i]
			}
		}
	}

	totalCount := len(allResults)

	// Apply pagination
	start := offset
	end := offset + limit
	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	results := allResults[start:end]

	// Load price data for the paginated results
	if len(results) > 0 {
		paginatedICCodes := make([]string, len(results))
		for i, result := range results {
			paginatedICCodes[i] = result["code"].(string)
		}
		s.enrichResultsWithPrice(ctx, results, paginatedICCodes)
	}

	log.Printf("✅ [VECTOR-SEARCH-IC] Found %d total results, returning %d results with relevance scores", totalCount, len(results))
	return results, totalCount, nil
}
