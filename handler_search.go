package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

func handleSearch(c *gin.Context) {
	start := time.Now()
	reqID := getNextRequestID()

	// Check if request context is already cancelled
	if c.Request.Context().Err() != nil {
		c.JSON(408, map[string]interface{}{
			"error":           "Request timeout",
			"message":         "Request was cancelled before processing",
			"timeout_seconds": config.RequestTimeout.Seconds(),
		})
		return
	}

	// Create timeout context for search operation
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.RequestTimeout)
	defer cancel()

	var request SearchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		duration := time.Since(start).Seconds() * 1000
		errorResponse := map[string]string{"error": "Invalid request format"}
		printResponseDetails(reqID, 400, errorResponse, duration)
		c.JSON(400, errorResponse)
		return
	}

	if request.Limit == 0 {
		request.Limit = 30
	}

	// Log request details
	requestBody := map[string]interface{}{
		"query":        request.Query,
		"limit":        request.Limit,
		"offset":       request.Offset,
		"query_length": len(request.Query),
		"language":     "Thai",
	}

	// Check if query contains non-ASCII characters (Thai)
	for _, r := range request.Query {
		if r > 127 {
			requestBody["language"] = "Thai"
			break
		} else {
			requestBody["language"] = "English"
		}
	}

	printRequestDetails("POST", "/search", reqID, nil, requestBody)

	fmt.Printf("\n🔍 [handleSearch] STARTING SEARCH OPERATION:\n")
	fmt.Printf("   [handleSearch] Query: '%s'\n", request.Query)
	fmt.Printf("   [handleSearch] Offset: %d\n", request.Offset)
	fmt.Printf("   [handleSearch] Limit: %d\n", request.Limit)
	fmt.Printf("   [handleSearch] Timeout: %v\n", config.RequestTimeout)

	// Check timeout before search operation
	if ctx.Err() != nil {
		c.JSON(408, map[string]interface{}{
			"error":           "Request timeout during processing",
			"timeout_seconds": config.RequestTimeout.Seconds(),
		})
		return
	}

	// Perform search using vector database with timeout
	resultsChan := make(chan struct {
		resultsJSON string
		err         error
	}, 1)

	go func() {
		resultsJSON, err := vectorDB.SearchProducts(request.Query, request.Limit, request.Offset)
		resultsChan <- struct {
			resultsJSON string
			err         error
		}{resultsJSON, err}
	}()

	var resultsJSON string
	var searchErr error

	select {
	case searchResult := <-resultsChan:
		resultsJSON = searchResult.resultsJSON
		searchErr = searchResult.err
	case <-ctx.Done():
		duration := time.Since(start).Seconds() * 1000
		errorResponse := map[string]interface{}{
			"error":           "Search operation timeout",
			"timeout_seconds": config.RequestTimeout.Seconds(),
			"duration_ms":     duration,
		}
		printResponseDetails(reqID, 408, errorResponse, duration)
		c.JSON(408, errorResponse)
		return
	}

	if searchErr != nil {
		duration := time.Since(start).Seconds() * 1000
		errorResponse := map[string]string{"error": searchErr.Error(), "query": request.Query}
		printResponseDetails(reqID, 500, errorResponse, duration)
		log.Printf("❌ Search error: %v", searchErr)
		c.JSON(500, errorResponse)
		return
	}

	var results map[string]interface{}
	if err := json.Unmarshal([]byte(resultsJSON), &results); err != nil {
		duration := time.Since(start).Seconds() * 1000
		errorResponse := map[string]string{"error": "Failed to parse search results"}
		printResponseDetails(reqID, 500, errorResponse, duration)
		c.JSON(500, errorResponse)
		return
	}

	duration := time.Since(start).Seconds() * 1000

	// Enhanced search results logging
	totalCount := results["total_count"]
	dataResults := results["data"]

	fmt.Printf("\n🔍 [handleSearch] SEARCH RESULTS DETAILS:\n")
	fmt.Printf("   [handleSearch] Query: '%s'\n", request.Query)
	fmt.Printf("   [handleSearch] Total Found: %v records\n", totalCount)
	if dataArray, ok := dataResults.([]interface{}); ok {
		fmt.Printf("   [handleSearch] Returned: %d results\n", len(dataArray))
		if len(dataArray) > 0 {
			fmt.Printf("   [handleSearch] Top Results:\n")
			for i, result := range dataArray {
				if i >= 5 {
					break
				}
				if resultMap, ok := result.(map[string]interface{}); ok {
					name := resultMap["name"]
					score := resultMap["similarity_score"]
					qty := resultMap["balance_qty"]
					fmt.Printf("     [handleSearch] %d. %v (score: %.3f, qty: %v)\n", i+1, name, score, qty)
				}
			}
		}
	}
	fmt.Printf("   [handleSearch] Offset: %d\n", request.Offset)
	fmt.Printf("   [handleSearch] Limit: %d\n", request.Limit)
	fmt.Printf("   [handleSearch] Duration: %.1fms\n", duration)

	printResponseDetails(reqID, 200, results, duration)

	if stats != nil {
		atomic.AddInt64(&stats.TotalRequests, 1)
	}

	c.JSON(200, results)
}
