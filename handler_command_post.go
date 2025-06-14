package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func handleCommandPost(c *gin.Context) {
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
	startDebugTrace(reqID, "POST", "/commandpost")

	// Step 1: Parse JSON request
	addDebugStep(reqID, "Parse JSON Request", nil)

	var request CommandPostRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		completeDebugStep(reqID, "Parse JSON Request", "ERROR", nil, err.Error(), nil)
		completeDebugTrace(reqID, "ERROR")

		duration := time.Since(start).Seconds() * 1000
		errorResponse := map[string]string{"error": "Invalid JSON format or missing required fields"}
		printResponseDetails(reqID, 400, errorResponse, duration)
		c.JSON(400, errorResponse)
		return
	}

	completeDebugStep(reqID, "Parse JSON Request", "SUCCESS", map[string]interface{}{
		"query_base64": request.QueryBase64,
		"query_length": len(request.QueryBase64),
	}, "", nil)

	// Step 2: Decode base64 query
	addDebugStep(reqID, "Decode Base64 Query", map[string]interface{}{
		"encoded_query": request.QueryBase64,
	})

	decodedQuery, err := decodeBase64Query(request.QueryBase64)
	if err != nil {
		completeDebugStep(reqID, "Decode Base64 Query", "ERROR", nil, err.Error(), nil)
		completeDebugTrace(reqID, "ERROR")

		duration := time.Since(start).Seconds() * 1000
		errorResponse := map[string]string{"error": fmt.Sprintf("Invalid base64 encoding: %v", err)}
		printResponseDetails(reqID, 400, errorResponse, duration)
		c.JSON(400, errorResponse)
		return
	}

	completeDebugStep(reqID, "Decode Base64 Query", "SUCCESS", map[string]interface{}{
		"decoded_query": decodedQuery,
		"query_length":  len(decodedQuery),
	}, "", nil)

	// Print request details
	requestBody := map[string]interface{}{
		"query_base64":   request.QueryBase64,
		"decoded_query":  decodedQuery,
		"query_length":   len(decodedQuery),
		"encoded_length": len(request.QueryBase64),
	}
	printRequestDetails("POST", "/commandpost", reqID, nil, requestBody)

	// Step 3: Execute command
	addDebugStep(reqID, "Execute SQL Command", map[string]interface{}{
		"sql_query": decodedQuery,
	})

	response := executeCommand(decodedQuery, reqID)
	if response == nil {
		completeDebugStep(reqID, "Execute SQL Command", "ERROR", nil, "Command execution returned nil", nil)
		completeDebugTrace(reqID, "ERROR")

		duration := time.Since(start).Seconds() * 1000
		errorResponse := map[string]string{"error": "Command execution failed"}
		printResponseDetails(reqID, 500, errorResponse, duration)
		c.JSON(500, errorResponse)
		return
	}

	completeDebugStep(reqID, "Execute SQL Command", "SUCCESS", map[string]interface{}{
		"result_type": fmt.Sprintf("%T", response.Result),
		"command":     response.Command,
	}, "", nil)

	// Step 4: Prepare response
	addDebugStep(reqID, "Prepare Response", nil)

	commandResponse := &CommandResponse{
		Result:     response.Result,
		Command:    response.Command,
		DecodedSQL: decodedQuery,
		Method:     "POST",
	}

	completeDebugStep(reqID, "Prepare Response", "SUCCESS", commandResponse, "", nil)
	completeDebugTrace(reqID, "SUCCESS")

	// Send response
	duration := time.Since(start).Seconds() * 1000
	printResponseDetails(reqID, 200, commandResponse, duration)
	c.JSON(200, commandResponse)
}
