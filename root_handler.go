package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RootHandler handles the root API documentation endpoint
func RootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":     "SMLGOAPI - ClickHouse REST API",
		"version":     "1.0.0",
		"api_version": "v1",
		"endpoints": gin.H{
			// Core endpoints
			"health": "GET /health",
			// API v1 endpoints (recommended)
			"v1_provinces":        "POST /v1/provinces",
			"v1_amphures":         "POST /v1/amphures",
			"v1_tambons":          "POST /v1/tambons",
			"v1_findbyzipcode":    "POST /v1/findbyzipcode",
			"v1_search_by_vector": "POST /v1/search-by-vector",
			"v1_command":          "POST /v1/command",
			"v1_select":           "POST /v1/select",
			"v1_pgcommand":        "POST /v1/pgcommand",
			"v1_pgselect":         "POST /v1/pgselect",
			"v1_tables":           "GET /v1/tables",

			// Legacy endpoints (backwards compatibility)
			"provinces":     "POST /get/provinces",
			"amphures":      "POST /get/amphures",
			"tambons":       "POST /get/tambons",
			"findbyzipcode": "POST /get/findbyzipcode",
			"command":       "POST /command",
			"select":        "POST /select",
			"pgcommand":     "POST /pgcommand",
			"pgselect":      "POST /pgselect",
			"tables":        "GET /api/tables",
			"imgproxy":      "GET /imgproxy?url=<image_url>",
		},
		"documentation":  "Use /v1/ endpoints for new integrations. Legacy endpoints maintained for backwards compatibility.",
		"migration_note": "Please migrate to /v1/ endpoints. Legacy endpoints will be deprecated in future versions.",
	})
}
