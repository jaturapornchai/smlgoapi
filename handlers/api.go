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
	thaiAdminService  *services.ThaiAdminService
}

func NewAPIHandler(clickHouseService *services.ClickHouseService) *APIHandler {
	vectorDB := services.NewTFIDFVectorDatabase(clickHouseService)
	imageProxyService := services.NewImageProxy()
	thaiAdminService := services.NewThaiAdminService()
	return &APIHandler{
		clickHouseService: clickHouseService,
		vectorDB:          vectorDB,
		imageProxyService: imageProxyService,
		thaiAdminService:  thaiAdminService,
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
		"version":       "1.0.0",
		"description":   "ClickHouse REST API with universal SQL execution, product search, and image proxy capabilities",
		"base_url":      "http://localhost:8008",
		"documentation": "Complete API guide for AI agents and developers",
		"last_updated":  "2025-06-17",

		"concepts": map[string]interface{}{
			"overview": "SMLGOAPI is a REST API that provides universal access to ClickHouse database operations, advanced product search with vector similarity, and image proxy services.",
			"core_features": []string{
				"Universal SQL execution via JSON (any INSERT, UPDATE, DELETE, CREATE, SELECT)",
				"Multi-step product search (code → name → vector similarity)",
				"Image proxy with caching",
				"Real-time health monitoring",
				"CORS-enabled for frontend integration",
			},
			"data_flow": "Frontend → JSON Request → API → ClickHouse → JSON Response → Frontend",
			"security":  "Open API (add authentication in production)",
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
				},
				"example_queries": []string{
					"SELECT * FROM test_api_guide ORDER BY id",
					"SELECT name FROM system.tables LIMIT 3",
					"SELECT 1 as test, now() as timestamp",
				},
			},

			"product_search": map[string]interface{}{
				"method":       "POST",
				"url":          "/search",
				"purpose":      "Multi-step product search with priority: code → name → vector similarity",
				"content_type": "application/json",
				"request_format": map[string]interface{}{
					"query": "string (required) - Search term",
					"limit": "number (optional) - Max results (default: 10)",
				},
				"request_example": map[string]interface{}{
					"query": "laptop gaming",
					"limit": 5,
				},
				"response_format": map[string]interface{}{
					"success": "boolean - Search status",
					"message": "string - Search result message",
					"data":    "array - Product results with flattened structure",
					"metadata": map[string]interface{}{
						"query":        "string - Search term",
						"total_found":  "number - Total products found",
						"search_steps": "array - Search steps executed",
						"duration_ms":  "number - Search time",
					},
				},
				"search_logic": map[string]interface{}{
					"step_1":        "Code search (priority 1) - Exact product code matching",
					"step_2":        "Name search (priority 2) - Product name pattern matching",
					"step_3":        "Vector search (priority 3) - TF-IDF similarity search",
					"deduplication": "Remove duplicates across steps",
					"ranking":       "Results ordered by search step priority, then relevance",
				},
				"response_example": map[string]interface{}{
					"success": true,
					"message": "Search completed successfully",
					"data": []map[string]interface{}{
						{
							"product_code":    "LAP001",
							"product_name":    "Gaming Laptop RTX 4080",
							"price":           1299.99,
							"category":        "Electronics",
							"search_step":     1,
							"relevance_score": 1.0,
						},
					},
					"metadata": map[string]interface{}{
						"query":        "laptop gaming",
						"total_found":  1,
						"search_steps": []string{"code_search", "name_search", "vector_search"},
						"duration_ms":  156.7,
					},
				},
				"use_cases": []string{
					"E-commerce product search",
					"Inventory lookup",
					"Product recommendations",
					"Catalog browsing",
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
