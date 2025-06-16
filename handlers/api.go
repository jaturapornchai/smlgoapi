package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"smlgoapi/models"
	"smlgoapi/services"

	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	clickHouseService *services.ClickHouseService
	vectorDB          *services.TFIDFVectorDatabase
	imageProxyService *services.ImageProxy
}

func NewAPIHandler(clickHouseService *services.ClickHouseService) *APIHandler {
	vectorDB := services.NewTFIDFVectorDatabase(clickHouseService)
	imageProxyService := services.NewImageProxy()
	return &APIHandler{
		clickHouseService: clickHouseService,
		vectorDB:          vectorDB,
		imageProxyService: imageProxyService,
	}
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Get the health status of the API and database
// @Tags health
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Router /health [get]
func (h *APIHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	version, err := h.clickHouseService.GetVersion(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Error:   "Database connection failed: " + err.Error(),
		})
		return
	}

	response := models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   version,
		Database:  "connected",
	}

	c.JSON(http.StatusOK, response)
}

// GetTables godoc
// @Summary Get all database tables
// @Description Retrieve a list of all tables in the database
// @Tags database
// @Produce json
// @Success 200 {object} models.APIResponse{data=[]models.Table}
// @Router /api/tables [get]
func (h *APIHandler) GetTables(c *gin.Context) {
	ctx := c.Request.Context()

	tables, err := h.clickHouseService.GetTables(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    tables,
		Message: "Tables retrieved successfully",
	})
}

// SearchProducts godoc
// @Summary Search products using vector similarity with JSON body
// @Description Search for products using TF-IDF vector similarity with JSON request body (supports all languages)
// @Tags search
// @Accept json
// @Produce json
// @Param search body models.SearchParameters true "Search parameters"
// @Success 200 {object} models.APIResponse
// @Router /search [post]
func (h *APIHandler) SearchProducts(c *gin.Context) {
	startTime := time.Now()

	// Parse JSON body directly
	var params models.SearchParameters
	if err := c.ShouldBindJSON(&params); err != nil {
		log.Printf("❌ [decode] JSON bind error: %v", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid JSON body: " + err.Error(),
		})
		return
	}

	log.Printf("🔍 [decode] Parsed parameters: query='%s', limit=%d, offset=%d",
		params.Query, params.Limit, params.Offset)

	// Validate query
	if params.Query == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Query field is required in params JSON",
		})
		return
	}

	query := params.Query

	// Set default values
	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	offset := params.Offset
	if offset < 0 {
		offset = 0
	} // Simple logging
	fmt.Printf("\n🔍 Search: '%s' (limit: %d)\n", query, limit)
	ctx := c.Request.Context()

	// Perform vector search
	results, err := h.vectorDB.SearchProducts(ctx, query, limit, offset)
	if err != nil {
		duration := time.Since(startTime).Seconds() * 1000
		fmt.Printf("❌ Error: %v (%.1fms)\n", err, duration)

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Search failed: %s", err.Error()),
		})
		return
	}
	duration := time.Since(startTime).Seconds() * 1000

	// Enhanced search results logging
	fmt.Printf("\n� [search_products] SEARCH RESULTS SUMMARY:\n")
	fmt.Printf("   [search_products] Query: '%s'\n", query)
	fmt.Printf("   [search_products] Total Found: %d records\n", results.TotalCount)
	fmt.Printf("   [search_products] Returned: %d results\n", len(results.Data))
	fmt.Printf("   [search_products] Page: %d (offset: %d, limit: %d)\n", (offset/limit)+1, offset, limit)
	fmt.Printf("   [search_products] Processing Duration: %.1fms\n", duration)
	if len(results.Data) > 0 {
		fmt.Printf("\n📋 [search_products] TOP RESULTS DETAILS:\n")
		maxShow := 10 // Show more results in log
		if len(results.Data) < maxShow {
			maxShow = len(results.Data)
		}
		for i := 0; i < maxShow; i++ {
			result := results.Data[i]

			// Extract metadata directly from fields
			itemCode := result.Code
			if itemCode == "" {
				itemCode = "N/A"
			}

			qty := fmt.Sprintf("%.2f", result.BalanceQty)

			category := "N/A" // Category not available in current struct

			fmt.Printf("     [search_products] %d. [%s] %s\n", i+1, itemCode, result.Name)
			fmt.Printf("         └─ Score: %.4f | Qty: %s | Category: %s\n",
				result.SimilarityScore, qty, category)
		}

		if len(results.Data) > maxShow {
			fmt.Printf("     [search_products] ... and %d more results\n", len(results.Data)-maxShow)
		}
	} else {
		fmt.Printf("   [search_products] ❌ No results found for query: '%s'\n", query)
	}

	fmt.Printf("\n✅ [search_products] SEARCH COMPLETED (%.1fms)\n", duration)
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    results,
		Message: "Search completed successfully",
	})
}

// ImageProxy godoc
// @Summary Proxy images with caching
// @Description Proxy and cache images from external URLs with CORS support
// @Tags proxy
// @Produce json,image/jpeg,image/png,image/gif,image/webp
// @Param url query string true "Image URL to proxy"
// @Success 200 {file} binary "Image file"
// @Router /imgproxy [get]
func (h *APIHandler) ImageProxy(c *gin.Context) {
	h.imageProxyService.ProxyHandler(c)
}

// ImageProxyHead godoc
// @Summary HEAD request for image proxy
// @Description Check if image exists without downloading content
// @Tags proxy
// @Param url query string true "Image URL to check"
// @Success 200 "Image exists"
// @Success 404 "Image not found"
// @Router /imgproxy [head]
func (h *APIHandler) ImageProxyHead(c *gin.Context) {
	h.imageProxyService.HeadHandler(c)
}
