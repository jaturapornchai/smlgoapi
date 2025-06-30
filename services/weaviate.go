package services

import (
	"context"
	"fmt"
	"log"
	"strconv"
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

// NewWeaviateService creates a new Weaviate service
func NewWeaviateService(config *config.Config) (*WeaviateService, error) {
	weaviateURL := config.GetWeaviateURL()
	scheme := config.GetWeaviateScheme()

	// If URL is empty, use default localhost
	if weaviateURL == "" {
		weaviateURL = "localhost:8080"
		scheme = "http"
		log.Printf("⚠️ Weaviate URL not configured, using default: http://localhost:8080")
	}

	cfg := weaviate.Config{
		Host:   weaviateURL,
		Scheme: scheme,
	}

	// Handle full URL format by extracting host part
	if weaviateURL != "" && (weaviateURL[:7] == "http://" || weaviateURL[:8] == "https://") {
		// If URL contains protocol, extract just the host:port part
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

	// Test connection by checking if Weaviate is ready
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ready, err := client.Misc().ReadyChecker().Do(ctx)
	if err != nil || !ready {
		log.Printf("⚠️ Weaviate connection test failed: %v", err)
		return nil, fmt.Errorf("Weaviate server not reachable at %s://%s", cfg.Scheme, cfg.Host)
	}

	log.Printf("🔗 Connected to Weaviate at: %s://%s", cfg.Scheme, cfg.Host)

	return &WeaviateService{
		client: client,
	}, nil
}

// SearchProducts performs vector search using Weaviate BM25
func (w *WeaviateService) SearchProducts(ctx context.Context, query string, limit int) ([]Product, error) {
	className := "Product"

	// Use BM25 search since vectorizer is "none"
	bm25 := w.client.GraphQL().Bm25ArgBuilder().
		WithQuery(query)

	result, err := w.client.GraphQL().Get().
		WithClassName(className).
		WithFields(
			graphql.Field{Name: "barcode"},
			graphql.Field{Name: "name"},
			graphql.Field{Name: "icCode"}, // แก้จาก ic_code เป็น icCode ตาม Weaviate schema
			graphql.Field{Name: "_additional", Fields: []graphql.Field{
				{Name: "score"},
			}},
		).
		WithBM25(bm25).
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

						if icCode, ok := product["icCode"].(string); ok { // แก้จาก ic_code เป็น icCode
							p.ICCode = icCode
						}

						// Calculate relevance percentage from BM25 score
						if additional, ok := product["_additional"].(map[string]interface{}); ok {
							var score float64
							var scoreOk bool

							// Handle different numeric types for score
							switch v := additional["score"].(type) {
							case float64:
								score = v
								scoreOk = true
							case float32:
								score = float64(v)
								scoreOk = true
							case int:
								score = float64(v)
								scoreOk = true
							case string:
								if parsed, err := strconv.ParseFloat(v, 64); err == nil {
									score = parsed
									scoreOk = true
								}
							}

							if scoreOk {
								// BM25 score can be any positive number, convert to percentage
								// Scale to more reasonable percentages
								p.Relevance = score * 10.0
								if p.Relevance > 100.0 {
									p.Relevance = 100.0
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
