package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"smlgoapi/config"
	"smlgoapi/models"
	"smlgoapi/services"

	"github.com/gin-gonic/gin"
)

// COMMENTED OUT FOR SPEED TESTING - DeepSeek API structures
/*
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
*/

type APIHandler struct {
	clickHouseService *services.ClickHouseService
	postgreSQLService *services.PostgreSQLService
	vectorDB          *services.TFIDFVectorDatabase
	thaiAdminService  *services.ThaiAdminService
	weaviateService   *services.WeaviateService
}

func NewAPIHandler(clickHouseService *services.ClickHouseService, postgreSQLService *services.PostgreSQLService) *APIHandler {
	var vectorDB *services.TFIDFVectorDatabase
	if clickHouseService != nil {
		vectorDB = services.NewTFIDFVectorDatabase(clickHouseService)
	}
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
// @Description Execute any PostgreSQL SQL command (INSERT, UPDATE, DELETE, CREATE, etc.) via JSON request with optional database selection
// @Tags database
// @Accept json
// @Produce json
// @Param command body models.CommandRequest true "Command to execute with optional collection/database"
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

	// Determine which database to connect to
	targetDB := "default"
	if commandReq.Collection != "" {
		targetDB = commandReq.Collection
	}

	log.Printf("🐘 [pgcommand] Executing PostgreSQL command in database '%s': %s", targetDB, commandReq.Query)

	ctx := c.Request.Context()

	var result interface{}
	var err error

	// Execute command using appropriate service method
	if commandReq.Collection != "" {
		// Use collection-specific method (creates new connection)
		result, err = h.postgreSQLService.ExecuteCommandInCollection(ctx, commandReq.Query, commandReq.Collection)
	} else {
		// Use default connection
		result, err = h.postgreSQLService.ExecuteCommand(ctx, commandReq.Query)
	}

	duration := float64(time.Since(startTime).Nanoseconds()) / 1e6

	if err != nil {
		log.Printf("❌ [pgcommand] Execution failed in database '%s': %v", targetDB, err)
		c.JSON(http.StatusInternalServerError, models.CommandResponse{
			Success:  false,
			Error:    fmt.Sprintf("PostgreSQL command execution failed: %s", err.Error()),
			Command:  commandReq.Query,
			Database: targetDB,
			Duration: duration,
		})
		return
	}

	log.Printf("✅ [pgcommand] Execution successful in database '%s' in %.2fms", targetDB, duration)

	c.JSON(http.StatusOK, models.CommandResponse{
		Success:  true,
		Message:  fmt.Sprintf("PostgreSQL command executed successfully in database '%s'", targetDB),
		Result:   result,
		Command:  commandReq.Query,
		Database: targetDB,
		Duration: duration,
	})
}

// PgSelectEndpoint godoc
// @Summary Execute PostgreSQL SELECT query
// @Description Execute PostgreSQL SELECT query and return data via JSON request with optional database selection
// @Tags database
// @Accept json
// @Produce json
// @Param select body models.SelectRequest true "SELECT query to execute with optional collection/database"
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

	// Determine which database to connect to
	targetDB := "default"
	if selectReq.Collection != "" {
		targetDB = selectReq.Collection
	}

	log.Printf("🐘 [pgselect] Executing PostgreSQL query in database '%s': %s", targetDB, selectReq.Query)

	ctx := c.Request.Context()

	var data []interface{}
	var err error

	// Execute select query using appropriate service method
	if selectReq.Collection != "" {
		// Use collection-specific method (creates new connection)
		data, err = h.postgreSQLService.ExecuteSelectInCollection(ctx, selectReq.Query, selectReq.Collection)
	} else {
		// Use default connection
		data, err = h.postgreSQLService.ExecuteSelect(ctx, selectReq.Query)
	}

	duration := float64(time.Since(startTime).Nanoseconds()) / 1e6

	if err != nil {
		log.Printf("❌ [pgselect] Query failed in database '%s': %v", targetDB, err)
		c.JSON(http.StatusInternalServerError, models.SelectResponse{
			Success:  false,
			Error:    fmt.Sprintf("PostgreSQL query execution failed: %s", err.Error()),
			Query:    selectReq.Query,
			Database: targetDB,
			Duration: duration,
		})
		return
	}

	rowCount := len(data)
	log.Printf("✅ [pgselect] Query successful in database '%s': %d rows returned in %.2fms", targetDB, rowCount, duration)

	c.JSON(http.StatusOK, models.SelectResponse{
		Success:  true,
		Message:  fmt.Sprintf("PostgreSQL query executed successfully in database '%s', %d rows returned", targetDB, rowCount),
		Data:     data,
		Query:    selectReq.Query,
		Database: targetDB,
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

			"database_tables": map[string]interface{}{
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
				"health_check": "curl http://localhost:8008/health",
				"create_table": "curl -X POST http://localhost:8008/command -H 'Content-Type: application/json' -d '{\"query\": \"CREATE TABLE test (id UInt32, name String) ENGINE = MergeTree() ORDER BY id\"}'",
				"insert_data":  "curl -X POST http://localhost:8008/command -H 'Content-Type: application/json' -d '{\"query\": \"INSERT INTO test VALUES (1, 'hello')\"}'",
				"select_data":  "curl -X POST http://localhost:8008/select -H 'Content-Type: application/json' -d '{\"query\": \"SELECT * FROM test\"}'",
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
// @Param search body models.SearchParameters true "Search parameters"
// @Success 200 {object} models.APIResponse
// @Router /search-product [post]

func (h *APIHandler) SearchProductsByVector(c *gin.Context) {
	startTime := time.Now()

	var params models.SearchParameters
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid JSON format: " + err.Error(),
		})
		return
	}

	if params.Query == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Query parameter is required",
		})
		return
	}

	query := params.Query
	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}
	collection := params.Collection // Get collection from request body

	ctx := c.Request.Context()

	// 1. Exact match by barcode (use collection-specific search if collection is provided)
	if h.postgreSQLService != nil {
		var barcodeResults []map[string]interface{}
		var total int
		var err error

		if collection != "" {
			// Use collection-specific search with exact barcode pattern
			barcodeResults, total, err = h.postgreSQLService.SearchProductsByExactBarcodeInCollection(ctx, query, limit, offset, collection)
		} else {
			barcodeResults, total, err = h.postgreSQLService.SearchProductsByExactBarcode(ctx, query, limit, offset)
		}

		if err == nil && total > 0 {
			var convertedResults []services.SearchResult
			for _, result := range barcodeResults {
				convertedResults = append(convertedResults, services.SearchResult{
					Code:            getStringValue(result, "code"),
					Name:            getStringValue(result, "name"),
					Barcode:         getStringValue(result, "barcode"),
					ImgURL:          getStringValue(result, "img_url"),
					SimilarityScore: getFloat64Value(result, "similarity_score"),
					Unit:            getStringValue(result, "unit"),
					// เพิ่ม field อื่นๆ ตามที่ SearchResult รองรับ
				})
			}
			results := &services.VectorSearchResponse{
				Data:       convertedResults,
				TotalCount: total,
				Query:      query,
				Duration:   time.Since(startTime).Seconds() * 1000,
			}
			c.JSON(http.StatusOK, models.APIResponse{
				Success: true,
				Data:    results,
				Message: "Found by barcode",
			})
			return
		}
	}

	// 2. Exact match by ic_code (use collection-specific search if collection is provided)
	if h.postgreSQLService != nil {
		var icCodeResults []map[string]interface{}
		var total int
		var err error

		if collection != "" {
			// Use collection-specific search with exact code pattern
			icCodeResults, total, err = h.postgreSQLService.SearchProductsByExactCodeInCollection(ctx, query, limit, offset, collection)
		} else {
			icCodeResults, total, err = h.postgreSQLService.SearchProductsByExactCode(ctx, query, limit, offset)
		}

		if err == nil && total > 0 {
			var convertedResults []services.SearchResult
			for _, result := range icCodeResults {
				convertedResults = append(convertedResults, services.SearchResult{
					Code:            getStringValue(result, "code"),
					Name:            getStringValue(result, "name"),
					Barcode:         getStringValue(result, "barcode"),
					ImgURL:          getStringValue(result, "img_url"),
					SimilarityScore: getFloat64Value(result, "similarity_score"),
					Unit:            getStringValue(result, "unit"),
					// เพิ่ม field อื่นๆ ตามที่ SearchResult รองรับ
				})
			}
			results := &services.VectorSearchResponse{
				Data:       convertedResults,
				TotalCount: total,
				Query:      query,
				Duration:   time.Since(startTime).Seconds() * 1000,
			}
			c.JSON(http.StatusOK, models.APIResponse{
				Success: true,
				Data:    results,
				Message: "Found by ic_code",
			})
			return
		}
	}

	// 3. Vector search (Weaviate)
	if h.weaviateService == nil {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Message: "Vector search service unavailable",
		})
		return
	}

	vectorProducts, err := h.weaviateService.SearchProductsWithCollection(ctx, query, limit, collection)
	if err != nil {
		log.Printf("❌ [VECTOR-SEARCH] Weaviate vector search failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Vector search failed: " + err.Error(),
		})
		return
	}

	if len(vectorProducts) == 0 {
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Data:    []services.SearchResult{},
			Message: "No products found matching the query",
		})
		return
	}

	icCodes, relevanceMap := h.weaviateService.GetICCodesWithRelevance(vectorProducts)
	var searchResults []map[string]interface{}
	var totalCount int

	if len(icCodes) > 0 && h.postgreSQLService != nil {
		barcodeMapping := h.weaviateService.GetICCodeToBarcodeMap(vectorProducts)

		// Use collection-specific search if collection is provided
		if collection != "" {
			searchResults, totalCount, err = h.postgreSQLService.SearchProductsByBarcodesWithRelevanceAndBarcodeMapInCollection(ctx, icCodes, relevanceMap, barcodeMapping, limit, offset, collection)
		} else {
			searchResults, totalCount, err = h.postgreSQLService.SearchProductsByBarcodesWithRelevanceAndBarcodeMap(ctx, icCodes, relevanceMap, barcodeMapping, limit, offset)
		}
		if err != nil {
			log.Printf("❌ [VECTOR-SEARCH] PostgreSQL search by IC codes failed: %v", err)
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Database search failed: " + err.Error(),
			})
			return
		}
	} else {
		for _, v := range vectorProducts {
			result := services.SearchResult{
				Code:            v.ICCode,
				Name:            v.Name,
				Barcode:         v.Barcode,
				SimilarityScore: v.Relevance,
			}
			searchResults = append(searchResults, map[string]interface{}{
				"code":             result.Code,
				"name":             result.Name,
				"barcode":          result.Barcode,
				"similarity_score": result.SimilarityScore,
			})
		}
		totalCount = len(searchResults)
	}

	var convertedResults []services.SearchResult
	for _, result := range searchResults {
		convertedResults = append(convertedResults, services.SearchResult{
			Code:            getStringValue(result, "code"),
			Name:            getStringValue(result, "name"),
			Barcode:         getStringValue(result, "barcode"),
			ImgURL:          getStringValue(result, "img_url"),
			SimilarityScore: getFloat64Value(result, "similarity_score"),
			Unit:            getStringValue(result, "unit"),
		})
	}

	results := &services.VectorSearchResponse{
		Data:       convertedResults,
		TotalCount: totalCount,
		Query:      query,
		Duration:   time.Since(startTime).Seconds() * 1000,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    results,
		Message: "Vector search completed successfully",
	})
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

// COMMENTED OUT FOR SPEED TESTING - DeepSeek AI Enhancement Functions
/*
// enhanceQueryForVectorSearch enhances search query specifically for vector search
// ปรับปรุงคำค้นหาสำหรับ vector search โดยเฉพาะ พร้อม fallback เมื่อ API ล้มเหลว
func (h *APIHandler) enhanceQueryForVectorSearch(originalQuery string) (string, error) {
	log.Printf("🤖 [vector-enhance] Processing query for vector search: '%s'", originalQuery)

	// ลอง DeepSeek API ก่อน แต่ถ้าล้มเหลวจะใช้ fallback ทันที
	enhancedQuery, err := h.callDeepSeekAPIForVector(originalQuery)
	if err != nil {
		log.Printf("⚠️ [vector-enhance] DeepSeek API failed (%v), using fallback enhancement", err)
		return originalQuery, nil
	}

	// ลบคำซ้ำและทำความสะอาด
	cleanedQuery := h.removeDuplicateWords(enhancedQuery)

	if cleanedQuery != originalQuery {
		log.Printf("🤖 [vector-enhance] DeepSeek enhanced & cleaned: '%s' -> '%s'", originalQuery, cleanedQuery)
		return cleanedQuery, nil
	}

	log.Printf("🤖 [vector-enhance] DeepSeek returned same query, using fallback")

	return originalQuery, nil
}

// callDeepSeekAPIForVector เรียก DeepSeek API สำหรับ vector search โดยเฉพาะ
func (h *APIHandler) callDeepSeekAPIForVector(originalQuery string) (string, error) {
	// สร้าง prompt ที่ครอบคลุมสำหรับ vector search
	prompt := fmt.Sprintf(`
ตัวอย่างคำพ้องเสียง ถ้าข้อความเป็นไทย ใช้คำพร้องเสียงเป็นภาษาอังกฤษ ถ้าเป็นภาษาอังกฤษ ใช้คำพ้องเสียงเป็นภาษาไทย:
- toyota = โตโยต้า
- honda = ฮอนด้า
- nissan = นิสสัน
- mazda = มาสด้า
- brake = เบรค
- oil = น้ำมัน
- light = ไฟ
- wheel = ล้อ
- tire = ยาง
- battery = แบตเตอรี่
- coil = คอยล์
- shock = โช๊ค
- filter = กรอง
- engine = เครื่องยนต์
ถ้าเป็นชื่อรุ่น ใช้คำพ้องเสียงเป็นภาษาอังกฤษ และภาษาไทย

ใช้คำพ้องเสียงเหล่านี้เพื่อปรับปรุงการค้นหา vector โดยเฉพาะ
result=ข้อความต้นฉบับ + space + ข้อความพ้องเสียงที่เกี่ยวข้อง
แบ่งคำให้ด้วย space และลบคำซ้ำ
ข้อความต้นฉบับ: "%s"

return เฉพาะ result
`, originalQuery)

	// เรียกใช้ DeepSeek API พร้อม timeout ที่เพิ่มขึ้น
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

	client := &http.Client{Timeout: 30 * time.Second} // เพิ่ม timeout เป็น 30 วินาที
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ [vector-enhance] DeepSeek API timeout/error: %v", err)
		return "", fmt.Errorf("failed to call DeepSeek API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("❌ [vector-enhance] DeepSeek API status error: %d", resp.StatusCode)
		return "", fmt.Errorf("DeepSeek API error: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ [vector-enhance] Failed to read DeepSeek response: %v", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var deepSeekResp DeepSeekResponse
	if err := json.Unmarshal(body, &deepSeekResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(deepSeekResp.Choices) == 0 {
		return "", fmt.Errorf("no response from DeepSeek")
	}

	enhancedQuery := strings.TrimSpace(deepSeekResp.Choices[0].Message.Content)
	log.Printf("🤖 [vector-enhance] DeepSeek API success: '%s' -> '%s'", originalQuery, enhancedQuery)
	return enhancedQuery, nil
}

// removeDuplicateWords ลบคำซ้ำจาก query string
func (h *APIHandler) removeDuplicateWords(query string) string {
	// ลบเครื่องหมายคำพูดออก
	query = strings.Trim(query, `"'`)

	// แยกคำด้วย space
	words := strings.Fields(query)

	// ใช้ map เพื่อเก็บคำที่ไม่ซ้ำ (case insensitive)
	seen := make(map[string]bool)
	var result []string

	for _, word := range words {
		// ลบ special characters และทำให้เป็น lowercase สำหรับเปรียบเทียบ
		normalizedWord := strings.ToLower(strings.Trim(word, ".,!?;:()[]{}"))

		// ถ้ายังไม่เคยเห็นคำนี้
		if !seen[normalizedWord] && normalizedWord != "" {
			seen[normalizedWord] = true
			result = append(result, word) // เก็บคำเดิม (ไม่ใช่ normalized)
		}
	}

	finalResult := strings.Join(result, " ")
	log.Printf("🧹 [vector-enhance] Removed duplicates: '%s' -> '%s'", query, finalResult)
	return finalResult
}
*/
