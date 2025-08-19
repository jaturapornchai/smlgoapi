package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"smlgoapi/config"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
)

// Product represents a product in the vector search results
type Product struct {
	Barcode   string  `json:"barcode"`
	Name      string  `json:"name"`
	ICCode    string  `json:"ic_code"`
	Relevance float64 `json:"relevance_percentage"`
}

// WeaviateService handles vector database operations
type WeaviateService struct {
	client *weaviate.Client
}

// EmbeddingRequest and EmbeddingResponse for embedding service
type EmbeddingRequest struct {
	Text string `json:"text"`
}
type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
	Error     string    `json:"error,omitempty"`
}

// NewWeaviateService creates a new Weaviate service
func NewWeaviateService(config *config.Config) (*WeaviateService, error) {
	weaviateURL := config.GetWeaviateURL()
	scheme := config.GetWeaviateScheme()

	if weaviateURL == "" {
		weaviateURL = "localhost:8080"
		scheme = "http"
		log.Printf("⚠️ Weaviate URL not configured, using default: http://localhost:8080")
	}

	cfg := weaviate.Config{
		Host:   weaviateURL,
		Scheme: scheme,
	}

	if weaviateURL != "" && (weaviateURL[:7] == "http://" || weaviateURL[:8] == "https://") {
		if weaviateURL[:7] == "http://" {
			cfg.Host = weaviateURL[7:]
			cfg.Scheme = "http"
		} else if weaviateURL[:8] == "https://" {
			cfg.Host = weaviateURL[8:]
			cfg.Scheme = "https"
		}
	}

	client, err := weaviate.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ready, err := client.Misc().ReadyChecker().Do(ctx)
	if err != nil || !ready {
		log.Printf("⚠️ Weaviate connection test failed: %v", err)
		return nil, fmt.Errorf("weaviate server not reachable at %s://%s", cfg.Scheme, cfg.Host)
	}

	log.Printf("🔗 Connected to Weaviate at: %s://%s", cfg.Scheme, cfg.Host)

	return &WeaviateService{
		client: client,
	}, nil
}

// getEmbeddingVector calls the embedding service to get a vector for the query
func getEmbeddingVector(query string) ([]float32, error) {
	embeddingServiceURL := os.Getenv("EMBEDDING_SERVICE_URL")

	reqBody := EmbeddingRequest{Text: query}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %v", err)
	}
	resp, err := http.Post(embeddingServiceURL+"/embed", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding service: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding service returned status %d", resp.StatusCode)
	}
	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %v", err)
	}
	if embeddingResp.Error != "" {
		return nil, fmt.Errorf("embedding service error: %s", embeddingResp.Error)
	}
	// Convert []float64 to []float32
	vector := make([]float32, len(embeddingResp.Embedding))
	for i, v := range embeddingResp.Embedding {
		vector[i] = float32(v)
	}
	return vector, nil
}

// SearchProducts performs vector search using Weaviate nearVector (embedding)
func (w *WeaviateService) SearchProducts(ctx context.Context, query string, limit int) ([]Product, error) {
	className := os.Getenv("WEAVIATE_COLLECTION")
	if className == "" {
		className = "Product"
	}

	// 1. Get embedding vector from embedding service
	queryVector, err := getEmbeddingVector(query)
	if err != nil {
		log.Printf("Embedding service error: %v", err)
		return nil, err
	}

	// 2. Use nearVector for semantic search
	nearVector := w.client.GraphQL().NearVectorArgBuilder().
		WithVector(queryVector).
		WithDistance(0.7) // similarity threshold, adjust as needed

	result, err := w.client.GraphQL().Get().
		WithClassName(className).
		WithFields(
			graphql.Field{Name: "barcode"},
			graphql.Field{Name: "name"},
			graphql.Field{Name: "icCode"},
			graphql.Field{Name: "_additional", Fields: []graphql.Field{
				{Name: "distance"},
				{Name: "certainty"},
			}},
		).
		WithNearVector(nearVector).
		WithLimit(limit).
		Do(ctx)

	if err != nil {
		log.Printf("Weaviate search error: %v", err)
		return nil, err
	}

	log.Printf("Weaviate GraphQL result received")

	var products []Product

	if result.Data != nil {
		if data, ok := result.Data["Get"].(map[string]interface{}); ok {
			if productList, ok := data[className].([]interface{}); ok {
				for _, item := range productList {
					if product, ok := item.(map[string]interface{}); ok {
						p := Product{}

						if barcode, ok := product["barcode"].(string); ok {
							p.Barcode = barcode
						}
						if name, ok := product["name"].(string); ok {
							p.Name = name
						}
						if icCode, ok := product["icCode"].(string); ok {
							p.ICCode = icCode
						}
						// Convert distance to similarity percentage
						if additional, ok := product["_additional"].(map[string]interface{}); ok {
							if distance, ok := additional["distance"].(float64); ok {
								p.Relevance = (1.0 - distance) * 100.0
								if p.Relevance < 0 {
									p.Relevance = 0
								}
							}
						}
						products = append(products, p)
					}
				}
			}
		}
	}

	log.Printf("Found %d products from Weaviate", len(products))
	return products, nil
}

// GetBarcodes extracts barcodes from search results
func (w *WeaviateService) GetBarcodes(products []Product) []string {
	barcodes := make([]string, len(products))
	for i, product := range products {
		barcodes[i] = product.Barcode
	}
	return barcodes
}

// GetICCodes extracts IC codes from search results
func (w *WeaviateService) GetICCodes(products []Product) []string {
	icCodes := make([]string, 0, len(products))
	for _, product := range products {
		if product.ICCode != "" {
			icCodes = append(icCodes, product.ICCode)
		}
	}
	return icCodes
}

// GetICCodesWithRelevance extracts IC codes and their relevance scores from search results
func (w *WeaviateService) GetICCodesWithRelevance(products []Product) ([]string, map[string]float64) {
	icCodes := make([]string, 0, len(products))
	relevanceMap := make(map[string]float64)

	for _, product := range products {
		if product.ICCode != "" {
			icCodes = append(icCodes, product.ICCode)
			relevanceMap[product.ICCode] = product.Relevance
		}
	}
	return icCodes, relevanceMap
}

// GetBarcodesWithRelevance extracts barcodes and their relevance scores from search results
func (w *WeaviateService) GetBarcodesWithRelevance(products []Product) ([]string, map[string]float64) {
	barcodes := make([]string, len(products))
	relevanceMap := make(map[string]float64)

	for i, product := range products {
		barcodes[i] = product.Barcode
		if product.ICCode != "" {
			relevanceMap[product.ICCode] = product.Relevance
		}
		// Also map by barcode for fallback
		relevanceMap[product.Barcode] = product.Relevance
	}
	return barcodes, relevanceMap
}

// GetICCodeToBarcodeMap creates a mapping from IC codes to barcodes from search results
func (w *WeaviateService) GetICCodeToBarcodeMap(products []Product) map[string]string {
	icCodeToBarcodeMap := make(map[string]string)

	for _, product := range products {
		if product.ICCode != "" && product.Barcode != "" {
			icCodeToBarcodeMap[product.ICCode] = product.Barcode
		}
	}

	return icCodeToBarcodeMap
}

// GetBarcodeToBarcodeMap creates a mapping from barcodes to barcodes (for consistency)
func (w *WeaviateService) GetBarcodeToBarcodeMap(products []Product) map[string]string {
	barcodeToBarcodeMap := make(map[string]string)

	for _, product := range products {
		if product.Barcode != "" {
			barcodeToBarcodeMap[product.Barcode] = product.Barcode
		}
	}

	return barcodeToBarcodeMap
}
