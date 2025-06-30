package main

import (
	"smlgoapi/handlers"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// setupRouter configures and returns the main Gin router with all endpoints
func setupRouter(apiHandler *handlers.APIHandler) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery()) // CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // In production, specify your frontend domain
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API documentation endpoint (root)
	router.GET("/", RootHandler)

	// All API endpoints under /v1
	v1 := router.Group("/v1")
	{
		// Health check endpoint
		v1.GET("/health", apiHandler.HealthCheck)

		// API documentation endpoints
		v1.GET("/docs", DocsHandler)
		v1.GET("/guide", apiHandler.GuideEndpoint)

		// Search endpoints
		v1.GET("/search-by-vector", apiHandler.SearchProductsByVector)
		v1.POST("/search-by-vector", apiHandler.SearchProductsByVector)

		// Database endpoints
		v1.GET("/tables", apiHandler.GetTables)
		v1.POST("/command", apiHandler.CommandEndpoint)
		v1.POST("/select", apiHandler.SelectEndpoint)
		v1.POST("/pgcommand", apiHandler.PgCommandEndpoint)
		v1.POST("/pgselect", apiHandler.PgSelectEndpoint)

		// Thai Administrative Data endpoints
		v1.POST("/provinces", apiHandler.GetProvinces)
		v1.POST("/amphures", apiHandler.GetAmphures)
		v1.POST("/tambons", apiHandler.GetTambons)
		v1.POST("/findbyzipcode", apiHandler.FindByZipCode)
	}

	return router
}
