package main

import (
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

func handleRoot(c *gin.Context) {
	start := time.Now()
	reqID := getNextRequestID()

	queryParams := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	printRequestDetails("GET", "/", reqID, queryParams, nil)

	responseData := map[string]interface{}{
		"message":       "Product Vector Search API - Enhanced Test Mode",
		"status":        "running",
		"timestamp":     time.Now().Format(time.RFC3339),
		"request_count": atomic.LoadInt64(&requestCounter),
	}

	duration := time.Since(start).Seconds() * 1000
	printResponseDetails(reqID, 200, responseData, duration)

	c.JSON(200, responseData)
}
