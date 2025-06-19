package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"smlgoapi/config"
	"smlgoapi/handlers"
	"smlgoapi/services"
	"syscall"
	"time"
)

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
