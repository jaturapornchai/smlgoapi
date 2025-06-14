package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func handleCommandGet(c *gin.Context) {
	start := time.Now()
	reqID := getNextRequestID()

	// Check if request context is already cancelled
	if c.Request.Context().Err() != nil {
		c.JSON(408, map[string]interface{}{
			"error":   "Request timeout",
			"message": "Request was cancelled before processing",
		})
		return
	}

	// Start debug trace
	startDebugTrace(reqID, "GET", "/commandget")

	// Step 1: Extract query parameter
	addDebugStep(reqID, "Extract Query Parameter", map[string]interface{}{
		"raw_query_params": c.Request.URL.Query(),
	})

	queryBase64 := c.Query("q")
	if queryBase64 == "" {
		completeDebugStep(reqID, "Extract Query Parameter", "ERROR", nil, "Missing required parameter 'q'", nil)
		completeDebugTrace(reqID, "ERROR")

		duration := time.Since(start).Seconds() * 1000
		errorResponse := map[string]string{"error": "Missing required parameter 'q' (base64 encoded query)"}
		printResponseDetails(reqID, 400, errorResponse, duration)
		c.JSON(400, errorResponse)
		return
	}

	completeDebugStep(reqID, "Extract Query Parameter", "SUCCESS", map[string]interface{}{
		"query_base64": queryBase64,
		"query_length": len(queryBase64),
	}, "", nil)

	// Step 2: Decode base64 query
	addDebugStep(reqID, "Decode Base64 Query", map[string]interface{}{
		"encoded_query": queryBase64,
	})

	decodedQuery, err := decodeBase64Query(queryBase64)
	if err != nil {
		completeDebugStep(reqID, "Decode Base64 Query", "ERROR", nil, err.Error(), nil)
		completeDebugTrace(reqID, "ERROR")

		duration := time.Since(start).Seconds() * 1000
		errorResponse := map[string]string{"error": "Invalid base64 encoding in parameter 'q'"}
		printResponseDetails(reqID, 400, errorResponse, duration)
		c.JSON(400, errorResponse)
		return
	}

	completeDebugStep(reqID, "Decode Base64 Query", "SUCCESS", map[string]interface{}{
		"decoded_query":  decodedQuery,
		"decoded_length": len(decodedQuery),
	}, "", nil)

	queryParams := map[string]string{
		"q": queryBase64,
	}

	printRequestDetails("GET", "/commandget", reqID, queryParams, map[string]interface{}{
		"decoded_query": decodedQuery,
	})

	// Check timeout again before SQL execution
	if c.Request.Context().Err() != nil {
		completeDebugTrace(reqID, "TIMEOUT")
		c.JSON(408, map[string]interface{}{
			"error":           "Request timeout during processing",
			"timeout_seconds": config.RequestTimeout.Seconds(),
		})
		return
	}

	// Step 3: Execute SQL command with context timeout
	addDebugStep(reqID, "Execute SQL Command", map[string]interface{}{
		"sql_query": decodedQuery,
	})

	result := executeCommandWithContext(c.Request.Context(), decodedQuery, reqID)

	if result.Result != nil {
		if errorResult, ok := result.Result.(map[string]interface{}); ok {
			if errorMsg, exists := errorResult["error"]; exists {
				completeDebugStep(reqID, "Execute SQL Command", "ERROR", result.Result, fmt.Sprintf("%v", errorMsg), nil)
				completeDebugTrace(reqID, "ERROR")
			} else {
				completeDebugStep(reqID, "Execute SQL Command", "SUCCESS", result.Result, "", nil)
			}
		} else {
			completeDebugStep(reqID, "Execute SQL Command", "SUCCESS", result.Result, "", nil)
		}
	}

	response := CommandResponse{
		Result:     result.Result,
		Command:    decodedQuery,
		DecodedSQL: decodedQuery,
		Method:     "GET",
	}

	// Add debug trace to response if enabled
	if config.DebugMode && config.LogStepByStep {
		if trace := getDebugTrace(reqID); trace != nil {
			if responseMap, ok := response.Result.(map[string]interface{}); ok {
				responseMap["debug_trace"] = trace
			} else {
				// Create new response structure with debug info
				response.Result = map[string]interface{}{
					"data":        response.Result,
					"debug_trace": trace,
				}
			}
		}
	}

	duration := time.Since(start).Seconds() * 1000
	logPerformanceMetrics(reqID, "/commandget", time.Since(start), map[string]interface{}{
		"sql_query_length": len(decodedQuery),
		"response_size":    len(fmt.Sprintf("%v", response.Result)),
	})

	completeDebugTrace(reqID, "SUCCESS")
	printResponseDetails(reqID, 200, response, duration)

	c.JSON(200, response)
}
