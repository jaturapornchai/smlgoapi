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
	router.Use(gin.Recovery())

	// CORS middleware
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

	// Health check endpoint (available at root level)
	router.GET("/health", apiHandler.HealthCheck)

	// All API endpoints under /v1
	v1 := router.Group("/v1")
	{
		// Health check endpoint (also available under v1)
		v1.GET("/health", apiHandler.HealthCheck)

		// API guide endpoint
		v1.GET("/guide", apiHandler.GuideEndpoint)

		// PostgreSQL Database endpoints
		v1.POST("/pgcommand", apiHandler.PgCommandEndpoint)
		v1.POST("/pgselect", apiHandler.PgSelectEndpoint)
		v1.POST("/pgselectgroup", apiHandler.PgSelectGroupEndpoint)
		v1.POST("/pgcheckdatabase", apiHandler.PgCheckDatabaseEndpoint)
		v1.POST("/pgchecktable", apiHandler.PgCheckTableEndpoint)

		// ClickHouse Database endpoints
		v1.POST("/chcommand", apiHandler.ChCommandEndpoint)
		v1.POST("/chselect", apiHandler.ChSelectEndpoint)
		v1.POST("/chselectgroup", apiHandler.ChSelectGroupEndpoint)
		v1.POST("/chcheckdatabase", apiHandler.ChCheckDatabaseEndpoint)
		v1.POST("/chchecktable", apiHandler.ChCheckTableEndpoint)

		// Result handling endpoints
		v1.POST("/resultfromquery", apiHandler.ResultFromQueryHandler)
		v1.POST("/resultget", apiHandler.ResultGetHandler)
		v1.POST("/resulttopdf", apiHandler.ResultToPDFHandler)
		v1.POST("/sendreportemail", apiHandler.SendReportEmailHandler)

		// MongoDB Atlas endpoints
		v1.POST("/mongoatlasupdate", handlers.MongoAtlasUpdateHandler) // Insert/Update (Upsert)
		v1.POST("/mongoatlasdelete", handlers.MongoAtlasDeleteHandler) // Delete
		v1.POST("/mongoatlasget", handlers.MongoAtlasGetHandler)       // Get/Query
	}

	return router
}
