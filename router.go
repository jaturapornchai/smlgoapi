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

	// Health check endpoint
	router.GET("/health", apiHandler.HealthCheck)

	// Legacy endpoints (maintain backwards compatibility)
	router.POST("/search", apiHandler.SearchProducts)
	router.GET("/imgproxy", apiHandler.ImageProxy)
	router.HEAD("/imgproxy", apiHandler.ImageProxyHead)
	router.POST("/command", apiHandler.CommandEndpoint)
	router.POST("/select", apiHandler.SelectEndpoint)
	router.POST("/pgcommand", apiHandler.PgCommandEndpoint)
	router.POST("/pgselect", apiHandler.PgSelectEndpoint)
	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Health check endpoint for v1
		v1.GET("/health", apiHandler.HealthCheck)

		// API documentation endpoint for AI/tools
		v1.GET("/docs", DocsHandler)

		// Thai Administrative Data endpoints
		v1.POST("/provinces", apiHandler.GetProvinces)
		v1.POST("/amphures", apiHandler.GetAmphures)
		v1.POST("/tambons", apiHandler.GetTambons)
		v1.POST("/findbyzipcode", apiHandler.FindByZipCode)
		// Search endpoints
		v1.POST("/search", apiHandler.SearchProducts)

		// Database endpoints
		v1.GET("/tables", apiHandler.GetTables)
		v1.POST("/command", apiHandler.CommandEndpoint)
		v1.POST("/select", apiHandler.SelectEndpoint)
		v1.POST("/pgcommand", apiHandler.PgCommandEndpoint)
		v1.POST("/pgselect", apiHandler.PgSelectEndpoint)

		// Image proxy endpoints
		v1.GET("/imgproxy", apiHandler.ImageProxy)
		v1.HEAD("/imgproxy", apiHandler.ImageProxyHead)
	}

	// Legacy API routes (maintain backwards compatibility)
	api := router.Group("/api")
	{
		// Database routes
		api.GET("/tables", apiHandler.GetTables)
	}

	// Legacy /get/ routes (maintain backwards compatibility)
	get := router.Group("/get")
	{
		get.POST("/provinces", apiHandler.GetProvinces)
		get.POST("/amphures", apiHandler.GetAmphures)
		get.POST("/tambons", apiHandler.GetTambons)
		get.POST("/findbyzipcode", apiHandler.FindByZipCode)
	}

	// API documentation endpoint
	router.GET("/", RootHandler)

	return router
}
