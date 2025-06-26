package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"smlgoapi/config"
	"smlgoapi/models"
	"smlgoapi/services"

	"github.com/gin-gonic/gin"
)

// DeepSeek API structures
type DeepSeekRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DeepSeekResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

const (
	DeepSeekAPIURL = "https://api.deepseek.com/v1/chat/completions"
	DeepSeekAPIKey = "sk-f57e7b56ab3f4f1a8030b7ae57500b85"
	DeepSeekModel  = "deepseek-chat"
)

type APIHandler struct {
	clickHouseService *services.ClickHouseService
	postgreSQLService *services.PostgreSQLService
	vectorDB          *services.TFIDFVectorDatabase
	imageProxyService *services.ImageProxy
	thaiAdminService  *services.ThaiAdminService
	weaviateService   *services.WeaviateService
}

func NewAPIHandler(clickHouseService *services.ClickHouseService, postgreSQLService *services.PostgreSQLService) *APIHandler {
	var vectorDB *services.TFIDFVectorDatabase
	if clickHouseService != nil {
		vectorDB = services.NewTFIDFVectorDatabase(clickHouseService)
	}
	imageProxyService := services.NewImageProxy()
	thaiAdminService := services.NewThaiAdminService()

	// Initialize Weaviate service with config
	var weaviateService *services.WeaviateService
	cfg := config.LoadConfig()
	ws, err := services.NewWeaviateService(cfg)
	if err != nil {
		log.Printf("⚠️ Failed to initialize Weaviate service: %v", err)
		weaviateService = nil
	} else {
		weaviateService = ws
	}

	return &APIHandler{
		clickHouseService: clickHouseService,
		postgreSQLService: postgreSQLService,
		vectorDB:          vectorDB,
		imageProxyService: imageProxyService,
		thaiAdminService:  thaiAdminService,
		weaviateService:   weaviateService,
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

	var version string
	var err error

	if h.clickHouseService != nil {
		version, err = h.clickHouseService.GetVersion(ctx)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, models.APIResponse{
				Success: false,
				Error:   "ClickHouse connection failed: " + err.Error(),
			})
			return
		}
	} else {
		version = "ClickHouse unavailable"
	}

	// Test PostgreSQL connection
	var pgVersion string
	if h.postgreSQLService != nil {
		pgVersion, err = h.postgreSQLService.GetVersion(ctx)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, models.APIResponse{
				Success: false,
				Error:   "PostgreSQL connection failed: " + err.Error(),
			})
			return
		}
	} else {
		pgVersion = "PostgreSQL unavailable"
	}

	response := models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   fmt.Sprintf("ClickHouse: %s, PostgreSQL: %s", version, pgVersion),
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
// @Summary Search products using vector similarity with JSON body or URL parameters
// @Description Search for products using TF-IDF vector similarity (supports both JSON body and URL parameters for all languages). AI parameter: 0=no AI enhancement, 1=use AI to enhance query
// @Tags search
// @Accept json
// @Produce json
// @Param search body models.SearchParameters true "Search parameters (for POST requests)"
// @Param q query string false "Search query (for GET requests)"
// @Param limit query integer false "Maximum number of results (for GET requests)"
// @Param offset query integer false "Offset for pagination (for GET requests)"
// @Param ai query integer false "AI mode: 0=no AI enhancement, 1=use AI to enhance query (for GET requests)"
// @Success 200 {object} models.APIResponse
// @Router /search [post]
// @Router /search [get]
func (h *APIHandler) SearchProducts(c *gin.Context) {
	startTime := time.Now()

	var params models.SearchParameters

	// Check if this is a GET request with query parameters
	if c.Request.Method == "GET" {
		// Parse URL query parameters
		params.Query = c.Query("q")
		if limit := c.Query("limit"); limit != "" {
			if l, err := strconv.Atoi(limit); err == nil {
				params.Limit = l
			}
		}
		if offset := c.Query("offset"); offset != "" {
			if o, err := strconv.Atoi(offset); err == nil {
				params.Offset = o
			}
		}
		if ai := c.Query("ai"); ai != "" {
			if a, err := strconv.Atoi(ai); err == nil {
				params.AI = a
			}
		}
	} else {
		// Parse JSON body for POST requests
		if err := c.ShouldBindJSON(&params); err != nil {
			log.Printf("❌ [decode] JSON bind error: %v", err)
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error:   "Invalid JSON body: " + err.Error(),
			})
			return
		}
	}

	log.Printf("🔍 [decode] Parsed parameters: query='%s', limit=%d, offset=%d, ai=%d",
		params.Query, params.Limit, params.Offset, params.AI)

	// Validate query
	if params.Query == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Query field is required (use 'q' parameter for GET requests or 'query' field in JSON body for POST requests)",
		})
		return
	}

	query := params.Query

	// AI Enhancement Processing
	if params.AI == 1 {
		log.Printf("🤖 [ai] AI mode enabled, enhancing query: '%s'", query)
		enhancedQuery := h.enhanceQueryWithAI(query)
		if enhancedQuery != query {
			log.Printf("🤖 [ai] Query enhanced from '%s' to '%s'", query, enhancedQuery)
			query = enhancedQuery
		} else {
			log.Printf("🤖 [ai] Query unchanged after AI processing")
		}
	} else {
		log.Printf("💭 [ai] AI mode disabled (ai=0), using original query")
	}

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
	fmt.Printf("\n🔍 Search: '%s' (limit: %d, ai: %d)\n", query, limit, params.AI)
	ctx := c.Request.Context()

	// Perform PostgreSQL search instead of vector search
	searchResults, totalCount, err := h.postgreSQLService.SearchProducts(ctx, query, limit, offset)
	if err != nil {
		duration := time.Since(startTime).Seconds() * 1000
		fmt.Printf("❌ Error: %v (%.1fms)\n", err, duration)

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Search failed: %s", err.Error()),
		})
		return
	}

	// Convert PostgreSQL results to the expected format
	var convertedResults []services.SearchResult
	for _, result := range searchResults {
		searchResult := services.SearchResult{
			ID:              getStringValue(result, "id"),
			Name:            getStringValue(result, "name"),
			Code:            getStringValue(result, "code"),
			Price:           getFloat64Value(result, "price"),
			BalanceQty:      getFloat64Value(result, "balance_qty"),
			Unit:            getStringValue(result, "unit"),
			SupplierCode:    getStringValue(result, "supplier_code"),
			ImgURL:          getStringValue(result, "img_url"),
			SimilarityScore: getFloat64Value(result, "similarity_score"),
			SearchPriority:  int(getFloat64Value(result, "search_priority")),

			// New fields
			SalePrice:        getFloat64Value(result, "sale_price"),
			PremiumWord:      getStringValue(result, "premium_word"),
			DiscountPrice:    getFloat64Value(result, "discount_price"),
			DiscountPercent:  getFloat64Value(result, "discount_percent"),
			FinalPrice:       getFloat64Value(result, "final_price"),
			SoldQty:          getFloat64Value(result, "sold_qty"),
			MultiPacking:     int(getFloat64Value(result, "multi_packing")),
			MultiPackingName: getStringValue(result, "multi_packing_name"),
			Barcodes:         getStringValue(result, "barcodes"),
			QtyAvailable:     getFloat64Value(result, "qty_available"),
		}
		convertedResults = append(convertedResults, searchResult)
	}

	// Create response in the expected format
	results := &services.VectorSearchResponse{
		Data:       convertedResults,
		TotalCount: totalCount,
		Query:      query,
		Duration:   time.Since(startTime).Seconds() * 1000,
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

// CommandEndpoint godoc
// @Summary Execute database command
// @Description Execute any SQL command (INSERT, UPDATE, DELETE, CREATE, etc.) via JSON request
// @Tags database
// @Accept json
// @Produce json
// @Param command body models.CommandRequest true "Command to execute"
// @Success 200 {object} models.CommandResponse
// @Router /command [post]
func (h *APIHandler) CommandEndpoint(c *gin.Context) {
	startTime := time.Now()

	// Parse JSON request
	var commandReq models.CommandRequest
	if err := c.ShouldBindJSON(&commandReq); err != nil {
		log.Printf("❌ [command] JSON bind error: %v", err)
		c.JSON(http.StatusBadRequest, models.CommandResponse{
			Success: false,
			Error:   "Invalid JSON body: " + err.Error(),
		})
		return
	}

	log.Printf("💻 [command] Executing command: %s", commandReq.Query)

	ctx := c.Request.Context()

	// Execute command using ClickHouse service
	result, err := h.clickHouseService.ExecuteCommand(ctx, commandReq.Query)
	duration := float64(time.Since(startTime).Nanoseconds()) / 1e6

	if err != nil {
		log.Printf("❌ [command] Execution failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.CommandResponse{
			Success:  false,
			Error:    fmt.Sprintf("Command execution failed: %s", err.Error()),
			Command:  commandReq.Query,
			Duration: duration,
		})
		return
	}

	log.Printf("✅ [command] Execution successful in %.2fms", duration)

	c.JSON(http.StatusOK, models.CommandResponse{
		Success:  true,
		Message:  "Command executed successfully",
		Result:   result,
		Command:  commandReq.Query,
		Duration: duration,
	})
}

// SelectEndpoint godoc
// @Summary Execute SELECT query
// @Description Execute SELECT query and return data via JSON request
// @Tags database
// @Accept json
// @Produce json
// @Param select body models.SelectRequest true "SELECT query to execute"
// @Success 200 {object} models.SelectResponse
// @Router /select [post]
func (h *APIHandler) SelectEndpoint(c *gin.Context) {
	startTime := time.Now()

	// Parse JSON request
	var selectReq models.SelectRequest
	if err := c.ShouldBindJSON(&selectReq); err != nil {
		log.Printf("❌ [select] JSON bind error: %v", err)
		c.JSON(http.StatusBadRequest, models.SelectResponse{
			Success: false,
			Error:   "Invalid JSON body: " + err.Error(),
		})
		return
	}

	log.Printf("🔍 [select] Executing query: %s", selectReq.Query)

	ctx := c.Request.Context()

	// Execute select query using ClickHouse service
	data, err := h.clickHouseService.ExecuteSelect(ctx, selectReq.Query)
	duration := float64(time.Since(startTime).Nanoseconds()) / 1e6

	if err != nil {
		log.Printf("❌ [select] Query failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.SelectResponse{
			Success:  false,
			Error:    fmt.Sprintf("Query execution failed: %s", err.Error()),
			Query:    selectReq.Query,
			Duration: duration,
		})
		return
	}

	rowCount := len(data)
	log.Printf("✅ [select] Query successful: %d rows returned in %.2fms", rowCount, duration)
	c.JSON(http.StatusOK, models.SelectResponse{
		Success:  true,
		Message:  fmt.Sprintf("Query executed successfully, %d rows returned", rowCount),
		Data:     data,
		Query:    selectReq.Query,
		RowCount: rowCount,
		Duration: duration,
	})
}

// PgCommandEndpoint godoc
// @Summary Execute PostgreSQL database command
// @Description Execute any PostgreSQL SQL command (INSERT, UPDATE, DELETE, CREATE, etc.) via JSON request
// @Tags database
// @Accept json
// @Produce json
// @Param command body models.CommandRequest true "Command to execute"
// @Success 200 {object} models.CommandResponse
// @Router /pgcommand [post]
func (h *APIHandler) PgCommandEndpoint(c *gin.Context) {
	startTime := time.Now()

	// Parse JSON request
	var commandReq models.CommandRequest
	if err := c.ShouldBindJSON(&commandReq); err != nil {
		log.Printf("❌ [pgcommand] JSON bind error: %v", err)
		c.JSON(http.StatusBadRequest, models.CommandResponse{
			Success: false,
			Error:   "Invalid JSON body: " + err.Error(),
		})
		return
	}

	log.Printf("🐘 [pgcommand] Executing PostgreSQL command: %s", commandReq.Query)

	ctx := c.Request.Context()

	// Execute command using PostgreSQL service
	result, err := h.postgreSQLService.ExecuteCommand(ctx, commandReq.Query)
	duration := float64(time.Since(startTime).Nanoseconds()) / 1e6

	if err != nil {
		log.Printf("❌ [pgcommand] Execution failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.CommandResponse{
			Success:  false,
			Error:    fmt.Sprintf("PostgreSQL command execution failed: %s", err.Error()),
			Command:  commandReq.Query,
			Duration: duration,
		})
		return
	}

	log.Printf("✅ [pgcommand] Execution successful in %.2fms", duration)

	c.JSON(http.StatusOK, models.CommandResponse{
		Success:  true,
		Message:  "PostgreSQL command executed successfully",
		Result:   result,
		Command:  commandReq.Query,
		Duration: duration,
	})
}

// PgSelectEndpoint godoc
// @Summary Execute PostgreSQL SELECT query
// @Description Execute PostgreSQL SELECT query and return data via JSON request
// @Tags database
// @Accept json
// @Produce json
// @Param select body models.SelectRequest true "SELECT query to execute"
// @Success 200 {object} models.SelectResponse
// @Router /pgselect [post]
func (h *APIHandler) PgSelectEndpoint(c *gin.Context) {
	startTime := time.Now()

	// Parse JSON request
	var selectReq models.SelectRequest
	if err := c.ShouldBindJSON(&selectReq); err != nil {
		log.Printf("❌ [pgselect] JSON bind error: %v", err)
		c.JSON(http.StatusBadRequest, models.SelectResponse{
			Success: false,
			Error:   "Invalid JSON body: " + err.Error(),
		})
		return
	}

	log.Printf("🐘 [pgselect] Executing PostgreSQL query: %s", selectReq.Query)

	ctx := c.Request.Context()

	// Execute select query using PostgreSQL service
	data, err := h.postgreSQLService.ExecuteSelect(ctx, selectReq.Query)
	duration := float64(time.Since(startTime).Nanoseconds()) / 1e6

	if err != nil {
		log.Printf("❌ [pgselect] Query failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.SelectResponse{
			Success:  false,
			Error:    fmt.Sprintf("PostgreSQL query execution failed: %s", err.Error()),
			Query:    selectReq.Query,
			Duration: duration,
		})
		return
	}

	rowCount := len(data)
	log.Printf("✅ [pgselect] Query successful: %d rows returned in %.2fms", rowCount, duration)

	c.JSON(http.StatusOK, models.SelectResponse{
		Success:  true,
		Message:  fmt.Sprintf("PostgreSQL query executed successfully, %d rows returned", rowCount),
		Data:     data,
		Query:    selectReq.Query,
		RowCount: rowCount,
		Duration: duration,
	})
}

// GuideEndpoint godoc
// @Summary API Guide for AI Agents
// @Description Complete API documentation and usage guide for AI agents and developers
// @Tags documentation
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /guide [get]
func (h *APIHandler) GuideEndpoint(c *gin.Context) {
	guide := map[string]interface{}{
		"api_name":      "SMLGOAPI",
		"version":       "2.0.0",
		"description":   "Advanced Auto Parts API with AI-powered search, multi-language support, and PostgreSQL backend",
		"base_url":      "http://localhost:8008/v1",
		"documentation": "Complete API guide for AI agents, developers, and frontend applications",
		"last_updated":  "2025-06-20",

		"concepts": map[string]interface{}{
			"overview": "SMLGOAPI is a modern REST API providing intelligent auto parts search with AI translation assistance, multi-language support (Thai/English), and comprehensive database operations.",
			"core_features": []string{
				"🤖 AI-powered query enhancement with DeepSeek API integration",
				"🌐 Multi-language search support (Thai ↔ English translation)",
				"🔍 Full-text search in part codes and names with OR logic",
				"🎯 Smart typo correction and query optimization",
				"🗄️ PostgreSQL and ClickHouse dual database support",
				"📷 Image proxy with intelligent caching",
				"⚡ Real-time health monitoring and performance metrics",
				"🌍 CORS-enabled for seamless frontend integration",
			},
			"new_features_v2": []string{
				"AI translation assistant for automotive terms",
				"Enhanced search with both original and translated terms",
				"Fallback enhancement for offline AI scenarios",
				"SQL query logging for debugging",
				"Improved error handling without mock data",
				"Search limited to code and name fields for better accuracy",
			},
			"data_flow": "Frontend → AI Translation → Enhanced Query → PostgreSQL → Structured Results → Frontend",
			"security":  "Open API (implement authentication for production use)",
		},

		"endpoints": map[string]interface{}{
			"health_check": map[string]interface{}{
				"method":     "GET",
				"url":        "/health",
				"purpose":    "Check API and database connectivity",
				"parameters": "None",
				"response_example": map[string]interface{}{
					"status":    "healthy",
					"timestamp": "2025-06-17T05:10:16.7356603+07:00",
					"version":   "25.5.1.2782",
					"database":  "connected",
				},
				"use_cases": []string{"Health monitoring", "Deployment verification", "Database connectivity check"},
			},

			"universal_command": map[string]interface{}{
				"method":       "POST",
				"url":          "/command",
				"purpose":      "Execute any SQL command (CREATE, INSERT, UPDATE, DELETE, ALTER, DROP, etc.)",
				"content_type": "application/json",
				"request_format": map[string]interface{}{
					"query": "string (required) - Any SQL command",
				},
				"request_example": map[string]interface{}{
					"query": "CREATE TABLE IF NOT EXISTS test_api_guide (id UInt32, name String, price Float64, created_at DateTime DEFAULT now()) ENGINE = MergeTree() ORDER BY id",
				},
				"response_format": map[string]interface{}{
					"success":     "boolean - Execution status",
					"message":     "string - Result message",
					"result":      "object - Command execution result",
					"command":     "string - The executed command",
					"duration_ms": "number - Execution time in milliseconds",
				},
				"response_example": map[string]interface{}{
					"success":     true,
					"message":     "Command executed successfully",
					"result":      map[string]interface{}{"query": "CREATE TABLE IF NOT EXISTS test_api_guide (id UInt32, name String, price Float64, created_at DateTime DEFAULT now()) ENGINE = MergeTree() ORDER BY id", "rows_affected": 0, "status": "success"},
					"command":     "CREATE TABLE IF NOT EXISTS test_api_guide (id UInt32, name String, price Float64, created_at DateTime DEFAULT now()) ENGINE = MergeTree() ORDER BY id",
					"duration_ms": 550.0,
				},
				"example_queries": []string{
					"CREATE TABLE IF NOT EXISTS test_api_guide (id UInt32, name String, price Float64, created_at DateTime DEFAULT now()) ENGINE = MergeTree() ORDER BY id",
					"INSERT INTO test_api_guide (id, name, price) VALUES (1, 'Test Product', 99.99)",
					"ALTER TABLE test_api_guide UPDATE price = 199.99 WHERE id = 1",
					"ALTER TABLE test_api_guide DELETE WHERE id = 1",
				},
			},

			"universal_select": map[string]interface{}{
				"method":       "POST",
				"url":          "/select",
				"purpose":      "Execute SELECT queries and retrieve data",
				"content_type": "application/json",
				"request_format": map[string]interface{}{
					"query": "string (required) - Any SELECT query",
				},
				"request_example": map[string]interface{}{
					"query": "SELECT * FROM test_api_guide ORDER BY id",
				},
				"response_format": map[string]interface{}{
					"success":     "boolean - Query status",
					"message":     "string - Result message",
					"data":        "array - Query result data",
					"query":       "string - The executed query",
					"row_count":   "number - Number of rows returned",
					"duration_ms": "number - Execution time in milliseconds",
				},
				"response_example": map[string]interface{}{
					"success": true,
					"message": "Query executed successfully, 1 rows returned",
					"data": []map[string]interface{}{
						{"id": 1, "name": "Test Product", "price": 99.99, "created_at": "2025-06-16T23:40:33Z"},
					},
					"query":       "SELECT * FROM test_api_guide ORDER BY id",
					"row_count":   1,
					"duration_ms": 550.0,
				}, "example_queries": []string{
					"SELECT * FROM ic_inventory ORDER BY name LIMIT 10",
					"SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'",
					"SELECT 1 as test, now() as timestamp",
				},
			},

			"product_search": map[string]interface{}{
				"methods":      []string{"GET", "POST"},
				"urls":         []string{"/v1/search", "/search"},
				"purpose":      "🚀 AI-powered auto parts search with multi-language support and intelligent query enhancement",
				"content_type": "application/json",

				"request_formats": map[string]interface{}{
					"GET": map[string]interface{}{
						"url_parameters": map[string]interface{}{
							"q":      "string (required) - Search query (supports Thai and English)",
							"limit":  "number (optional) - Max results (default: 10, max: 100)",
							"offset": "number (optional) - Pagination offset (default: 0)",
							"ai":     "number (optional) - AI enhancement mode (0=off, 1=on, default: 0)",
						},
						"example_url": "/v1/search?q=โตโยต้า+เบรค&limit=5&ai=1",
					},
					"POST": map[string]interface{}{
						"body_parameters": map[string]interface{}{
							"query":  "string (required) - Search query",
							"limit":  "number (optional) - Max results (default: 10, max: 100)",
							"offset": "number (optional) - Pagination offset (default: 0)",
							"ai":     "number (optional) - AI enhancement mode (0=off, 1=on, default: 0)",
						},
						"example_body": map[string]interface{}{
							"query":  "toyota brake",
							"limit":  5,
							"offset": 0,
							"ai":     1,
						},
					},
				},

				"ai_enhancement": map[string]interface{}{
					"description": "🤖 AI-powered query enhancement using DeepSeek API",
					"features": []string{
						"Thai ↔ English translation (โตโยต้า ↔ toyota)",
						"Typo correction (break → brake)",
						"Term expansion (brake → brake เบรค brakes)",
						"Automotive terminology optimization",
						"Multi-keyword generation for comprehensive search",
					},
					"examples": map[string]interface{}{
						"thai_input":    "เบรค → เบรค brake brakes ระบบเบรค brake pad",
						"typo_fix":      "break → brake เบรค brakes",
						"brand_enhance": "toyoda → toyota โตโยต้า TOYOTA",
					},
					"fallback": "Local enhancement rules when AI API is unavailable",
				},

				"search_algorithm": map[string]interface{}{
					"scope":         "Searches only in 'code' and 'name' fields for maximum accuracy",
					"logic":         "Full-text OR search - each word creates OR conditions",
					"priority":      "Code exact match > Code partial > Name partial",
					"unicode":       "Full Unicode support for Thai text using ILIKE",
					"deduplication": "Automatic duplicate removal",
					"ordering":      "Priority score DESC → Name length ASC → Name alphabetical",
				},

				"response_format": map[string]interface{}{
					"success": "boolean - Search execution status",
					"message": "string - Human-readable result message",
					"data": map[string]interface{}{
						"data":        "array - Product results",
						"total_count": "number - Total matching products found",
						"query":       "string - Final query used (may be AI-enhanced)",
						"duration":    "number - Search execution time in milliseconds",
					},
					"product_fields": map[string]interface{}{
						"id":               "string - Product identifier (same as code)",
						"code":             "string - Product code/SKU",
						"name":             "string - Product name",
						"similarity_score": "number - Search relevance score",
						"balance_qty":      "number - Available quantity",
						"price":            "number - Product price",
						"supplier_code":    "string - Supplier identifier",
						"search_priority":  "number - Internal priority score",
					},
				},

				"response_examples": map[string]interface{}{
					"successful_search": map[string]interface{}{
						"success": true,
						"message": "Search completed successfully",
						"data": map[string]interface{}{
							"data": []map[string]interface{}{
								{
									"id":               "A-88711-0KC50",
									"name":             "ท่อแอร์ TOYOTA",
									"code":             "A-88711-0KC50",
									"similarity_score": 1.0,
									"balance_qty":      0.0,
									"price":            0.0,
									"supplier_code":    "N/A",
									"search_priority":  5,
								},
							},
							"total_count": 2182,
							"query":       "toyota brake (AI enhanced)",
							"duration":    525.3,
						},
					},
					"no_results": map[string]interface{}{
						"success": true,
						"message": "No products found matching your search criteria",
						"data": map[string]interface{}{
							"data":        []interface{}{},
							"total_count": 0,
							"query":       "nonexistent part",
							"duration":    45.2,
						},
					},
					"error_response": map[string]interface{}{
						"success": false,
						"error":   "Search failed: table 'ic_inventory' not found in database",
					},
				},

				"practical_examples": map[string]interface{}{
					"basic_search": map[string]interface{}{
						"description": "Simple search without AI",
						"curl":        `curl -X GET "http://localhost:8008/v1/search?q=toyota&limit=5"`,
						"javascript":  `fetch('/v1/search?q=toyota&limit=5').then(r => r.json())`,
					},
					"ai_enhanced_search": map[string]interface{}{
						"description": "AI-powered search with translation",
						"curl":        `curl -X GET "http://localhost:8008/v1/search?q=เบรค&ai=1&limit=10"`,
						"javascript":  `fetch('/v1/search?q=เบรค&ai=1&limit=10').then(r => r.json())`,
					},
					"post_search": map[string]interface{}{
						"description": "POST request with JSON body",
						"curl":        `curl -X POST "http://localhost:8008/v1/search" -H "Content-Type: application/json" -d '{"query": "toyota compressor", "ai": 1, "limit": 5}'`,
						"javascript":  `fetch('/v1/search', {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({query: 'toyota compressor', ai: 1, limit: 5})})`,
					},
					"pagination": map[string]interface{}{
						"description": "Paginated search results",
						"curl":        `curl -X GET "http://localhost:8008/v1/search?q=brake&limit=20&offset=40"`,
						"javascript":  `fetch('/v1/search?q=brake&limit=20&offset=40').then(r => r.json())`,
					},
				},

				"use_cases": []string{
					"🏪 E-commerce auto parts search",
					"📦 Inventory management systems",
					"🔧 Service workshop part lookup",
					"🤖 AI chatbot integration for customer service",
					"📱 Mobile app product discovery",
					"🌐 Multi-language marketplace platforms",
					"📊 Analytics and business intelligence",
					"🔍 Advanced filtering and recommendation engines",
				},

				"performance_tips": []string{
					"Use specific search terms for faster results",
					"Enable AI (ai=1) for cross-language searches",
					"Use pagination for large result sets",
					"Cache common search results on frontend",
					"Monitor duration field for performance optimization",
				},
			},

			"image_proxy": map[string]interface{}{
				"method":  "GET",
				"url":     "/imgproxy",
				"purpose": "Proxy and cache external images",
				"parameters": map[string]interface{}{
					"url": "string (required) - External image URL to proxy",
				},
				"request_example": "/imgproxy?url=https://example.com/image.jpg",
				"response":        "Image binary data with appropriate headers",
				"features": []string{
					"Image caching for performance",
					"CORS headers for frontend use",
					"Support for various image formats",
					"HEAD request support for metadata",
				},
				"use_cases": []string{
					"Frontend image display",
					"Bypass CORS restrictions",
					"Image caching and optimization",
				},
			}, "database_tables": map[string]interface{}{
				"method":     "GET",
				"url":        "/api/tables",
				"purpose":    "List all available database tables",
				"parameters": "None",
				"response_example": []map[string]interface{}{
					{"name": "products", "engine": "MergeTree", "rows": 1500},
					{"name": "categories", "engine": "MergeTree", "rows": 25},
				},
				"use_cases": []string{
					"Database exploration",
					"Schema discovery",
					"Table listing for admin interfaces",
				},
			},

			"thai_provinces": map[string]interface{}{
				"method":       "POST",
				"url":          "/get/provinces",
				"purpose":      "Get all Thai provinces with Thai and English names",
				"content_type": "application/json",
				"request_body": map[string]interface{}{}, // Empty object
				"response_format": map[string]interface{}{
					"success": "boolean - Request status",
					"message": "string - Success message with count",
					"data":    "array - List of provinces",
				},
				"response_example": map[string]interface{}{
					"success": true,
					"message": "Retrieved 77 provinces successfully",
					"data": []map[string]interface{}{
						{"id": 1, "name_th": "กรุงเทพมหานคร", "name_en": "Bangkok"},
						{"id": 2, "name_th": "สมุทรปราการ", "name_en": "Samut Prakan"},
					},
				},
				"use_cases": []string{
					"Address form population",
					"Location-based filtering",
					"Administrative boundary lookup",
				},
			},

			"thai_amphures": map[string]interface{}{
				"method":       "POST",
				"url":          "/get/amphures",
				"purpose":      "Get all districts (amphures) in a specified province",
				"content_type": "application/json",
				"request_format": map[string]interface{}{
					"province_id": "integer (required) - Province ID from /get/provinces",
				},
				"request_example": map[string]interface{}{
					"province_id": 1,
				},
				"response_format": map[string]interface{}{
					"success": "boolean - Request status",
					"message": "string - Success message with count",
					"data":    "array - List of amphures in the province",
				},
				"response_example": map[string]interface{}{
					"success": true,
					"message": "Retrieved 50 amphures for province_id 1",
					"data": []map[string]interface{}{
						{"id": 1001, "name_th": "เขตพระนคร", "name_en": "Khet Phra Nakhon"},
						{"id": 1002, "name_th": "เขตดุสิต", "name_en": "Khet Dusit"},
					},
				},
				"use_cases": []string{
					"Two-level address selection",
					"District-based operations",
					"Administrative subdivision lookup",
				},
			},

			"thai_tambons": map[string]interface{}{
				"method":       "POST",
				"url":          "/get/tambons",
				"purpose":      "Get all sub-districts (tambons) in a specified amphure and province",
				"content_type": "application/json",
				"request_format": map[string]interface{}{
					"amphure_id":  "integer (required) - Amphure ID from /get/amphures",
					"province_id": "integer (required) - Province ID for validation",
				},
				"request_example": map[string]interface{}{
					"amphure_id":  1001,
					"province_id": 1,
				},
				"response_format": map[string]interface{}{
					"success": "boolean - Request status",
					"message": "string - Success message with count",
					"data":    "array - List of tambons in the amphure",
				},
				"response_example": map[string]interface{}{
					"success": true,
					"message": "Retrieved 12 tambons for amphure_id 1001 in province_id 1",
					"data": []map[string]interface{}{
						{"id": 100101, "name_th": "พระบรมมหาราชวัง", "name_en": "Phra Borom Maha Ratchawang"},
						{"id": 100102, "name_th": "วังบูรพาภิรมย์", "name_en": "Wang Burapha Phirom"},
					},
				},
				"use_cases": []string{
					"Complete address hierarchy",
					"Fine-grained location services",
					"Full administrative address validation",
				},
			},

			"api_documentation": map[string]interface{}{
				"method":   "GET",
				"url":      "/",
				"purpose":  "Basic API information and endpoint list",
				"response": "JSON object with API metadata and available endpoints",
			},
		},

		"ai_agent_instructions": map[string]interface{}{
			"overview": "This API is designed to be AI-friendly with comprehensive JSON responses and clear error messages.",
			"best_practices": []string{
				"Always check /health before executing operations",
				"Use /command for data modification (INSERT, UPDATE, DELETE, CREATE)",
				"Use /select for data retrieval and analytics",
				"Handle both success and error responses appropriately",
				"Check duration_ms for performance monitoring",
				"Use proper JSON formatting in requests",
			},
			"error_handling": map[string]interface{}{
				"all_endpoints_return": "Consistent JSON structure with success boolean",
				"error_format": map[string]interface{}{
					"success": false,
					"error":   "Detailed error message",
					"query":   "The query that failed (for SQL endpoints)",
				},
				"common_errors": []string{
					"Invalid JSON syntax",
					"SQL syntax errors",
					"Database connection issues",
					"Missing required parameters",
				},
			},
			"data_types": map[string]interface{}{
				"clickhouse_types": []string{"UInt32", "String", "Float64", "DateTime", "Array", "Nullable"},
				"json_mapping":     "ClickHouse types automatically mapped to JSON equivalents",
			},
		},

		"integration_examples": map[string]interface{}{
			"curl_examples": map[string]interface{}{
				"health_check":    "curl http://localhost:8008/health",
				"create_table":    "curl -X POST http://localhost:8008/command -H 'Content-Type: application/json' -d '{\"query\": \"CREATE TABLE test (id UInt32, name String) ENGINE = MergeTree() ORDER BY id\"}'",
				"insert_data":     "curl -X POST http://localhost:8008/command -H 'Content-Type: application/json' -d '{\"query\": \"INSERT INTO test VALUES (1, 'hello')\"}'",
				"select_data":     "curl -X POST http://localhost:8008/select -H 'Content-Type: application/json' -d '{\"query\": \"SELECT * FROM test\"}'",
				"search_products": "curl -X POST http://localhost:8008/search -H 'Content-Type: application/json' -d '{\"query\": \"laptop\", \"limit\": 5}'",
			},
			"javascript_fetch": map[string]interface{}{
				"async_example": `
async function executeSQL(endpoint, query) {
  const response = await fetch('http://localhost:8008/' + endpoint, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query: query })
  });
  return await response.json();
}

// Usage
const result = await executeSQL('select', 'SELECT * FROM products LIMIT 5');
`,
			},
		},

		"production_considerations": map[string]interface{}{
			"security": []string{
				"Add authentication (JWT/API keys)",
				"Implement query validation/whitelisting",
				"Configure CORS for specific domains",
				"Add rate limiting",
				"Enable HTTPS",
			},
			"performance": []string{
				"Monitor query execution times",
				"Implement query caching",
				"Use connection pooling (already implemented)",
				"Add query timeout limits",
				"Monitor memory usage",
			},
			"monitoring": []string{
				"Log all SQL executions",
				"Monitor error rates",
				"Track response times",
				"Set up health check alerts",
			},
		},

		"database_schema": map[string]interface{}{
			"note": "Use /api/tables to discover available tables",
			"common_tables": []string{
				"products - Product catalog data",
				"categories - Product categories",
				"system.tables - ClickHouse system table for metadata",
			},
			"query_tips": []string{
				"Always use LIMIT for large result sets",
				"Use proper WHERE clauses for filtering",
				"Leverage ClickHouse's columnar storage with SELECT specific columns",
				"Use ORDER BY for consistent result ordering",
			},
		},

		"support_information": map[string]interface{}{
			"troubleshooting": map[string]interface{}{
				"connection_issues":  "Check /health endpoint and ClickHouse server status",
				"query_errors":       "Validate SQL syntax and table/column names",
				"cors_issues":        "API includes CORS headers for localhost and * origins",
				"performance_issues": "Monitor duration_ms in responses",
			},
			"logs":    "Check server console for detailed error information",
			"testing": "Use included Postman collection or test frontend HTML",
		},
	}

	c.JSON(http.StatusOK, guide)
}

// Thai Administrative Data Endpoints

// GetProvinces godoc
// @Summary Get all Thai provinces
// @Description Retrieve all provinces in Thailand with Thai and English names
// @Tags thai-admin
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse{data=[]models.Province}
// @Router /get/provinces [post]
func (h *APIHandler) GetProvinces(c *gin.Context) {
	provinces, err := h.thaiAdminService.GetProvinces()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to load provinces: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    provinces,
		Message: fmt.Sprintf("Retrieved %d provinces successfully", len(provinces)),
	})
}

// GetAmphures godoc
// @Summary Get all amphures in a province
// @Description Retrieve all districts (amphures) in a specified province
// @Tags thai-admin
// @Accept json
// @Produce json
// @Param request body models.AmphureRequest true "Province ID"
// @Success 200 {object} models.APIResponse{data=[]models.Amphure}
// @Router /get/amphures [post]
func (h *APIHandler) GetAmphures(c *gin.Context) {
	var req models.AmphureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	amphures, err := h.thaiAdminService.GetAmphuresByProvinceID(req.ProvinceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to load amphures: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    amphures,
		Message: fmt.Sprintf("Retrieved %d amphures for province_id %d", len(amphures), req.ProvinceID),
	})
}

// GetTambons godoc
// @Summary Get all tambons in an amphure
// @Description Retrieve all sub-districts (tambons) in a specified amphure and province
// @Tags thai-admin
// @Accept json
// @Produce json
// @Param request body models.TambonRequest true "Amphure and Province IDs"
// @Success 200 {object} models.APIResponse{data=[]models.Tambon}
// @Router /get/tambons [post]
func (h *APIHandler) GetTambons(c *gin.Context) {
	var req models.TambonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	tambons, err := h.thaiAdminService.GetTambonsByAmphureAndProvince(req.AmphureID, req.ProvinceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to load tambons: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    tambons,
		Message: fmt.Sprintf("Retrieved %d tambons for amphure_id %d in province_id %d", len(tambons), req.AmphureID, req.ProvinceID),
	})
}

// FindByZipCode godoc
// @Summary Find location by zip code
// @Description Find complete location information (province, amphure, tambon) by Thai postal code
// @Tags thai-admin
// @Accept json
// @Produce json
// @Param request body models.ZipCodeRequest true "Zip code to search"
// @Success 200 {object} models.APIResponse{data=[]models.CompleteLocationData}
// @Router /get/findbyzipcode [post]
func (h *APIHandler) FindByZipCode(c *gin.Context) {
	var req models.ZipCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	locations, err := h.thaiAdminService.FindByZipCode(req.ZipCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to find locations: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    locations,
		Message: fmt.Sprintf("Found %d locations for zip code %d", len(locations), req.ZipCode),
	})
}

// SearchProductsByVector godoc
// @Summary Search products using vector database first, then PostgreSQL
// @Description Search for products using Weaviate vector database to get IC codes (primary) or barcodes (fallback), then search PostgreSQL for detailed product information
// @Tags search
// @Accept json
// @Produce json
// @Param search body models.SearchParameters true "Search parameters (for POST requests)"
// @Param q query string false "Search query (for GET requests)"
// @Param limit query integer false "Maximum number of results (for GET requests)"
// @Param offset query integer false "Offset for pagination (for GET requests)"
// @Success 200 {object} models.APIResponse
// @Router /search-by-vector [post]
// @Router /search-by-vector [get]
func (h *APIHandler) SearchProductsByVector(c *gin.Context) {
	startTime := time.Now()

	var params models.SearchParameters

	// Check if this is a GET request with query parameters
	if c.Request.Method == "GET" {
		// Parse URL query parameters
		params.Query = c.Query("q")
		if limit := c.Query("limit"); limit != "" {
			if l, err := strconv.Atoi(limit); err == nil {
				params.Limit = l
			}
		}
		if offset := c.Query("offset"); offset != "" {
			if o, err := strconv.Atoi(offset); err == nil {
				params.Offset = o
			}
		}
	} else {
		// Parse JSON body for POST requests
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "Invalid JSON format: " + err.Error(),
			})
			return
		}
	}

	log.Printf("🔍 [VECTOR-SEARCH] Parsed parameters: query='%s', limit=%d, offset=%d", params.Query, params.Limit, params.Offset)

	// Validate query
	if params.Query == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Query parameter is required",
		})
		return
	}

	query := params.Query

	// Set default values
	limit := params.Limit
	if limit <= 0 {
		limit = 20 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	// Enhanced logging
	fmt.Printf("\n🚀 [VECTOR-SEARCH] === STARTING SEARCH ===\n")
	fmt.Printf("   📝 Query: '%s'\n", query)
	fmt.Printf("   📊 Limit: %d, Offset: %d\n", limit, offset)
	fmt.Printf("   =====================================\n")
	ctx := c.Request.Context()

	// Step 1: Search Weaviate vector database first to get IC codes and barcodes
	if h.weaviateService == nil {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Message: "Vector search service not available",
		})
		return
	}

	// Search vector database with higher limit to get more barcodes for better matching
	vectorLimit := limit * 3 // Get more results from vector DB to compensate for potential mismatches
	if vectorLimit > 300 {
		vectorLimit = 300
	}

	vectorProducts, err := h.weaviateService.SearchProducts(ctx, query, vectorLimit)
	if err != nil {
		log.Printf("❌ [VECTOR-SEARCH] Weaviate vector search failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Vector search failed: " + err.Error(),
		})
		return
	}

	log.Printf("🎲 [VECTOR-SEARCH] Weaviate returned %d products from vector database", len(vectorProducts))

	if len(vectorProducts) == 0 {
		log.Printf("ℹ️ [VECTOR-SEARCH] No products found in Weaviate vector database")
		// Return empty results instead of error
		results := &services.VectorSearchResponse{
			Data:       []services.SearchResult{},
			TotalCount: 0,
			Query:      query,
			Duration:   time.Since(startTime).Seconds() * 1000,
		}

		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Data:    results,
			Message: "No products found matching the query",
		})
		return
	} // Step 2: Extract IC codes from vector search results (preferred) or fallback to barcodes
	icCodes, relevanceMap := h.weaviateService.GetICCodesWithRelevance(vectorProducts)

	var searchResults []map[string]interface{}
	var totalCount int
	var searchMethod string

	if len(icCodes) > 0 {
		searchMethod = "IC Code"
		log.Printf("🎯 [VECTOR-SEARCH] Extracting IC codes from Weaviate: %d codes found", len(icCodes))

		// Get barcode mapping for IC codes
		barcodeMapping := h.weaviateService.GetICCodeToBarcodeMap(vectorProducts)

		// Step 3: Search PostgreSQL using the IC codes with relevance scores and barcode mapping
		searchResults, totalCount, err = h.postgreSQLService.SearchProductsByBarcodesWithRelevanceAndBarcodeMap(ctx, icCodes, relevanceMap, barcodeMapping, limit, offset)
		if err != nil {
			log.Printf("❌ [VECTOR-SEARCH] PostgreSQL search by IC codes failed: %v", err)
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Database search failed: " + err.Error(),
			})
			return
		}

		if len(searchResults) > 0 {
			log.Printf("✅ [VECTOR-SEARCH] Found %d products using IC codes", len(searchResults))
		} else {
			log.Printf("⚠️ [VECTOR-SEARCH] No products found with IC codes, trying barcodes as fallback...")
			// Fallback to barcode search
			barcodes, barcodeRelevanceMap := h.weaviateService.GetBarcodesWithRelevance(vectorProducts)
			if len(barcodes) > 0 {
				searchMethod = "Barcode (Fallback)"
				log.Printf("🔄 [VECTOR-SEARCH] Fallback: extracting barcodes: %d codes found", len(barcodes))

				// Get barcode mapping for barcodes
				barcodeMappingFallback := h.weaviateService.GetBarcodeToBarcodeMap(vectorProducts)

				// Step 3: Search PostgreSQL using the barcodes with relevance scores and barcode mapping
				searchResults, totalCount, err = h.postgreSQLService.SearchProductsByBarcodesWithRelevanceAndBarcodeMap(ctx, barcodes, barcodeRelevanceMap, barcodeMappingFallback, limit, offset)
				if err != nil {
					log.Printf("❌ [VECTOR-SEARCH] PostgreSQL fallback search by barcodes failed: %v", err)
					c.JSON(http.StatusInternalServerError, models.APIResponse{
						Success: false,
						Message: "Database search failed: " + err.Error(),
					})
					return
				}

				if len(searchResults) > 0 {
					log.Printf("✅ [VECTOR-SEARCH] Found %d products using barcode fallback", len(searchResults))
				}
			}
		}
	} else {
		// No IC codes available, use barcodes
		barcodes, barcodeRelevanceMap := h.weaviateService.GetBarcodesWithRelevance(vectorProducts)
		searchMethod = "Barcode (Primary)"
		log.Printf("🎯 [VECTOR-SEARCH] No IC codes available, extracting barcodes: %d codes found", len(barcodes))

		// Get barcode mapping for barcodes
		barcodeMappingPrimary := h.weaviateService.GetBarcodeToBarcodeMap(vectorProducts)

		// Step 3: Search PostgreSQL using the barcodes with relevance scores and barcode mapping
		searchResults, totalCount, err = h.postgreSQLService.SearchProductsByBarcodesWithRelevanceAndBarcodeMap(ctx, barcodes, barcodeRelevanceMap, barcodeMappingPrimary, limit, offset)
		if err != nil {
			log.Printf("❌ [VECTOR-SEARCH] PostgreSQL search by barcodes failed: %v", err)
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Database search failed: " + err.Error(),
			})
			return
		}

		if len(searchResults) > 0 {
			log.Printf("✅ [VECTOR-SEARCH] Found %d products using barcodes", len(searchResults))
		}
	}

	// Convert PostgreSQL results to the expected format
	var convertedResults []services.SearchResult
	for _, result := range searchResults {
		convertedResult := services.SearchResult{
			ID:               getStringValue(result, "id"),
			Code:             getStringValue(result, "code"),
			Name:             getStringValue(result, "name"),
			Price:            getFloat64Value(result, "price"),
			Unit:             getStringValue(result, "unit"),
			SupplierCode:     getStringValue(result, "supplier_code"),
			ImgURL:           getStringValue(result, "img_url"),
			SimilarityScore:  getFloat64Value(result, "similarity_score"),
			SalePrice:        getFloat64Value(result, "sale_price"),
			PremiumWord:      getStringValue(result, "premium_word"),
			DiscountPrice:    getFloat64Value(result, "discount_price"),
			DiscountPercent:  getFloat64Value(result, "discount_percent"),
			FinalPrice:       getFloat64Value(result, "final_price"),
			SoldQty:          getFloat64Value(result, "sold_qty"),
			MultiPacking:     int(getFloat64Value(result, "multi_packing")),
			MultiPackingName: getStringValue(result, "multi_packing_name"),
			Barcodes:         getStringValue(result, "barcodes"),
			Barcode:          getStringValue(result, "barcode"), // Add the barcode field from Weaviate
			QtyAvailable:     getFloat64Value(result, "qty_available"),
			BalanceQty:       getFloat64Value(result, "balance_qty"),
			SearchPriority:   int(getFloat64Value(result, "search_priority")),
		}
		convertedResults = append(convertedResults, convertedResult)
	}

	// Create response in the expected format
	results := &services.VectorSearchResponse{
		Data:       convertedResults,
		TotalCount: totalCount,
		Query:      query,
		Duration:   time.Since(startTime).Seconds() * 1000,
	}
	duration := time.Since(startTime).Seconds() * 1000

	// Enhanced search results logging
	fmt.Printf("\n🎯 [VECTOR-SEARCH] === SEARCH RESULTS SUMMARY ===\n")
	fmt.Printf("   📝 Query: '%s'\n", query)
	fmt.Printf("   🔗 Search Method: %s\n", searchMethod)
	fmt.Printf("   🎲 Vector Database: %d products found\n", len(vectorProducts))
	fmt.Printf("   📊 PostgreSQL Total: %d records\n", results.TotalCount)
	fmt.Printf("   📋 Returned Results: %d products\n", len(results.Data))
	fmt.Printf("   📄 Page Info: page %d (offset: %d, limit: %d)\n", (offset/limit)+1, offset, limit)
	fmt.Printf("   ⏱️  Processing Time: %.1fms\n", duration)
	if len(results.Data) > 0 {
		fmt.Printf("   🏆 Top Results:\n")
		for i, product := range results.Data {
			if i >= 3 {
				break
			}
			fmt.Printf("     %d. [%s] %s (Relevance: %.1f%%)\n", i+1, product.Code, product.Name, product.SimilarityScore)
		}
	} else {
		fmt.Printf("   ❌ No results found\n")
	}

	fmt.Printf("   ===============================\n")
	fmt.Printf("✅ [VECTOR-SEARCH] COMPLETED (%.1fms)\n\n", duration)
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    results,
		Message: "Vector search completed successfully",
	})
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper functions for type conversion from map[string]interface{}
func getStringValue(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
		if b, ok := val.([]uint8); ok {
			return string(b)
		}
		return fmt.Sprintf("%v", val)
	}
	return ""
}

func getFloat64Value(data map[string]interface{}, key string) float64 {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return 0.0
}

// enhanceQueryWithAI enhances the search query using DeepSeek AI API
// This function applies AI-based query enhancement when ai=1 parameter is used
func (h *APIHandler) enhanceQueryWithAI(originalQuery string) string {
	log.Printf("🤖 [ai_enhance] Processing query with DeepSeek AI: '%s'", originalQuery)

	// Call DeepSeek API
	enhancedQuery, err := h.callDeepSeekAPI(originalQuery)
	if err != nil {
		log.Printf("❌ [ai_enhance] DeepSeek API error: %v, using fallback", err)
		return h.fallbackEnhancement(originalQuery)
	}

	if enhancedQuery != originalQuery {
		log.Printf("🤖 [ai_enhance] DeepSeek enhanced: '%s' -> '%s'", originalQuery, enhancedQuery)
		return enhancedQuery
	}

	log.Printf("🤖 [ai_enhance] DeepSeek returned same query: '%s'", originalQuery)
	return originalQuery
}

// callDeepSeekAPI calls the DeepSeek API for query enhancement
func (h *APIHandler) callDeepSeekAPI(query string) (string, error) {
	prompt := fmt.Sprintf(`You are a translation assistant for automotive parts database search. Your job is to help users find parts by providing both the original term and its translations, creating the most comprehensive search query possible.

RULES:
1. Always include BOTH the original query AND translations
2. Support Thai ↔ English translation for automotive terms
3. Fix common typos and provide correct spellings
4. Include both singular and plural forms when relevant
5. Add related terms that users might search for
6. Return multiple terms separated by spaces for OR search logic

TRANSLATION EXAMPLES:
- "โตโยต้า" → "โตโยต้า toyota TOYOTA"
- "เบรค" → "เบรค brake brakes"
- "toyota" → "toyota โตโยต้า TOYOTA"
- "brake" → "brake เบรค brakes"
- "โชคอัมพาต" → "โชคอัมพาต shock absorber damper"

TYPO CORRECTION:
- "toyoda" → "toyota โตโยต้า TOYOTA"
- "break" → "brake เบรค brakes"
- "compresser" → "compressor คอมเพรสเซอร์"

Return ALL relevant search terms separated by spaces (for OR search logic).

User query: %s

Enhanced search terms:`, query)

	reqBody := DeepSeekRequest{
		Model: DeepSeekModel,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", DeepSeekAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+DeepSeekAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var deepSeekResp DeepSeekResponse
	if err := json.Unmarshal(body, &deepSeekResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(deepSeekResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	enhancedQuery := strings.TrimSpace(deepSeekResp.Choices[0].Message.Content)
	return enhancedQuery, nil
}

// fallbackEnhancement provides a fallback enhancement when DeepSeek API fails
// Returns multiple search terms including original and translations
func (h *APIHandler) fallbackEnhancement(originalQuery string) string {
	log.Printf("🔄 [ai_enhance] Using fallback enhancement for: '%s'", originalQuery)

	// Normalize the query (trim and lowercase for comparison)
	normalizedQuery := strings.TrimSpace(strings.ToLower(originalQuery))

	// Translation and enhancement rules - return multiple terms
	enhancements := map[string]string{
		// Thai to English with both forms
		"เบรค":      "เบรค brake brakes",
		"ผ้าเบรค":   "ผ้าเบรค brake pad brakes",
		"โตโยต้า":   "โตโยต้า toyota TOYOTA",
		"น้ำมัน":    "น้ำมัน oil",
		"ไฟ":        "ไฟ light lights",
		"ล้อ":       "ล้อ wheel wheels",
		"ยาง":       "ยาง tire tires",
		"แบตเตอรี่": "แบตเตอรี่ battery",
		"คอยล์":     "คอยล์ coil coils",
		"โชคอัมพาต": "โชคอัมพาต shock absorber damper",

		// English to Thai with variants
		"toyota":  "toyota โตโยต้า TOYOTA",
		"brake":   "brake เบรค brakes",
		"oil":     "oil น้ำมัน",
		"light":   "light ไฟ lights",
		"wheel":   "wheel ล้อ wheels",
		"tire":    "tire ยาง tires",
		"battery": "battery แบตเตอรี่",
		"coil":    "coil คอยล์ coils",
		"shock":   "shock โชคอัมพาต absorber",

		// Common typos with corrections
		"break":      "brake เบรค brakes",
		"toyoda":     "toyota โตโยต้า TOYOTA",
		"compresser": "compressor คอมเพรสเซอร์",
	}

	// Check for direct enhancements
	if enhanced, exists := enhancements[normalizedQuery]; exists {
		log.Printf("🔄 [ai_enhance] Fallback enhancement: '%s' -> '%s'", originalQuery, enhanced)
		return enhanced
	}

	// Check for partial matches and expansions
	var allTerms []string
	allTerms = append(allTerms, originalQuery) // Always include original

	for key, value := range enhancements {
		if strings.Contains(normalizedQuery, key) {
			// Add all terms from the enhancement
			terms := strings.Fields(value)
			allTerms = append(allTerms, terms...)
			break
		}
	}

	// Remove duplicates and return
	uniqueTerms := make(map[string]bool)
	var result []string
	for _, term := range allTerms {
		if !uniqueTerms[term] {
			uniqueTerms[term] = true
			result = append(result, term)
		}
	}

	enhanced := strings.Join(result, " ")
	if enhanced != originalQuery {
		log.Printf("🔄 [ai_enhance] Fallback partial match: '%s' -> '%s'", originalQuery, enhanced)
		return enhanced
	}

	// If no enhancement found, return original
	log.Printf("🔄 [ai_enhance] Fallback: no enhancement for '%s'", originalQuery)
	return originalQuery
}
