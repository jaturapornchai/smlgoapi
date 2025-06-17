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
	clickHouseService, err := services.NewClickHouseService(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize ClickHouse service: %v", err)
	}
	defer clickHouseService.Close()

	// Initialize API handlers
	apiHandler := handlers.NewAPIHandler(clickHouseService)

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
		log.Printf("🌐 API Endpoints:")
		log.Printf("  - Health Check: http://%s/health", displayURL)
		log.Printf("  - API Base: http://%s/api", displayURL)

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

	// Health check endpoint
	router.GET("/health", apiHandler.HealthCheck)
	// Search endpoint
	router.POST("/search", apiHandler.SearchProducts)
	// Image proxy endpoint (GET and HEAD)
	router.GET("/imgproxy", apiHandler.ImageProxy)
	router.HEAD("/imgproxy", apiHandler.ImageProxyHead)
	// Universal SQL endpoints
	router.POST("/command", apiHandler.CommandEndpoint)
	router.POST("/select", apiHandler.SelectEndpoint)

	// Documentation endpoint for AI agents
	router.GET("/guide", apiHandler.GuideEndpoint)

	// API routes
	api := router.Group("/api")
	{
		// Database routes
		api.GET("/tables", apiHandler.GetTables)
	} // API documentation endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "SMLGOAPI - ClickHouse REST API",
			"version": "1.0.0", "endpoints": gin.H{
				"health":   "/health",
				"search":   "POST /search",
				"imgproxy": "GET /imgproxy?url=<image_url>",
				"command":  "POST /command",
				"select":   "POST /select",
				"guide":    "GET /guide",
				"tables":   "/api/tables",
			},
			"documentation": "Available endpoints listed above",
		})
	})

	return router
}
