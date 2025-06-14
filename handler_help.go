package main

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

func handleHelp(c *gin.Context) {
	reqID := getNextRequestID()
	start := time.Now()

	printRequestDetails("GET", "/help", reqID, nil, nil)

	log.Printf("[handleHelp] 📚 GET /help - API documentation accessed")
	// Comprehensive API structure for AI frontend integration
	apiStructure := map[string]interface{}{
		"api_info": map[string]interface{}{
			"name":        "Enhanced Product Vector Search API",
			"version":     "1.2.0",
			"description": "Production-ready API with advanced vector search, multi-view image processing, real-time debugging, and comprehensive timeout protection",
			"base_url":    fmt.Sprintf("http://localhost:%s", config.ServerPort),
			"environment": "production-test",
			"capabilities": map[string]interface{}{
				"search": map[string]interface{}{
					"vector_similarity": true,
					"text_search":       true,
					"multilingual":      true,
					"thai_language":     true,
					"autocomplete":      false,
					"fuzzy_matching":    true,
					"pagination":        true,
					"real_time":         true,
				},
				"image_processing": map[string]interface{}{
					"multi_view":         true,
					"similarity_search":  true,
					"histogram_analysis": true,
					"color_features":     true,
					"supported_formats":  []string{"JPEG", "PNG", "GIF", "BMP", "WEBP"},
					"max_file_size_mb":   10,
					"vector_dimensions":  99,
				},
				"database": map[string]interface{}{
					"clickhouse_support": true,
					"sql_execution":      true,
					"real_time_queries":  true,
					"connection_pooling": true,
					"timeout_protection": true,
				},
				"debugging": map[string]interface{}{
					"request_tracing":     config.LogRequestResponse,
					"step_by_step":        config.LogStepByStep,
					"sql_logging":         config.LogSQLExecution,
					"performance_metrics": config.LogPerformance,
					"debug_levels":        5,
				},
			},
			"features": []string{
				"Advanced vector similarity search with TF-IDF",
				"Multi-language support (Thai, English)",
				"SQL command execution via GET/POST with base64 encoding",
				"Multi-view image upload and processing",
				"Image similarity search with configurable thresholds",
				"Real-time debug tracing and performance monitoring",
				"Comprehensive timeout protection on all operations",
				"Request/response logging with detailed metrics",
				"Connection pooling and resource management",
				"Health monitoring and status reporting",
			},
			"architecture": map[string]interface{}{
				"server_framework": "Go + Gin",
				"database":         "ClickHouse",
				"vector_engine":    "Custom TF-IDF implementation",
				"cache_system":     "In-memory with TTL",
				"concurrency":      fmt.Sprintf("%d workers", config.MaxWorkers),
			},
		}, "endpoints": map[string]interface{}{
			"GET /": map[string]interface{}{
				"description":   "Root endpoint - API status and basic information",
				"method":        "GET",
				"url":           fmt.Sprintf("http://localhost:%s/", config.ServerPort),
				"auth_required": false,
				"parameters":    map[string]interface{}{},
				"headers": map[string]interface{}{
					"required": []string{},
					"optional": []string{"User-Agent", "Accept"},
				},
				"response_schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"message": map[string]interface{}{
							"type":        "string",
							"description": "API status message",
							"example":     "Product Vector Search API - Enhanced Test Mode",
						},
						"status": map[string]interface{}{
							"type":        "string",
							"description": "Current running status",
							"enum":        []string{"running", "maintenance", "error"},
							"example":     "running",
						},
						"timestamp": map[string]interface{}{
							"type":        "string",
							"format":      "date-time",
							"description": "ISO 8601 timestamp",
							"example":     "2025-06-14T14:31:47+07:00",
						},
						"request_count": map[string]interface{}{
							"type":        "integer",
							"description": "Total requests processed since startup",
							"minimum":     0,
							"example":     1,
						},
					},
					"required": []string{"message", "status", "timestamp", "request_count"},
				},
				"example_response": map[string]interface{}{
					"message":       "Product Vector Search API - Enhanced Test Mode",
					"status":        "running",
					"timestamp":     "2025-06-14T14:31:47+07:00",
					"request_count": 1,
				},
				"performance": map[string]interface{}{
					"typical_response_time_ms": "< 5",
					"timeout_seconds":          config.RequestTimeout.Seconds(),
				},
			},
			"GET /health": map[string]interface{}{
				"description":   "Health check endpoint for monitoring and load balancing",
				"method":        "GET",
				"url":           fmt.Sprintf("http://localhost:%s/health", config.ServerPort),
				"auth_required": false,
				"parameters":    map[string]interface{}{},
				"response_schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"status": map[string]interface{}{
							"type":        "string",
							"description": "Service health status",
							"enum":        []string{"healthy", "degraded", "unhealthy"},
							"example":     "healthy",
						},
						"message": map[string]interface{}{
							"type":        "string",
							"description": "Health status message",
							"example":     "Vector search service is running",
						},
						"timestamp": map[string]interface{}{
							"type":        "string",
							"format":      "date-time",
							"description": "Health check timestamp",
							"example":     "2025-06-14T14:31:47+07:00",
						},
						"request_count": map[string]interface{}{
							"type":        "integer",
							"description": "Total health check requests",
							"minimum":     0,
							"example":     2,
						},
					},
					"required": []string{"status", "message", "timestamp", "request_count"},
				},
				"example_response": map[string]interface{}{
					"status":        "healthy",
					"message":       "Vector search service is running",
					"timestamp":     "2025-06-14T14:31:47+07:00",
					"request_count": 2,
				},
				"use_cases": []string{
					"Load balancer health checks",
					"Monitoring system integration",
					"Service availability verification",
					"Automated alerting systems",
				},
			},
			"POST /search": map[string]interface{}{
				"description":   "Advanced product search using vector similarity and text matching",
				"method":        "POST",
				"url":           fmt.Sprintf("http://localhost:%s/search", config.ServerPort),
				"content_type":  "application/json",
				"auth_required": false,
				"request_schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "Search query text (supports Thai and English)",
							"minLength":   1,
							"maxLength":   1000,
							"example":     "ปรอท thermometer temperature",
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"description": "Maximum number of results to return",
							"minimum":     1,
							"maximum":     100,
							"default":     30,
							"example":     10,
						},
						"offset": map[string]interface{}{
							"type":        "integer",
							"description": "Number of results to skip for pagination",
							"minimum":     0,
							"default":     0,
							"example":     0,
						},
					},
					"required": []string{"query"},
				},
				"response_schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"total_count": map[string]interface{}{
							"type":        "integer",
							"description": "Total number of matching results",
							"minimum":     0,
						},
						"data": map[string]interface{}{
							"type":        "array",
							"description": "Array of matching products",
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"barcode": map[string]interface{}{
										"type":        "string",
										"description": "Product barcode identifier",
									},
									"name": map[string]interface{}{
										"type":        "string",
										"description": "Product name",
									},
									"description": map[string]interface{}{
										"type":        "string",
										"description": "Product description",
									},
									"similarity_score": map[string]interface{}{
										"type":        "number",
										"description": "Similarity score (0.0 to 1.0)",
										"minimum":     0,
										"maximum":     1,
									},
									"balance_qty": map[string]interface{}{
										"type":        "integer",
										"description": "Available quantity",
										"minimum":     0,
									},
								},
							},
						},
						"offset": map[string]interface{}{
							"type":        "integer",
							"description": "Current pagination offset",
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"description": "Current results limit",
						},
					},
					"required": []string{"total_count", "data", "offset", "limit"},
				},
				"example_request": map[string]interface{}{
					"query":  "brake pad ผ้าเบรค",
					"limit":  5,
					"offset": 0,
				},
				"example_response": map[string]interface{}{
					"total_count": 2,
					"data": []map[string]interface{}{
						{
							"barcode":          "123456789",
							"name":             "Premium Brake Pad Set",
							"description":      "High quality brake pads for sedan vehicles",
							"similarity_score": 0.95,
							"balance_qty":      100,
						},
					},
					"offset": 0,
					"limit":  5,
				},
				"performance": map[string]interface{}{
					"typical_response_time_ms": "< 50",
					"timeout_seconds":          config.RequestTimeout.Seconds(),
				},
				"features": []string{
					"Vector similarity matching",
					"Multi-language support (Thai/English)",
					"Fuzzy text matching",
					"Pagination support",
					"Configurable result limits",
				},
			}, "GET /commandget": map[string]interface{}{
				"description":   "Execute SQL commands via GET method with base64 encoded queries",
				"method":        "GET",
				"url":           fmt.Sprintf("http://localhost:%s/commandget", config.ServerPort),
				"auth_required": false,
				"parameters": map[string]interface{}{
					"q": map[string]interface{}{
						"type":        "string",
						"description": "Base64 encoded SQL query",
						"required":    true,
						"encoding":    "base64",
						"example":     "U0VMRUNUIDE=", // SELECT 1
					},
				},
				"query_examples": map[string]interface{}{
					"simple_select": map[string]interface{}{
						"sql":      "SELECT 1",
						"base64":   "U0VMRUNUIDE=",
						"full_url": fmt.Sprintf("http://localhost:%s/commandget?q=U0VMRUNUIDE=", config.ServerPort),
					},
					"current_time": map[string]interface{}{
						"sql":      "SELECT now() as current_time",
						"base64":   "U0VMRUNUIG5vdygpIGFzIGN1cnJlbnRfdGltZQ==",
						"full_url": fmt.Sprintf("http://localhost:%s/commandget?q=U0VMRUNUIG5vdygpIGFzIGN1cnJlbnRfdGltZQ==", config.ServerPort),
					},
				},
				"response_schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"result": map[string]interface{}{
							"type":        "array|object",
							"description": "Query execution results",
						},
						"command": map[string]interface{}{
							"type":        "string",
							"description": "Original SQL command",
						},
						"decoded_sql": map[string]interface{}{
							"type":        "string",
							"description": "Decoded SQL query for verification",
						},
						"method": map[string]interface{}{
							"type":        "string",
							"description": "HTTP method used",
							"enum":        []string{"GET"},
						},
					},
					"required": []string{"result", "command", "method"},
				},
				"performance": map[string]interface{}{
					"typical_response_time_ms": "< 100",
					"timeout_seconds":          config.SQLTimeout.Seconds(),
				},
				"use_cases": []string{
					"Quick SQL queries via URL",
					"Database health checks",
					"Simple data retrieval",
					"Integration with monitoring tools",
				},
			},
			"POST /commandpost": map[string]interface{}{
				"description":   "Execute SQL commands via POST method with JSON payload",
				"method":        "POST",
				"url":           fmt.Sprintf("http://localhost:%s/commandpost", config.ServerPort),
				"content_type":  "application/json",
				"auth_required": false,
				"request_schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query_base64": map[string]interface{}{
							"type":        "string",
							"description": "Base64 encoded SQL query",
							"encoding":    "base64",
							"example":     "U0VMRUNUIDE=",
						},
					},
					"required": []string{"query_base64"},
				},
				"response_schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"result": map[string]interface{}{
							"type":        "array|object",
							"description": "Query execution results",
						},
						"command": map[string]interface{}{
							"type":        "string",
							"description": "Original SQL command",
						},
						"decoded_sql": map[string]interface{}{
							"type":        "string",
							"description": "Decoded SQL query for verification",
						},
						"method": map[string]interface{}{
							"type":        "string",
							"description": "HTTP method used",
							"enum":        []string{"POST"},
						},
					},
					"required": []string{"result", "command", "method"},
				},
				"example_request": map[string]interface{}{
					"query_base64": "U0VMRUNUIG5vdygpIGFzIGN1cnJlbnRfdGltZQ==",
				},
				"example_response": map[string]interface{}{
					"result": []map[string]interface{}{
						{"current_time": "2025-06-14T09:44:41Z"},
					},
					"command":     "SELECT now() as current_time",
					"decoded_sql": "SELECT now() as current_time",
					"method":      "POST",
				},
				"performance": map[string]interface{}{
					"typical_response_time_ms": "< 100",
					"timeout_seconds":          config.SQLTimeout.Seconds(),
				},
				"use_cases": []string{
					"Complex SQL queries",
					"Data analysis operations",
					"Bulk data operations",
					"Advanced database interactions",
				},
			}, "POST /imgupload": map[string]interface{}{
				"description":   "Upload and process product images with advanced vector generation",
				"method":        "POST",
				"url":           fmt.Sprintf("http://localhost:%s/imgupload", config.ServerPort),
				"content_type":  "application/json",
				"auth_required": false,
				"request_schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"barcode": map[string]interface{}{
							"type":        "string",
							"description": "Unique product barcode identifier",
							"minLength":   3,
							"maxLength":   50,
							"pattern":     "^[A-Za-z0-9-_]+$",
							"example":     "ABC123456789",
						},
						"imagenumber": map[string]interface{}{
							"type":        "integer",
							"description": "Sequential image number for the product",
							"minimum":     1,
							"maximum":     999,
							"default":     1,
							"example":     1,
						},
						"image_data": map[string]interface{}{
							"type":        "string",
							"description": "Base64 encoded image data",
							"format":      "base64",
							"maxLength":   15000000, // ~10MB in base64
							"example":     "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChAHI3+o3XgAAAABJRU5ErkJggg==",
						},
						"use_multi_view": map[string]interface{}{
							"type":        "boolean",
							"description": "Enable multi-view processing for enhanced accuracy",
							"default":     false,
							"example":     true,
						},
					},
					"required": []string{"barcode", "image_data"},
				},
				"response_schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"status": map[string]interface{}{
							"type":        "string",
							"description": "Processing status",
							"enum":        []string{"success", "error", "warning"},
						},
						"message": map[string]interface{}{
							"type":        "string",
							"description": "Processing status message",
						},
						"barcode": map[string]interface{}{
							"type":        "string",
							"description": "Product barcode from request",
						},
						"imagenumber": map[string]interface{}{
							"type":        "integer",
							"description": "Image number from request",
						},
						"total_views_generated": map[string]interface{}{
							"type":        "integer",
							"description": "Number of views generated (multi-view only)",
							"minimum":     0,
						},
						"total_vectors_stored": map[string]interface{}{
							"type":        "integer",
							"description": "Number of vectors stored in database",
							"minimum":     0,
						},
						"vector_size": map[string]interface{}{
							"type":        "integer",
							"description": "Dimension size of generated vectors",
							"example":     99,
						},
						"processing_time_ms": map[string]interface{}{
							"type":        "number",
							"description": "Processing duration in milliseconds",
							"minimum":     0,
						},
					},
					"required": []string{"status", "message", "barcode", "imagenumber", "processing_time_ms"},
				},
				"example_request": map[string]interface{}{
					"barcode":        "XYZ987654",
					"imagenumber":    2,
					"image_data":     "(base64 encoded image string)",
					"use_multi_view": true,
				},
				"example_response": map[string]interface{}{
					"status":                "success",
					"message":               "Image processed and vectors stored successfully with multi-view.",
					"barcode":               "XYZ987654",
					"imagenumber":           2,
					"total_views_generated": 5,
					"total_vectors_stored":  5,
					"vector_size":           99,
					"processing_time_ms":    150.75,
				},
				"performance": map[string]interface{}{
					"typical_response_time_ms": "< 200 (depending on image size and multi-view)",
					"timeout_seconds":          config.RequestTimeout.Seconds(),
				},
				"use_cases": []string{
					"Adding new product images to the database",
					"Updating existing product images",
					"Batch image processing for catalog ingestion",
					"Generating image vectors for similarity search",
				},
			},
			"POST /imgsearch": map[string]interface{}{
				"description":   "Search for similar images using a query image and vector similarity",
				"method":        "POST",
				"url":           fmt.Sprintf("http://localhost:%s/imgsearch", config.ServerPort),
				"content_type":  "application/json",
				"auth_required": false,
				"request_schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"image_data": map[string]interface{}{
							"type":        "string",
							"description": "Base64 encoded query image data",
							"format":      "base64",
							"maxLength":   15000000, // ~10MB in base64
							"example":     "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChAHI3+o3XgAAAABJRU5ErkJggg==",
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"description": "Maximum number of similar images to return",
							"minimum":     1,
							"maximum":     50,
							"default":     10,
							"example":     5,
						},
						"similarity_threshold": map[string]interface{}{
							"type":        "number",
							"description": "Minimum similarity score for results (0.0 to 1.0)",
							"minimum":     0.0,
							"maximum":     1.0,
							"default":     0.75,
							"example":     0.8,
						},
						"use_multi_view_query": map[string]interface{}{
							"type":        "boolean",
							"description": "Process query image with multi-view for potentially better results",
							"default":     false,
							"example":     true,
						},
					},
					"required": []string{"image_data"},
				},
				"response_schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"total_count": map[string]interface{}{
							"type":        "integer",
							"description": "Total number of matching similar images found",
							"minimum":     0,
						},
						"results": map[string]interface{}{
							"type":        "array",
							"description": "Array of similar images",
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"barcode": map[string]interface{}{
										"type":        "string",
										"description": "Barcode of the similar product",
									},
									"imagenumber": map[string]interface{}{
										"type":        "integer",
										"description": "Image number of the similar product image",
									},
									"similarity_score": map[string]interface{}{
										"type":        "number",
										"description": "Similarity score (0.0 to 1.0)",
										"minimum":     0.0,
										"maximum":     1.0,
									},
									"image_url": map[string]interface{}{ // Assuming a way to retrieve/link images
										"type":        "string",
										"description": "URL or path to the similar image (if applicable)",
										"example":     "/images/DEF456/1.jpg",
									},
								},
								"required": []string{"barcode", "imagenumber", "similarity_score"},
							},
						},
						"query_vector_size": map[string]interface{}{
							"type":        "integer",
							"description": "Dimension size of the query image vector(s)",
							"example":     99,
						},
						"processing_time_ms": map[string]interface{}{
							"type":        "number",
							"description": "Total search duration in milliseconds",
							"minimum":     0,
						},
					},
					"required": []string{"total_count", "results", "processing_time_ms"},
				},
				"example_request": map[string]interface{}{
					"image_data":           "(base64 encoded query image string)",
					"limit":                3,
					"similarity_threshold": 0.85,
					"use_multi_view_query": false,
				},
				"example_response": map[string]interface{}{
					"total_count": 1,
					"results": []map[string]interface{}{
						{
							"barcode":          "DEF456789",
							"imagenumber":      1,
							"similarity_score": 0.88,
							"image_url":        "/images/DEF456789/1.jpg",
						},
					},
					"query_vector_size":  99,
					"processing_time_ms": 75.2,
				},
				"performance": map[string]interface{}{
					"typical_response_time_ms": "< 150 (depending on query image and DB size)",
					"timeout_seconds":          config.RequestTimeout.Seconds(),
				},
				"use_cases": []string{
					"Finding visually similar products",
					"Reverse image search for product identification",
					"Powering 'shop the look' features",
					"Identifying duplicate or near-duplicate images",
				},
			},
		},
		"debug_endpoints": map[string]interface{}{
			"GET /debug/trace": map[string]interface{}{
				"description": "Get debug trace information (only available when debug mode is enabled)",
				"method":      "GET",
				"parameters": map[string]interface{}{
					"request_id": "number (optional) - specific request ID to trace",
				},
				"response": map[string]interface{}{
					"total_traces": "number - total available traces",
					"traces":       "object - trace data by request ID",
				},
				"availability": fmt.Sprintf("Debug mode: %t", config.DebugMode),
			},
			"GET /help": map[string]interface{}{
				"description": "This endpoint - API documentation and structure",
				"method":      "GET",
				"parameters":  "none",
				"response":    "object - complete API documentation",
			},
		},
		"configuration": map[string]interface{}{
			"timeouts": map[string]interface{}{
				"request_timeout_seconds":  config.RequestTimeout.Seconds(),
				"sql_timeout_seconds":      config.SQLTimeout.Seconds(),
				"database_timeout_seconds": config.DatabaseTimeout.Seconds(),
				"http_timeout_seconds":     config.HttpTimeout.Seconds(),
			},
			"debug_settings": map[string]interface{}{
				"debug_mode":               config.DebugMode,
				"debug_level":              config.DebugLevel,
				"step_by_step_logging":     config.LogStepByStep,
				"sql_execution_logging":    config.LogSQLExecution,
				"request_response_logging": config.LogRequestResponse,
				"performance_logging":      config.LogPerformance,
			},
			"server_settings": map[string]interface{}{
				"port":                 config.ServerPort,
				"max_workers":          config.MaxWorkers,
				"cache_enabled":        config.CacheEnabled,
				"cache_ttl_minutes":    config.CacheTTL.Minutes(),
				"similarity_threshold": config.SimilarityThreshold,
			},
		}, "integration_guide": map[string]interface{}{
			"base64_encoding": map[string]interface{}{
				"description": "SQL queries must be base64 encoded for security and URL safety",
				"implementation_examples": map[string]interface{}{
					"javascript": map[string]interface{}{
						"function": "btoa('SELECT 1')",
						"example":  "const encoded = btoa('SELECT now()'); // U0VMRUNUIG5vdygp",
						"decode":   "atob('U0VMRUNUIDE=')", // To decode back
					},
					"python": map[string]interface{}{
						"function": "base64.b64encode('SELECT 1'.encode()).decode()",
						"example":  "import base64; encoded = base64.b64encode('SELECT now()'.encode()).decode()",
						"decode":   "base64.b64decode(encoded).decode()",
					},
					"curl": map[string]interface{}{
						"function": "echo 'SELECT 1' | base64",
						"example":  "echo 'SELECT now()' | base64 | tr -d '\n'",
						"usage":    "curl \"http://localhost:8008/commandget?q=$(echo 'SELECT 1' | base64)\"",
					},
					"php": map[string]interface{}{
						"function": "base64_encode('SELECT 1')",
						"example":  "$encoded = base64_encode('SELECT now()');",
						"decode":   "base64_decode($encoded)",
					},
					"go": map[string]interface{}{
						"function": "base64.StdEncoding.EncodeToString([]byte('SELECT 1'))",
						"example":  "encoded := base64.StdEncoding.EncodeToString([]byte('SELECT now()'))",
						"decode":   "decoded, _ := base64.StdEncoding.DecodeString(encoded)",
					},
				},
			},
			"image_processing": map[string]interface{}{
				"description":            "Images must be base64 encoded for upload and search operations",
				"supported_formats":      []string{"JPEG", "PNG", "GIF", "BMP", "WEBP"},
				"max_file_size_mb":       10,
				"recommended_resolution": "800x600 to 1920x1080",
				"encoding_examples": map[string]interface{}{
					"javascript": "const reader = new FileReader(); reader.readAsDataURL(file);",
					"python":     "import base64; with open('image.jpg', 'rb') as f: encoded = base64.b64encode(f.read()).decode()",
					"curl":       "base64 -i image.jpg | tr -d '\\n'",
				},
				"multi_view_benefits": []string{
					"Enhanced accuracy with multiple viewing angles",
					"Rotation and scale invariance",
					"Better feature extraction",
					"Improved similarity matching",
					"Robust against lighting variations",
				},
				"single_vs_multi_view": map[string]interface{}{
					"single_view": map[string]interface{}{
						"processing_time": "< 1ms",
						"accuracy":        "Standard",
						"use_case":        "Quick matching, simple products",
					},
					"multi_view": map[string]interface{}{
						"processing_time": "< 5ms",
						"accuracy":        "Enhanced",
						"use_case":        "Complex products, high accuracy needs",
						"views_generated": 5,
					},
				},
			},
			"error_handling": map[string]interface{}{
				"http_status_codes": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Success - Request completed successfully",
						"action":      "Process the response data",
					},
					"400": map[string]interface{}{
						"description": "Bad Request - Invalid input parameters or malformed request",
						"action":      "Check request format, parameters, and encoding",
						"common_causes": []string{
							"Missing required parameters",
							"Invalid base64 encoding",
							"Malformed JSON",
							"Invalid data types",
						},
					},
					"404": map[string]interface{}{
						"description": "Not Found - Endpoint or resource does not exist",
						"action":      "Verify URL and endpoint path",
						"note":        "Check available endpoints in this documentation",
					},
					"408": map[string]interface{}{
						"description": "Request Timeout - Operation exceeded configured timeout limit",
						"action":      "Retry with simpler query or check system load",
						"timeouts": map[string]interface{}{
							"request_timeout": fmt.Sprintf("%.0f seconds", config.RequestTimeout.Seconds()),
							"sql_timeout":     fmt.Sprintf("%.0f seconds", config.SQLTimeout.Seconds()),
						},
					},
					"500": map[string]interface{}{
						"description": "Internal Server Error - Unexpected server processing error",
						"action":      "Check server logs, retry after delay, contact support if persistent",
					},
				},
				"retry_strategy": map[string]interface{}{
					"recommended_approach": "Exponential backoff with jitter",
					"max_retries":          3,
					"initial_delay_ms":     100,
					"max_delay_ms":         5000,
					"timeout_handling":     "Reduce query complexity or increase client timeout",
				},
			},
			"best_practices": map[string]interface{}{
				"performance": []string{
					"Use appropriate pagination limits (10-50 results per request)",
					"Enable multi-view only when high accuracy is required",
					"Cache frequently accessed data on client side",
					"Monitor response times and adjust timeouts accordingly",
					"Use connection pooling for high-volume applications",
				},
				"reliability": []string{
					"Always handle timeout scenarios (408 responses)",
					"Implement proper error handling for all status codes",
					"Use exponential backoff for retry logic",
					"Validate input data before sending requests",
					"Monitor API health via /health endpoint",
				},
				"security": []string{
					"Validate and sanitize all input data",
					"Use HTTPS in production environments",
					"Implement rate limiting on client side",
					"Log and monitor API usage patterns",
					"Keep base64 encoded data within reasonable size limits",
				},
				"integration": []string{
					"Start with /health endpoint to verify connectivity",
					"Use /help endpoint to get latest API documentation",
					"Enable debug mode during development for detailed logging",
					"Test with small datasets before scaling up",
					"Monitor debug traces for troubleshooting (/debug/trace)",
				},
			}, "code_examples": map[string]interface{}{
				"javascript_fetch": "// Product Search Example\nconst searchProducts = async (query) => {\n  try {\n    const response = await fetch('http://localhost:8008/search', {\n      method: 'POST',\n      headers: { 'Content-Type': 'application/json' },\n      body: JSON.stringify({ query, limit: 10, offset: 0 })\n    });\n    const data = await response.json();\n    return data;\n  } catch (error) {\n    console.error('Search failed:', error);\n  }\n};\n\n// SQL Command Example\nconst executeSQL = async (sqlQuery) => {\n  const encoded = btoa(sqlQuery);\n  const response = await fetch(`http://localhost:8008/commandget?q=${encoded}`);\n  return await response.json();\n};",
				"python_requests":  "import requests\nimport base64\nimport json\n\n# Product Search Example\ndef search_products(query, limit=10, offset=0):\n    url = \"http://localhost:8008/search\"\n    payload = {\"query\": query, \"limit\": limit, \"offset\": offset}\n    response = requests.post(url, json=payload)\n    return response.json()\n\n# SQL Command Example\ndef execute_sql(sql_query):\n    encoded = base64.b64encode(sql_query.encode()).decode()\n    url = f\"http://localhost:8008/commandget?q={encoded}\"\n    response = requests.get(url)\n    return response.json()\n\n# Image Upload Example\ndef upload_image(barcode, image_path, multi_view=True):\n    with open(image_path, 'rb') as f:\n        image_data = base64.b64encode(f.read()).decode()\n    \n    payload = {\n        \"barcode\": barcode,\n        \"image_data\": image_data,\n        \"use_multi_view\": multi_view\n    }\n    response = requests.post(\"http://localhost:8008/imgupload\", json=payload)\n    return response.json()",
				"curl_examples":    "# Health Check\ncurl -X GET \"http://localhost:8008/health\"\n\n# Product Search\ncurl -X POST \"http://localhost:8008/search\" \\\n  -H \"Content-Type: application/json\" \\\n  -d '{\"query\": \"brake pad\", \"limit\": 5}'\n\n# SQL Command (GET)\ncurl -X GET \"http://localhost:8008/commandget?q=$(echo 'SELECT 1' | base64)\"\n\n# SQL Command (POST)\ncurl -X POST \"http://localhost:8008/commandpost\" \\\n  -H \"Content-Type: application/json\" \\\n  -d '{\"query_base64\": \"'$(echo 'SELECT now()' | base64)'\"}'",
			},
		},
	}

	duration := time.Since(start).Seconds() * 1000
	printResponseDetails(reqID, 200, apiStructure, duration)

	if stats != nil {
		atomic.AddInt64(&stats.TotalRequests, 1)
	}

	c.JSON(200, apiStructure)
}
