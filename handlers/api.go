package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"smlgoapi/config"
	"smlgoapi/models"
	"smlgoapi/services"

	"github.com/gin-gonic/gin"
)

// HTTP status messages
const (
	msgInvalidJSON   = "Invalid JSON body"
	msgServiceError  = "Service error"
	msgDatabaseError = "Database error"
	msgSuccess       = "Operation completed successfully"
	msgHealthy       = "healthy"
	msgUnavailable   = "unavailable"
	msgConnected     = "connected"
)

type APIHandler struct {
	postgreSQLService *services.PostgreSQLService
	clickHouseService *services.ClickHouseService
	emailService      *services.EmailService
}

func NewAPIHandler(postgreSQLService *services.PostgreSQLService) *APIHandler {
	cfg := config.LoadConfig()

	// Initialize ClickHouse service with config
	var clickHouseService *services.ClickHouseService
	chs, err := services.NewClickHouseService(cfg)
	if err != nil {
		log.Printf("⚠️ Failed to initialize ClickHouse service: %v", err)
		clickHouseService = nil
	} else {
		clickHouseService = chs
	}

	// Initialize Email service
	emailService := services.NewEmailService(cfg.Email.APIKey)

	return &APIHandler{
		postgreSQLService: postgreSQLService,
		clickHouseService: clickHouseService,
		emailService:      emailService,
	}
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the API service is healthy and database connection is working
// @Tags health
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Router /health [get]
func (h *APIHandler) HealthCheck(c *gin.Context) {
	// Test database connection
	dbStatus := msgConnected
	if db, err := h.postgreSQLService.GetDatabaseConnection("postgres"); err != nil {
		dbStatus = msgUnavailable
		log.Printf("❌ Health check: Database connection failed: %v", err)
	} else {
		defer db.Close()
		// Try a simple query
		if err := db.Ping(); err != nil {
			dbStatus = msgUnavailable
			log.Printf("❌ Health check: Database ping failed: %v", err)
		}
	}

	response := models.HealthResponse{
		Status:    msgHealthy,
		Timestamp: time.Now(),
		Database:  dbStatus,
	}

	log.Printf("✅ Health check completed: database=%s", dbStatus)
	c.JSON(http.StatusOK, response)
}

// GuideEndpoint godoc
// @Summary API Guide
// @Description Get comprehensive API documentation with all available endpoints
// @Tags documentation
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /guide [get]
func (h *APIHandler) GuideEndpoint(c *gin.Context) {
	baseURL := fmt.Sprintf("http://%s", c.Request.Host)

	guide := map[string]interface{}{
		"api_name":       "SMLGOAPI",
		"version":        "1.0.0",
		"description":    "Smart Machine Learning Go API - Database and Vector Search Operations",
		"documentation":  "Complete API Guide with Examples",
		"base_url":       baseURL,
		"content_type":   "application/json",
		"authentication": "None required for current endpoints",

		"endpoints": map[string]interface{}{
			"health_check": map[string]interface{}{
				"url":         fmt.Sprintf("%s/health", baseURL),
				"method":      "GET",
				"description": "Check API and database health status",
			},
			"api_guide": map[string]interface{}{
				"url":         fmt.Sprintf("%s/v1/guide", baseURL),
				"method":      "GET",
				"description": "Get this comprehensive API documentation",
			},

			// PostgreSQL Database Operations
			"postgresql": map[string]interface{}{
				"command": map[string]interface{}{
					"url":         fmt.Sprintf("%s/v1/pgcommand", baseURL),
					"method":      "POST",
					"description": "Execute PostgreSQL commands (INSERT, UPDATE, DELETE, CREATE, etc.)",
					"body_example": map[string]interface{}{
						"database_name": "your_database",
						"query":         "INSERT INTO users (name, email) VALUES ('John', 'john@example.com')",
					},
				},
				"select": map[string]interface{}{
					"url":         fmt.Sprintf("%s/v1/pgselect", baseURL),
					"method":      "POST",
					"description": "Execute PostgreSQL SELECT queries and return data",
					"body_example": map[string]interface{}{
						"database_name": "your_database",
						"query":         "SELECT * FROM users WHERE active = true LIMIT 10",
					},
				},
				"select_group": map[string]interface{}{
					"url":         fmt.Sprintf("%s/v1/pgselectgroup", baseURL),
					"method":      "POST",
					"description": "Execute multiple PostgreSQL SELECT queries in a single request",
					"body_example": map[string]interface{}{
						"database_name": "your_database",
						"queries": []string{
							"SELECT COUNT(*) as total_users FROM users",
							"SELECT * FROM users WHERE created_at >= NOW() - INTERVAL '7 days'",
							"SELECT status, COUNT(*) as count FROM orders GROUP BY status",
						},
					},
				},
				"check_database": map[string]interface{}{
					"url":         fmt.Sprintf("%s/v1/pgcheckdatabase", baseURL),
					"method":      "POST",
					"description": "Check if PostgreSQL database exists and get table list",
					"body_example": map[string]interface{}{
						"database_name": "your_database",
					},
				},
				"check_table": map[string]interface{}{
					"url":         fmt.Sprintf("%s/v1/pgchecktable", baseURL),
					"method":      "POST",
					"description": "Check if PostgreSQL table exists in database",
					"body_example": map[string]interface{}{
						"database_name": "your_database",
						"table_name":    "users",
					},
				},
			},

			// ClickHouse Database Operations
			"clickhouse": map[string]interface{}{
				"command": map[string]interface{}{
					"url":         fmt.Sprintf("%s/v1/chcommand", baseURL),
					"method":      "POST",
					"description": "Execute ClickHouse commands (INSERT, CREATE, ALTER, etc.)",
					"body_example": map[string]interface{}{
						"database_name": "analytics",
						"query":         "INSERT INTO events (user_id, event_type, timestamp) VALUES (123, 'click', now())",
					},
				},
				"select": map[string]interface{}{
					"url":         fmt.Sprintf("%s/v1/chselect", baseURL),
					"method":      "POST",
					"description": "Execute ClickHouse SELECT queries for analytics",
					"body_example": map[string]interface{}{
						"database_name": "analytics",
						"query":         "SELECT event_type, count() as events FROM events WHERE date >= today()-7 GROUP BY event_type",
					},
				},
				"select_group": map[string]interface{}{
					"url":         fmt.Sprintf("%s/v1/chselectgroup", baseURL),
					"method":      "POST",
					"description": "Execute multiple ClickHouse SELECT queries for complex analytics",
					"body_example": map[string]interface{}{
						"database_name": "analytics",
						"queries": []string{
							"SELECT count() as total_events FROM events WHERE date = today()",
							"SELECT user_id, count() as sessions FROM events GROUP BY user_id LIMIT 10",
							"SELECT hour, count() as events FROM events WHERE date = today() GROUP BY toHour(timestamp)",
						},
					},
				},
				"check_database": map[string]interface{}{
					"url":         fmt.Sprintf("%s/v1/chcheckdatabase", baseURL),
					"method":      "POST",
					"description": "Check if ClickHouse database exists and get table list",
					"body_example": map[string]interface{}{
						"database_name": "analytics",
					},
				},
				"check_table": map[string]interface{}{
					"url":         fmt.Sprintf("%s/v1/chchecktable", baseURL),
					"method":      "POST",
					"description": "Check if ClickHouse table exists in database",
					"body_example": map[string]interface{}{
						"database_name": "analytics",
						"table_name":    "events",
					},
				},
			},
		},

		"response_format": map[string]interface{}{
			"success_response": map[string]interface{}{
				"success":  true,
				"message":  "Operation completed successfully",
				"data":     "{}",
				"duration": "123.45",
			},
			"error_response": map[string]interface{}{
				"success": false,
				"error":   "Error description",
				"message": "Optional error message",
			},
		},

		"database_support": map[string]interface{}{
			"postgresql": map[string]string{
				"description": "Full CRUD operations with transaction support",
				"use_cases":   "OLTP, relational data, complex queries",
			},
			"clickhouse": map[string]string{
				"description": "High-performance analytics and aggregations",
				"use_cases":   "OLAP, time-series data, real-time analytics",
			},
		},

		"examples": map[string]interface{}{
			"curl_examples": map[string]interface{}{
				"health_check": fmt.Sprintf("curl -X GET %s/health", baseURL),
				"postgresql_select": fmt.Sprintf(`curl -X POST %s/v1/pgselect \
  -H "Content-Type: application/json" \
  -d '{"database_name": "postgres", "query": "SELECT version()"}'`, baseURL),
				"clickhouse_analytics": fmt.Sprintf(`curl -X POST %s/v1/chselect \
  -H "Content-Type: application/json" \
  -d '{"database_name": "default", "query": "SELECT count() FROM system.tables"}'`, baseURL),
			},
		},

		"notes": []string{
			"All POST endpoints require Content-Type: application/json",
			"PostgreSQL endpoints support multiple databases",
			"ClickHouse endpoints optimized for analytical queries",
			"Check database/table endpoints return existence status and metadata",
			"Group select endpoints allow batch query execution",
			"All responses include execution duration in milliseconds",
		},

		"generated_at": time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, guide)
}
