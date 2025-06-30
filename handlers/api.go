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

	// AI Enhancement for Vector Search - ปรับปรุงคำค้นหาด้วย DeepSeek
	enhancedQuery, err := h.enhanceQueryForVectorSearch(query)
	if err != nil {
		log.Printf("⚠️ [VECTOR-SEARCH] DeepSeek enhancement failed, using original query: %v", err)
		enhancedQuery = query // ใช้คำค้นหาเดิม
	}

	// ใช้ enhanced query สำหรับการค้นหา
	searchQuery := enhancedQuery
	log.Printf("🤖 [VECTOR-SEARCH] Enhanced query: '%s' -> '%s'", query, searchQuery)

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
	fmt.Printf("   📝 Original Query: '%s'\n", query)
	fmt.Printf("   🤖 Enhanced Query: '%s'\n", searchQuery)
	fmt.Printf("   📊 Limit: %d, Offset: %d\n", limit, offset)
	fmt.Printf("   =====================================\n")
	ctx := c.Request.Context()

	// Step 1: Search Weaviate vector database first to get IC codes and barcodes
	if h.weaviateService == nil {
		// Fallback to regular search when Weaviate is not available
		log.Printf("⚠️ [VECTOR-SEARCH] Weaviate service not available, falling back to regular search")

		// Use regular PostgreSQL search as fallback
		searchResults, totalCount, err := h.postgreSQLService.SearchProducts(ctx, searchQuery, limit, offset)
		if err != nil {
			log.Printf("❌ [VECTOR-SEARCH] PostgreSQL fallback search failed: %v", err)
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Search failed: " + err.Error(),
			})
			return
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
			Query:      searchQuery + " (fallback to regular search)",
			Duration:   time.Since(startTime).Seconds() * 1000,
		}

		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Data:    results,
			Message: "Search completed successfully using fallback method (Weaviate unavailable)",
		})
		return
	}

	// Search vector database with higher limit to get more barcodes for better matching
	vectorLimit := limit * 3 // Get more results from vector DB to compensate for potential mismatches
	if vectorLimit > 300 {
		vectorLimit = 300
	}

	vectorProducts, err := h.weaviateService.SearchProducts(ctx, searchQuery, vectorLimit)
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
		Query:      searchQuery,
		Duration:   time.Since(startTime).Seconds() * 1000,
	}
	duration := time.Since(startTime).Seconds() * 1000

	// Enhanced search results logging
	fmt.Printf("\n🎯 [VECTOR-SEARCH] === SEARCH RESULTS SUMMARY ===\n")
	fmt.Printf("   📝 Original Query: '%s'\n", query)
	fmt.Printf("   🤖 Enhanced Query: '%s'\n", searchQuery)
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
