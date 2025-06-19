package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"smlgoapi/config"
	"smlgoapi/handlers"
	"smlgoapi/services"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// getLocalIP returns the local IP address of the machine
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Fallback to localhost if can't determine IP
		return "localhost"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// getDisplayURL returns a user-friendly URL for display
func getDisplayURL(serverAddr string) string {
	localIP := getLocalIP()

	// Replace 0.0.0.0 with actual IP for display purposes
	if strings.HasPrefix(serverAddr, "0.0.0.0:") {
		port := strings.TrimPrefix(serverAddr, "0.0.0.0:")
		return localIP + ":" + port
	}

	return serverAddr
}

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	// Initialize ClickHouse service
	var clickHouseService *services.ClickHouseService
	clickHouseService, err := services.NewClickHouseService(cfg)
	if err != nil {
		log.Printf("⚠️ ClickHouse service unavailable: %v", err)
		log.Println("🔄 Continuing with PostgreSQL-only mode...")
		clickHouseService = nil
	} else {
		defer clickHouseService.Close()
	}

	// Initialize PostgreSQL service
	postgreSQLService, err := services.NewPostgreSQLService(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize PostgreSQL service: %v", err)
	}
	defer postgreSQLService.Close()

	// Initialize API handlers
	apiHandler := handlers.NewAPIHandler(clickHouseService, postgreSQLService)

	// Setup Gin router
	router := setupRouter(apiHandler)
	// Create HTTP server
	srv := &http.Server{
		Addr:    cfg.GetServerAddress(),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		displayURL := getDisplayURL(cfg.GetServerAddress())
		log.Printf("🚀 SMLGOAPI Server starting on %s", cfg.GetServerAddress())
		log.Printf("📊 ClickHouse: %s@%s:%s/%s",
			cfg.ClickHouse.User,
			cfg.ClickHouse.Host,
			cfg.ClickHouse.Port,
			cfg.ClickHouse.Database)
		log.Printf("🐘 PostgreSQL: %s@%s:%s/%s",
			cfg.PostgreSQL.User,
			cfg.PostgreSQL.Host,
			cfg.PostgreSQL.Port,
			cfg.PostgreSQL.Database)
		log.Printf("🌐 API Endpoints:")
		log.Printf("  - Health Check: http://%s/health", displayURL)
		log.Printf("  - API v1 Base: http://%s/v1", displayURL)
		log.Printf("  - API Legacy: http://%s/api", displayURL)
		log.Printf("  - Documentation: http://%s/", displayURL)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")

	// Give a 5 second timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited")
}

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
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":     "SMLGOAPI - ClickHouse REST API",
			"version":     "1.0.0",
			"api_version": "v1",
			"endpoints": gin.H{
				// Core endpoints
				"health":           "GET /health", // API v1 endpoints (recommended)
				"v1_provinces":     "POST /v1/provinces",
				"v1_amphures":      "POST /v1/amphures",
				"v1_tambons":       "POST /v1/tambons",
				"v1_findbyzipcode": "POST /v1/findbyzipcode",
				"v1_search":        "POST /v1/search",
				"v1_command":       "POST /v1/command",
				"v1_select":        "POST /v1/select",
				"v1_pgcommand":     "POST /v1/pgcommand",
				"v1_pgselect":      "POST /v1/pgselect",
				"v1_tables":        "GET /v1/tables",
				"v1_imgproxy":      "GET /v1/imgproxy?url=<image_url>",

				// Legacy endpoints (backwards compatibility)
				"provinces":     "POST /get/provinces",
				"amphures":      "POST /get/amphures",
				"tambons":       "POST /get/tambons",
				"findbyzipcode": "POST /get/findbyzipcode",
				"search":        "POST /search",
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
	})

	return router
}
