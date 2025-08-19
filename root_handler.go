package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RootHandler handles the root API documentation endpoint
func RootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":     "SMLGOAPI - PostgreSQL REST API",
		"version":     "1.0.0",
		"api_version": "v1",
		"endpoints": gin.H{
			// Core endpoints
			"health": "GET /health",
			// API v1 endpoints (recommended)
			"v1_search_by_vector": "POST /v1/search-by-vector",
			"v1_pgcommand":        "POST /v1/pgcommand",
			"v1_pgselect":         "POST /v1/pgselect",
			"v1_pgselectgroup":    "POST /v1/pgselectgroup",
			"v1_pgcheckdatabase":  "POST /v1/pgcheckdatabase",
			"v1_pgchecktable":     "POST /v1/pgchecktable",

			// Legacy endpoints (backwards compatibility)
			"pgcommand": "POST /pgcommand",
			"pgselect":  "POST /pgselect",
		},
		"documentation":  "Use /v1/ endpoints for new integrations. Legacy endpoints maintained for backwards compatibility.",
		"migration_note": "Please migrate to /v1/ endpoints. Legacy endpoints will be deprecated in future versions.",
	})
}
