package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"smlgoapi/config"
	"smlgoapi/handlers"
	"smlgoapi/services"
)

const (
	shutdownTimeout  = 5 * time.Second
	serverStartMsg   = "🚀 SMLGOAPI Server starting on %s"
	postgresMsg      = " PostgreSQL: %s@%s:%s/%s"
	endpointsMsg     = "🌐 API Endpoints:"
	healthCheckMsg   = "  - Health Check: http://%s/health"
	healthCheckV1Msg = "  - Health Check v1: http://%s/v1/health"
	apiV1Msg         = "  - API v1 Base: http://%s/v1"
	apiLegacyMsg     = "  - API Legacy: http://%s/api"
	shutdownMsg      = "🛑 Shutting down server..."
	exitMsg          = "✅ Server exited"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize PostgreSQL service
	postgreSQLService, err := services.NewPostgreSQLService(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize PostgreSQL service: %v", err)
	}
	defer postgreSQLService.Close()

	// Initialize API handlers
	apiHandler := handlers.NewAPIHandler(postgreSQLService)

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
		log.Printf(serverStartMsg, cfg.GetServerAddress())
		log.Printf(postgresMsg,
			cfg.PostgreSQL.User,
			cfg.PostgreSQL.Host,
			cfg.PostgreSQL.Port,
			cfg.PostgreSQL.Database)
		log.Println(endpointsMsg)
		log.Printf(healthCheckMsg, displayURL)
		log.Printf(healthCheckV1Msg, displayURL)
		log.Printf(apiV1Msg, displayURL)
		log.Printf(apiLegacyMsg, displayURL)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println(shutdownMsg)

	// Give a timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}
	log.Println(exitMsg)
}
