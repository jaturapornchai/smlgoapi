package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/go-ego/gse"
	"github.com/kljensen/snowball"
)

type TFIDFVectorDatabase struct {
	clickHouseService *ClickHouseService
	seg               gse.Segmenter
	documents         map[string]*Document
	idf               map[string]float64
	totalDocs         int
}

type Document struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Content  string                 `json:"content"`
	ImgURL   string                 `json:"img_url"`
	Metadata map[string]interface{} `json:"metadata"`
	TF       map[string]float64     `json:"tf"`
}

type SearchResult struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	SimilarityScore float64 `json:"similarity_score"`
	Code            string  `json:"code"`
	BalanceQty      float64 `json:"balance_qty"`
	Price           float64 `json:"price"`
	SupplierCode    string  `json:"supplier_code"`
	Unit            string  `json:"unit"`
	ImgURL          string  `json:"img_url"`
	SearchPriority  int     `json:"search_priority"`

	// New pricing and inventory fields
	SalePrice        float64 `json:"sale_price"`
	PremiumWord      string  `json:"premium_word"`
	DiscountPrice    float64 `json:"discount_price"`
	DiscountPercent  float64 `json:"discount_percent"`
	FinalPrice       float64 `json:"final_price"`
	SoldQty          float64 `json:"sold_qty"`
	MultiPacking     int     `json:"multi_packing"`
	MultiPackingName string  `json:"multi_packing_name"`
	Barcodes         string  `json:"barcodes"`
	QtyAvailable     float64 `json:"qty_available"`
}

type VectorSearchResponse struct {
	Data       []SearchResult `json:"data"`
	TotalCount int            `json:"total_count"`
	Query      string         `json:"query"`
	Duration   float64        `json:"duration_ms"`
}

func NewTFIDFVectorDatabase(clickHouseService *ClickHouseService) *TFIDFVectorDatabase {
	seg, err := gse.New()
	if err != nil {
		// Fallback to default segmenter
		seg = gse.Segmenter{}
	}
	// Load default dictionary
	if err := seg.LoadDict(); err != nil {
		// Log error but continue - dictionary loading is optional
		fmt.Printf("Warning: Failed to load segmenter dictionary: %v\n", err)
	}

	return &TFIDFVectorDatabase{
		clickHouseService: clickHouseService,
		seg:               seg,
		documents:         make(map[string]*Document),
		idf:               make(map[string]float64),
	}
}

func (vdb *TFIDFVectorDatabase) LoadDocuments(ctx context.Context) error { // Query all products from ClickHouse
	query := `
		SELECT code, name
		FROM ic_inventory
		WHERE name != '' AND name IS NOT NULL
	`

	rows, err := vdb.clickHouseService.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	termFreq := make(map[string]map[string]int)
	docCount := make(map[string]int)
	for rows.Next() {
		var code, name string

		if err := rows.Scan(&code, &name); err != nil {
			continue
		}

		// Create document
		content := fmt.Sprintf("%s %s", name, code)
		doc := &Document{
			ID:      code,
			Name:    name,
			ImgURL:  "", // Will be fetched later during search
			Content: content,
			Metadata: map[string]interface{}{
				"code": code,
				// Other fields will be fetched later during search
			},
			TF: make(map[string]float64),
		}

		// Tokenize and calculate term frequency
		tokens := vdb.tokenize(content)
		if len(tokens) == 0 {
			continue
		}

		termCount := make(map[string]int)
		for _, token := range tokens {
			termCount[token]++
		}

		// Calculate TF
		for term, count := range termCount {
			doc.TF[term] = float64(count) / float64(len(tokens))

			// Track for IDF calculation
			if termFreq[term] == nil {
				termFreq[term] = make(map[string]int)
			}
			termFreq[term][code] = count
			docCount[term]++
		}

		vdb.documents[code] = doc
	}

	vdb.totalDocs = len(vdb.documents)

	// Calculate IDF
	for term := range docCount {
		vdb.idf[term] = math.Log(float64(vdb.totalDocs) / float64(docCount[term]))
	}

	return nil
}

func (vdb *TFIDFVectorDatabase) tokenize(text string) []string {
	text = strings.ToLower(text)

	// Remove non-alphanumeric characters except Thai characters
	var cleaned strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			cleaned.WriteRune(r)
		} else {
			cleaned.WriteRune(' ')
		}
	}

	text = cleaned.String()

	// Check if text contains Thai characters
	hasthai := false
	for _, r := range text {
		if r >= 0x0E00 && r <= 0x0E7F {
			hasthai = true
			break
		}
	}

	var tokens []string

	if hasthai {
		// Use GSE for Thai text
		segments := vdb.seg.Segment([]byte(text))
		for _, seg := range segments {
			token := strings.TrimSpace(seg.Token().Text())
			if len(token) > 1 {
				tokens = append(tokens, token)
			}
		}
	} else {
		// Simple whitespace tokenization for English
		words := strings.Fields(text)
		for _, word := range words {
			word = strings.TrimSpace(word)
			if len(word) > 2 {
				// Apply stemming for English words
				stemmed, err := snowball.Stem(word, "english", true)
				if err == nil && len(stemmed) > 1 {
					tokens = append(tokens, stemmed)
				} else {
					tokens = append(tokens, word)
				}
			}
		}
	}

	return tokens
}

func (vdb *TFIDFVectorDatabase) calculateTFIDF(doc *Document) map[string]float64 {
	tfidf := make(map[string]float64)
	for term, tf := range doc.TF {
		if idf, exists := vdb.idf[term]; exists {
			tfidf[term] = tf * idf
		}
	}
	return tfidf
}

func (vdb *TFIDFVectorDatabase) cosineSimilarity(vec1, vec2 map[string]float64) float64 {
	var dotProduct, norm1, norm2 float64

	// Calculate dot product and norms
	for term, val1 := range vec1 {
		if val2, exists := vec2[term]; exists {
			dotProduct += val1 * val2
		}
		norm1 += val1 * val1
	}

	for _, val2 := range vec2 {
		norm2 += val2 * val2
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// fetchAdditionalData queries ic_inventory for additional product information
func (vdb *TFIDFVectorDatabase) fetchAdditionalData(ctx context.Context, productCodes []string) (map[string]string, map[string]map[string]interface{}, error) {
	if len(productCodes) == 0 {
		return make(map[string]string), make(map[string]map[string]interface{}), nil
	}

	// Create placeholders for IN clause
	placeholders := make([]string, len(productCodes))
	args := make([]interface{}, len(productCodes))
	for i, code := range productCodes {
		placeholders[i] = "?"
		args[i] = code
	}

	query := fmt.Sprintf(`
		SELECT code, image_url, unit_standard, balance_qty, supplier_code, 100 as price
		FROM ic_inventory 
		WHERE code IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := vdb.clickHouseService.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch additional data: %w", err)
	}
	defer rows.Close()

	imageMap := make(map[string]string)
	dataMap := make(map[string]map[string]interface{})

	for rows.Next() {
		var code, imageURL, unit, supplierCode string
		var qty, price float64

		if err := rows.Scan(&code, &imageURL, &unit, &qty, &supplierCode, &price); err != nil {
			continue
		}

		// Clean up image URL
		imageURL = strings.TrimSpace(imageURL)
		imageURL = strings.ReplaceAll(imageURL, "[\"", "")
		imageURL = strings.ReplaceAll(imageURL, "\"]", "")
		imageURL = strings.ReplaceAll(imageURL, "[]", "")

		if imageURL != "" && imageURL != "N/A" {
			imageMap[code] = imageURL
		}

		// Store all metadata
		dataMap[code] = map[string]interface{}{
			"code":          code,
			"unit":          unit,
			"balance_qty":   qty,
			"supplier_code": supplierCode,
			"price":         price,
			"img_url":       imageURL,
		}
	}

	return imageMap, dataMap, rows.Err()
}

// SearchProducts performs multi-step search with priority ranking:
// 1. Full text search by code (highest priority)
// 2. Full text search by name (medium priority)
// 3. Vector search (lowest priority)
func (vdb *TFIDFVectorDatabase) SearchProducts(ctx context.Context, query string, limit, offset int) (*VectorSearchResponse, error) {
	startTime := time.Now()

	// Ensure documents are loaded
	if len(vdb.documents) == 0 {
		if err := vdb.LoadDocuments(ctx); err != nil {
			return nil, fmt.Errorf("failed to load documents: %w", err)
		}
	}

	// Step 1: Full text search by code (highest priority)
	codeResults, err := vdb.searchByCode(ctx, query, limit*2)
	if err != nil {
		return nil, fmt.Errorf("failed to search by code: %v", err)
	}

	// Step 2: Full text search by name (medium priority)
	nameResults, err := vdb.searchByName(ctx, query, limit*2)
	if err != nil {
		return nil, fmt.Errorf("failed to search by name: %v", err)
	}

	// Step 3: Vector search (lowest priority)
	vectorResults, err := vdb.performVectorSearch(ctx, query, limit*2)
	if err != nil {
		return nil, fmt.Errorf("failed to perform vector search: %v", err)
	}

	// Combine results with priority and deduplication
	combinedResults := vdb.combineSearchResults(codeResults, nameResults, vectorResults)
	// Fetch additional data for all unique results
	var productCodes []string
	for _, result := range combinedResults {
		productCodes = append(productCodes, result.ID)
	}

	additionalImages, additionalData, dataErr := vdb.fetchAdditionalData(ctx, productCodes)
	if dataErr != nil {
		fmt.Printf("Warning: Failed to fetch additional data: %v\n", dataErr)
	} else {
		// Update results with additional data
		for i, result := range combinedResults {
			if additionalImg, exists := additionalImages[result.ID]; exists && additionalImg != "" {
				combinedResults[i].ImgURL = additionalImg
			}
			if data, exists := additionalData[result.ID]; exists {
				// Update individual fields from additional data
				if balanceQty, ok := data["balance_qty"].(float64); ok {
					combinedResults[i].BalanceQty = balanceQty
				}
				if price, ok := data["price"].(float64); ok {
					combinedResults[i].Price = price
				}
				if supplierCode, ok := data["supplier_code"].(string); ok {
					combinedResults[i].SupplierCode = supplierCode
				}
				if unit, ok := data["unit"].(string); ok {
					combinedResults[i].Unit = unit
				}
			}
		}
	}

	// Sort by priority and relevance
	vdb.sortResultsByPriority(combinedResults)

	totalCount := len(combinedResults)

	// Apply pagination
	if offset >= len(combinedResults) {
		combinedResults = []SearchResult{}
	} else {
		end := offset + limit
		if end > len(combinedResults) {
			end = len(combinedResults)
		}
		combinedResults = combinedResults[offset:end]
	}

	duration := float64(time.Since(startTime).Nanoseconds()) / 1e6

	return &VectorSearchResponse{
		Data:       combinedResults,
		TotalCount: totalCount,
		Query:      query,
		Duration:   duration,
	}, nil
}

// searchByCode performs full text search on product codes
func (vdb *TFIDFVectorDatabase) searchByCode(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	var results []SearchResult
	queryLower := strings.ToLower(query)

	for _, doc := range vdb.documents { // Check if document ID (product code) contains the query
		if strings.Contains(strings.ToLower(doc.ID), queryLower) {
			imgURL := ""
			if url, exists := doc.Metadata["img_url"]; exists {
				if urlStr, ok := url.(string); ok {
					imgURL = urlStr
				}
			}

			result := SearchResult{
				ID:              doc.ID,
				Name:            doc.Name,
				Code:            doc.ID,
				SimilarityScore: 1.0, // High score for exact code matches
				ImgURL:          imgURL,
				SearchPriority:  1,
				// Default values - will be updated by fetchAdditionalData
				BalanceQty:   0,
				Price:        0,
				SupplierCode: "",
				Unit:         "",
			}

			results = append(results, result)
		}
	}

	// Sort by code relevance (exact matches first, then partial matches)
	sort.Slice(results, func(i, j int) bool {
		iExact := strings.EqualFold(results[i].ID, query)
		jExact := strings.EqualFold(results[j].ID, query)
		if iExact != jExact {
			return iExact // Exact matches first
		}
		return results[i].ID < results[j].ID // Alphabetical order for partial matches
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// searchByName performs full text search on product names
func (vdb *TFIDFVectorDatabase) searchByName(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	var results []SearchResult
	queryLower := strings.ToLower(query)

	for _, doc := range vdb.documents { // Check if document name contains the query
		if strings.Contains(strings.ToLower(doc.Name), queryLower) {
			imgURL := ""
			if url, exists := doc.Metadata["img_url"]; exists {
				if urlStr, ok := url.(string); ok {
					imgURL = urlStr
				}
			}

			result := SearchResult{
				ID:              doc.ID,
				Name:            doc.Name,
				Code:            doc.ID,
				SimilarityScore: 0.8, // Medium score for name matches
				ImgURL:          imgURL,
				SearchPriority:  2,
				// Default values - will be updated by fetchAdditionalData
				BalanceQty:   0,
				Price:        0,
				SupplierCode: "",
				Unit:         "",
			}

			results = append(results, result)
		}
	}

	// Sort by name relevance
	sort.Slice(results, func(i, j int) bool {
		iExact := strings.EqualFold(results[i].Name, query)
		jExact := strings.EqualFold(results[j].Name, query)
		if iExact != jExact {
			return iExact // Exact matches first
		}
		return results[i].Name < results[j].Name // Alphabetical order for partial matches
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// performVectorSearch performs the original TF-IDF vector search
func (vdb *TFIDFVectorDatabase) performVectorSearch(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Tokenize query
	queryTokens := vdb.tokenize(query)
	if len(queryTokens) == 0 {
		return []SearchResult{}, nil
	}

	// Calculate query TF-IDF
	queryTF := make(map[string]float64)
	for _, token := range queryTokens {
		queryTF[token]++
	}
	for token := range queryTF {
		queryTF[token] /= float64(len(queryTokens))
	}

	queryTFIDF := make(map[string]float64)
	for term, tf := range queryTF {
		if idf, exists := vdb.idf[term]; exists {
			queryTFIDF[term] = tf * idf
		}
	}

	// Calculate similarity for all documents
	var results []SearchResult
	for _, doc := range vdb.documents {
		docTFIDF := vdb.calculateTFIDF(doc)
		similarity := vdb.cosineSimilarity(queryTFIDF, docTFIDF)
		if similarity > 0.01 { // Only keep results with reasonable similarity
			imgURL := ""
			if url, exists := doc.Metadata["img_url"]; exists {
				if urlStr, ok := url.(string); ok {
					imgURL = urlStr
				}
			}

			result := SearchResult{
				ID:              doc.ID,
				Name:            doc.Name,
				Code:            doc.ID,
				SimilarityScore: similarity,
				ImgURL:          imgURL,
				SearchPriority:  3,
				// Default values - will be updated by fetchAdditionalData
				BalanceQty:   0,
				Price:        0,
				SupplierCode: "",
				Unit:         "",
			}

			results = append(results, result)
		}
	}

	// Sort by similarity score
	sort.Slice(results, func(i, j int) bool {
		return results[i].SimilarityScore > results[j].SimilarityScore
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// combineSearchResults combines and deduplicates results from different search methods
func (vdb *TFIDFVectorDatabase) combineSearchResults(codeResults, nameResults, vectorResults []SearchResult) []SearchResult {
	resultMap := make(map[string]SearchResult) // Use map to avoid duplicates

	// Add code results (highest priority)
	for _, result := range codeResults {
		resultMap[result.ID] = result
	}

	// Add name results (medium priority) - only if not already found
	for _, result := range nameResults {
		if _, exists := resultMap[result.ID]; !exists {
			resultMap[result.ID] = result
		}
	}

	// Add vector results (lowest priority) - only if not already found
	for _, result := range vectorResults {
		if _, exists := resultMap[result.ID]; !exists {
			resultMap[result.ID] = result
		}
	}

	// Convert map back to slice
	var combined []SearchResult
	for _, result := range resultMap {
		combined = append(combined, result)
	}

	return combined
}

// sortResultsByPriority sorts results by search priority and relevance
func (vdb *TFIDFVectorDatabase) sortResultsByPriority(results []SearchResult) {
	sort.Slice(results, func(i, j int) bool {
		// Get search priorities directly from the struct field
		iPriority := results[i].SearchPriority
		jPriority := results[j].SearchPriority

		// Sort by priority first (lower number = higher priority)
		if iPriority != jPriority {
			return iPriority < jPriority
		}

		// Within same priority, sort by similarity score and image availability
		scoreDiff := math.Abs(results[i].SimilarityScore - results[j].SimilarityScore)
		if scoreDiff < 0.1 { // If scores are close
			// Prioritize products with images
			iHasImage := results[i].ImgURL != "" && results[i].ImgURL != "N/A"
			jHasImage := results[j].ImgURL != "" && results[j].ImgURL != "N/A"
			if iHasImage != jHasImage {
				return iHasImage
			}
		}

		// Sort by similarity score
		return results[i].SimilarityScore > results[j].SimilarityScore
	})
}
