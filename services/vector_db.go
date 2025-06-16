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
	Metadata map[string]interface{} `json:"metadata"`
	TF       map[string]float64     `json:"tf"`
}

type SearchResult struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	SimilarityScore float64                `json:"similarity_score"`
	Metadata        map[string]interface{} `json:"metadata"`
	ImgURL          string                 `json:"img_url"`
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
	seg.LoadDict()

	return &TFIDFVectorDatabase{
		clickHouseService: clickHouseService,
		seg:               seg,
		documents:         make(map[string]*Document),
		idf:               make(map[string]float64),
	}
}

func (vdb *TFIDFVectorDatabase) LoadDocuments(ctx context.Context) error {
	// Query all products from ClickHouse
	query := `
		SELECT code, name, unit_standard, balance_qty, supplier_code,100 as price,'https://f.ptcdn.info/468/065/000/pw5l8933TR0cL0CH7f-o.jpg' as img_url
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
		var code, name, unit, supplierCode string
		var qty float64
		var price float64
		var imgURL string

		if err := rows.Scan(&code, &name, &unit, &qty, &supplierCode, &price, &imgURL); err != nil {
			continue
		}

		// Create document
		content := fmt.Sprintf("%s %s %s %f %s", name, code, unit, price, imgURL)
		doc := &Document{
			ID:      code,
			Name:    name,
			Content: content,
			Metadata: map[string]interface{}{
				"code":          code,
				"unit":          unit,
				"balance_qty":   qty,
				"supplier_code": supplierCode,
				"price":         price,
				"img_url":       imgURL,
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
	for term, _ := range docCount {
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

func (vdb *TFIDFVectorDatabase) SearchProducts(ctx context.Context, query string, limit, offset int) (*VectorSearchResponse, error) {
	startTime := time.Now()

	// Ensure documents are loaded
	if len(vdb.documents) == 0 {
		if err := vdb.LoadDocuments(ctx); err != nil {
			return nil, fmt.Errorf("failed to load documents: %w", err)
		}
	}

	// Tokenize query
	queryTokens := vdb.tokenize(query)
	if len(queryTokens) == 0 {
		return &VectorSearchResponse{
			Data:       []SearchResult{},
			TotalCount: 0,
			Query:      query,
			Duration:   float64(time.Since(startTime).Nanoseconds()) / 1e6,
		}, nil
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

	// Calculate similarities
	var results []SearchResult
	for _, doc := range vdb.documents {
		docTFIDF := vdb.calculateTFIDF(doc)
		similarity := vdb.cosineSimilarity(queryTFIDF, docTFIDF)
		if similarity > 0 {
			imgURL := ""
			if url, exists := doc.Metadata["img_url"]; exists {
				if urlStr, ok := url.(string); ok {
					imgURL = urlStr
				}
			}

			result := SearchResult{
				ID:              doc.ID,
				Name:            doc.Name,
				SimilarityScore: similarity,
				Metadata:        doc.Metadata,
				ImgURL:          imgURL,
			}
			results = append(results, result)
		}
	}

	// Sort by similarity score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].SimilarityScore > results[j].SimilarityScore
	})

	totalCount := len(results)

	// Apply pagination
	if offset >= len(results) {
		results = []SearchResult{}
	} else {
		end := offset + limit
		if end > len(results) {
			end = len(results)
		}
		results = results[offset:end]
	}

	duration := float64(time.Since(startTime).Nanoseconds()) / 1e6

	return &VectorSearchResponse{
		Data:       results,
		TotalCount: totalCount,
		Query:      query,
		Duration:   duration,
	}, nil
}
