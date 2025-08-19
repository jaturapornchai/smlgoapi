package services

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// PriceInfo holds price information from ic_inventory_price_formula
type PriceInfo struct {
	ICCode string
	Price0 float64 // Parsed from string
	Price1 float64
	Price2 float64
	Price3 float64
	Price4 float64
}

// enrichResultsWithPrice adds price information to search results
func (s *PostgreSQLService) enrichResultsWithPrice(ctx context.Context, results []map[string]interface{}, icCodes []string) {
	if len(icCodes) == 0 {
		return
	}

	// Load price data
	priceMap, err := s.LoadPriceFormulaFiltered(ctx, icCodes)
	if err != nil {
		log.Printf("⚠️ Failed to load price data: %v", err)
		priceMap = make(map[string]*PriceInfo)
	}

	// Enrich results with price data
	for _, result := range results {
		code := result["code"].(string)

		// Add price information
		if priceInfo, exists := priceMap[code]; exists {
			result["price_0"] = priceInfo.Price0
			result["price_1"] = priceInfo.Price1
			result["price_2"] = priceInfo.Price2
			result["price_3"] = priceInfo.Price3
			result["price_4"] = priceInfo.Price4
		} else {
			result["price_0"] = 0.0
			result["price_1"] = 0.0
			result["price_2"] = 0.0
			result["price_3"] = 0.0
			result["price_4"] = 0.0
		}
	}
}

// LoadPriceFormulaFiltered loads price data for specific IC codes
func (s *PostgreSQLService) LoadPriceFormulaFiltered(ctx context.Context, icCodes []string) (map[string]*PriceInfo, error) {
	if len(icCodes) == 0 {
		return make(map[string]*PriceInfo), nil
	}

	// Create placeholders for the IN clause
	placeholders := make([]string, len(icCodes))
	args := make([]interface{}, len(icCodes))
	for i, code := range icCodes {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = code
	}

	query := fmt.Sprintf(`
		SELECT ic_code, price_0, price_1, price_2, price_3, price_4
		FROM ic_inventory_price_formula
		WHERE ic_code IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query price formula: %w", err)
	}
	defer rows.Close()

	priceMap := make(map[string]*PriceInfo)

	for rows.Next() {
		var icCode, price0Str, price1Str, price2Str, price3Str, price4Str string
		err := rows.Scan(&icCode, &price0Str, &price1Str, &price2Str, &price3Str, &price4Str)
		if err != nil {
			log.Printf("⚠️ Failed to scan price row: %v", err)
			continue
		}

		// Parse price strings to float64
		price0, _ := strconv.ParseFloat(price0Str, 64)
		price1, _ := strconv.ParseFloat(price1Str, 64)
		price2, _ := strconv.ParseFloat(price2Str, 64)
		price3, _ := strconv.ParseFloat(price3Str, 64)
		price4, _ := strconv.ParseFloat(price4Str, 64)

		priceMap[icCode] = &PriceInfo{
			ICCode: icCode,
			Price0: price0,
			Price1: price1,
			Price2: price2,
			Price3: price3,
			Price4: price4,
		}
	}

	return priceMap, rows.Err()
}
