package main

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

func handleHealth(c *gin.Context) {
	reqID := getNextRequestID()
	start := time.Now()

	printRequestDetails("GET", "/health", reqID, nil, nil)

	log.Printf("[handleHealth] 💚 GET /health - Health check accessed")

	response := HealthResponse{
		Status:       "healthy",
		Message:      "Vector search service is running",
		Timestamp:    time.Now().Format(time.RFC3339),
		RequestCount: atomic.LoadInt64(&requestCounter),
	}

	duration := time.Since(start).Seconds() * 1000
	printResponseDetails(reqID, 200, response, duration)

	c.JSON(200, response)
}
