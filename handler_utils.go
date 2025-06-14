package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"math"
	"strings"
	"sync/atomic"
	"time"
)

// ===== UTILITY FUNCTIONS =====

// Request ID management
func getNextRequestID() int64 {
	return atomic.AddInt64(&requestCounter, 1)
}

// Base64 encoding/decoding utilities
func decodeBase64Query(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}
	return string(decoded), nil
}

func encodeQueryToBase64(query string) string {
	return base64.StdEncoding.EncodeToString([]byte(query))
}

// ===== DEBUG UTILITY FUNCTIONS =====

func startDebugTrace(reqID int64, method, endpoint string) *DebugTrace {
	if !config.DebugMode || !config.LogStepByStep {
		return nil
	}

	debugMutex.Lock()
	defer debugMutex.Unlock()

	trace := &DebugTrace{
		RequestID:   reqID,
		Method:      method,
		Endpoint:    endpoint,
		StartTime:   time.Now(),
		Steps:       make([]DebugStep, 0),
		Completed:   false,
		FinalStatus: "IN_PROGRESS",
	}

	debugTraces[reqID] = trace

	logDebug(DEBUG_INFO, fmt.Sprintf("🚀 Starting debug trace for %s %s", method, endpoint), map[string]interface{}{
		"request_id": reqID,
		"start_time": trace.StartTime.Format("2006-01-02 15:04:05.000"),
	})

	return trace
}

func addDebugStep(reqID int64, stepName string, input interface{}) *DebugStep {
	if !config.DebugMode || !config.LogStepByStep {
		return &DebugStep{StepName: stepName, Status: "SKIPPED"}
	}

	debugMutex.Lock()
	defer debugMutex.Unlock()

	trace := debugTraces[reqID]
	if trace == nil {
		return &DebugStep{StepName: stepName, Status: "NO_TRACE"}
	}

	step := DebugStep{
		StepNumber: len(trace.Steps) + 1,
		StepName:   stepName,
		Status:     "STARTED",
		StartTime:  time.Now(),
		Input:      input,
		Details:    make(map[string]interface{}),
	}

	trace.Steps = append(trace.Steps, step)
	trace.TotalSteps = len(trace.Steps)

	logDebug(DEBUG_DEBUG, fmt.Sprintf("📝 Step %d: %s STARTED", step.StepNumber, stepName), map[string]interface{}{
		"request_id": reqID,
		"input":      input,
		"start_time": step.StartTime.Format("15:04:05.000"),
	})

	return &trace.Steps[len(trace.Steps)-1]
}

func completeDebugStep(reqID int64, stepName string, status string, output interface{}, errorMsg string, details map[string]interface{}) {
	if !config.DebugMode || !config.LogStepByStep {
		return
	}

	debugMutex.Lock()
	defer debugMutex.Unlock()

	trace := debugTraces[reqID]
	if trace == nil {
		return
	}

	// Find the step to complete
	for i := len(trace.Steps) - 1; i >= 0; i-- {
		step := &trace.Steps[i]
		if step.StepName == stepName && step.Status == "STARTED" {
			now := time.Now()
			step.EndTime = &now
			step.Status = status
			step.Output = output
			step.Error = errorMsg
			step.Duration = now.Sub(step.StartTime).String()

			for k, v := range details {
				step.Details[k] = v
			}

			statusIcon := "✅"
			if status == "ERROR" {
				statusIcon = "❌"
			} else if status == "WARN" {
				statusIcon = "⚠️"
			}

			logDebug(DEBUG_DEBUG, fmt.Sprintf("%s Step %d: %s %s (%s)", statusIcon, step.StepNumber, stepName, status, step.Duration), map[string]interface{}{
				"request_id": reqID,
				"output":     output,
				"error":      errorMsg,
				"duration":   step.Duration,
			})

			break
		}
	}
}

func completeDebugTrace(reqID int64, finalStatus string) {
	if !config.DebugMode || !config.LogStepByStep {
		return
	}

	debugMutex.Lock()
	defer debugMutex.Unlock()

	trace := debugTraces[reqID]
	if trace == nil {
		return
	}

	trace.Completed = true
	trace.FinalStatus = finalStatus
	trace.TotalTime = time.Since(trace.StartTime).String()

	successCount := 0
	errorCount := 0
	for _, step := range trace.Steps {
		if step.Status == "SUCCESS" {
			successCount++
		} else if step.Status == "ERROR" {
			errorCount++
		}
	}

	statusIcon := "🎉"
	if finalStatus == "ERROR" {
		statusIcon = "💥"
	} else if errorCount > 0 {
		statusIcon = "⚠️"
	}

	logDebug(DEBUG_INFO, fmt.Sprintf("%s Debug trace completed for %s %s", statusIcon, trace.Method, trace.Endpoint), map[string]interface{}{
		"request_id":    reqID,
		"final_status":  finalStatus,
		"total_time":    trace.TotalTime,
		"total_steps":   trace.TotalSteps,
		"success_steps": successCount,
		"error_steps":   errorCount,
	})

	// Clean up old traces (keep last 100)
	if len(debugTraces) > 100 {
		oldestID := int64(math.MaxInt64)
		for id := range debugTraces {
			if id < oldestID {
				oldestID = id
			}
		}
		delete(debugTraces, oldestID)
	}
}

func getDebugTrace(reqID int64) *DebugTrace {
	if !config.DebugMode {
		return nil
	}

	debugMutex.RLock()
	defer debugMutex.RUnlock()

	return debugTraces[reqID]
}

func logSQLExecution(reqID int64, query string, params []interface{}, duration time.Duration, rowCount int, err error) {
	if !config.DebugMode || !config.LogSQLExecution {
		return
	}

	status := "SUCCESS"
	errorMsg := ""
	if err != nil {
		status = "ERROR"
		errorMsg = err.Error()
	}

	logDebug(DEBUG_DEBUG, fmt.Sprintf("🗄️ SQL Execution %s", status), map[string]interface{}{
		"request_id": reqID,
		"query":      query,
		"params":     params,
		"duration":   duration.String(),
		"row_count":  rowCount,
		"error":      errorMsg,
	})
}

func logPerformanceMetrics(reqID int64, endpoint string, totalDuration time.Duration, details map[string]interface{}) {
	if !config.DebugMode || !config.LogPerformance {
		return
	}

	logDebug(DEBUG_INFO, fmt.Sprintf("📊 Performance metrics for %s", endpoint), map[string]interface{}{
		"request_id":     reqID,
		"total_duration": totalDuration.String(),
		"details":        details,
	})
}

// ===== SQL COMMAND EXECUTION UTILITIES =====

// Helper function to execute commands with timeout (shared by both GET and POST handlers)
func executeCommand(query string, reqID int64) *CommandResponse {
	return executeCommandWithContext(context.Background(), query, reqID)
}

func executeCommandWithContext(parentCtx context.Context, query string, reqID int64) *CommandResponse {
	fmt.Printf("\n💻 [executeCommand] COMMAND EXECUTION:\n")
	fmt.Printf("   [executeCommand] Request ID: %d\n", reqID)
	fmt.Printf("   [executeCommand] Processing command\n")
	fmt.Printf("   [executeCommand] Query: %s\n", query)

	var result interface{}
	sqlStart := time.Now()

	// Create timeout context for SQL execution
	ctx, cancel := context.WithTimeout(parentCtx, config.SQLTimeout)
	defer cancel()

	if clickhouseDB != nil {
		fmt.Printf("   🔄 [executeCommand] Executing: %s...\n", query[:min(100, len(query))])

		// Create a channel to handle the SQL execution result
		resultChan := make(chan struct {
			rows *sql.Rows
			err  error
		}, 1)

		// Execute the command in a goroutine
		go func() {
			rows, err := clickhouseDB.QueryContext(ctx, query)
			resultChan <- struct {
				rows *sql.Rows
				err  error
			}{rows, err}
		}()

		// Wait for result or timeout
		select {
		case sqlResult := <-resultChan:
			sqlDuration := time.Since(sqlStart)

			if sqlResult.err != nil {
				fmt.Printf("   ❌ [executeCommand] Execution: FAILED - %v\n", sqlResult.err)
				logSQLExecution(reqID, query, nil, sqlDuration, 0, sqlResult.err)
				return &CommandResponse{
					Result:  map[string]interface{}{"error": sqlResult.err.Error()},
					Command: query,
				}
			}

			if sqlResult.rows != nil {
				defer sqlResult.rows.Close()

				// Convert rows to result
				columns, err := sqlResult.rows.Columns()
				if err != nil {
					result = "Command executed successfully"
					logSQLExecution(reqID, query, nil, sqlDuration, 0, nil)
				} else {
					var resultRows []map[string]interface{}
					rowCount := 0

					for sqlResult.rows.Next() {
						// Check if context was cancelled
						if ctx.Err() != nil {
							result = map[string]interface{}{"error": "Query execution timeout"}
							logSQLExecution(reqID, query, nil, sqlDuration, rowCount, ctx.Err())
							break
						}

						values := make([]interface{}, len(columns))
						valuePtrs := make([]interface{}, len(columns))
						for i := range values {
							valuePtrs[i] = &values[i]
						}

						if err := sqlResult.rows.Scan(valuePtrs...); err != nil {
							continue
						}

						row := make(map[string]interface{})
						for i, col := range columns {
							row[col] = values[i]
						}
						resultRows = append(resultRows, row)
						rowCount++
					}

					if len(resultRows) > 0 {
						result = resultRows
					} else if ctx.Err() == nil {
						result = "Command executed successfully"
					}

					logSQLExecution(reqID, query, nil, sqlDuration, rowCount, nil)
				}
			}

			fmt.Printf("   ✅ [executeCommand] Execution: SUCCESS\n")

		case <-ctx.Done():
			sqlDuration := time.Since(sqlStart)
			fmt.Printf("   ⏰ [executeCommand] Execution: TIMEOUT after %v\n", sqlDuration)
			logSQLExecution(reqID, query, nil, sqlDuration, 0, ctx.Err())
			return &CommandResponse{
				Result: map[string]interface{}{
					"error":           fmt.Sprintf("SQL execution timeout after %v", config.SQLTimeout),
					"timeout_seconds": config.SQLTimeout.Seconds(),
				},
				Command: query,
			}
		}
	} else {
		sqlDuration := time.Since(sqlStart)
		result = "Mock command execution - ClickHouse not connected"
		fmt.Printf("   ⚠️ [executeCommand] Mock execution (no DB connection)\n")
		logSQLExecution(reqID, query, nil, sqlDuration, 0, fmt.Errorf("no database connection"))
	}

	return &CommandResponse{
		Result:  result,
		Command: query,
	}
}

// ===== IMAGE PROCESSING UTILITIES =====

func generateColorHistogram(imageData []byte) ([]float32, error) {
	// Mock color histogram generation
	// In a real implementation, you would decode the image and compute actual histograms
	histogram := make([]float32, 99) // 32*3 + 3 + 3 = 99 features

	// Generate some pseudo-random but deterministic features based on image data
	for i := range histogram {
		val := float32(imageData[i%len(imageData)]) / 255.0
		if i%2 == 0 {
			val = val * 0.8
		}
		histogram[i] = val
	}

	// Normalize
	var sum float32
	for _, v := range histogram {
		sum += v
	}
	if sum > 0 {
		for i := range histogram {
			histogram[i] /= sum
		}
	}

	return histogram, nil
}

// ===== LOGGING UTILITIES =====

func printRequestDetails(method, endpoint string, reqID int64, queryParams map[string]string, body interface{}) {
	if !config.LogRequestResponse {
		return
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("[printRequestDetails] 📨 REQUEST #%d - %s %s\n", reqID, method, endpoint)
	fmt.Printf("%s\n", strings.Repeat("=", 80))
	fmt.Printf("[printRequestDetails] 🕐 Timestamp: %s\n", time.Now().Format("2006-01-02 15:04:05.000"))

	if len(queryParams) > 0 {
		fmt.Printf("[printRequestDetails] 📋 Query Parameters:\n")
		for key, value := range queryParams {
			fmt.Printf("   [printRequestDetails] • %s: %s\n", key, value)
		}
	}

	if body != nil {
		fmt.Printf("[printRequestDetails] 📦 Request Body:\n")
		if bodyMap, ok := body.(map[string]interface{}); ok {
			for key, value := range bodyMap {
				valueStr := fmt.Sprintf("%v", value)
				if len(valueStr) > 100 {
					fmt.Printf("   [printRequestDetails] • %s: %s... (truncated)\n", key, valueStr[:100])
				} else {
					fmt.Printf("   [printRequestDetails] • %s: %v\n", key, value)
				}
			}
		} else {
			bodyStr := fmt.Sprintf("%v", body)
			if len(bodyStr) > 200 {
				fmt.Printf("   [printRequestDetails] %s... (truncated)\n", bodyStr[:200])
			} else {
				fmt.Printf("   [printRequestDetails] %s\n", bodyStr)
			}
		}
	}
}

func printResponseDetails(reqID int64, statusCode int, responseData interface{}, durationMS float64) {
	if !config.LogRequestResponse {
		return
	}

	fmt.Printf("\n[printResponseDetails] 📤 RESPONSE #%d\n", reqID)
	fmt.Printf("%s\n", strings.Repeat("-", 50))
	fmt.Printf("[printResponseDetails] 📊 Status: %d\n", statusCode)
	fmt.Printf("[printResponseDetails] ⏱️  Duration: %.1fms\n", durationMS)

	if responseMap, ok := responseData.(map[string]interface{}); ok {
		fmt.Printf("[printResponseDetails] 📦 Response Data:\n")

		// Handle search response format
		if totalCount, exists := responseMap["total_count"]; exists {
			if dataResults, exists := responseMap["data"]; exists {
				fmt.Printf("   [printResponseDetails] • Total Found: %v records\n", totalCount)
				if dataArray, ok := dataResults.([]interface{}); ok {
					fmt.Printf("   [printResponseDetails] • Returned: %d results\n", len(dataArray))
				}
				if offset, exists := responseMap["offset"]; exists {
					fmt.Printf("   [printResponseDetails] • Offset: %v\n", offset)
				}
				if limit, exists := responseMap["limit"]; exists {
					fmt.Printf("   [printResponseDetails] • Limit: %v\n", limit)
				}

				// Show first 3 results
				if dataArray, ok := dataResults.([]interface{}); ok {
					for i, result := range dataArray {
						if i >= 3 {
							break
						}
						if resultMap, ok := result.(map[string]interface{}); ok {
							name := resultMap["name"]
							score := resultMap["similarity_score"]
							barcode := resultMap["barcode"]
							fmt.Printf("     [printResponseDetails] %d. %v\n", i+1, name)
							fmt.Printf("        [printResponseDetails] 📋 Barcode: %v\n", barcode)
							fmt.Printf("        [printResponseDetails] 🎯 Score: %.4f\n", score)
						}
					}
				}
			}
		} else {
			// General handling
			for key, value := range responseMap {
				valueStr := fmt.Sprintf("%v", value)
				if len(valueStr) > 100 {
					fmt.Printf("   [printResponseDetails] • %s: %s... (truncated)\n", key, valueStr[:100])
				} else {
					fmt.Printf("   [printResponseDetails] • %s: %v\n", key, value)
				}
			}
		}
	} else {
		responseStr := fmt.Sprintf("%v", responseData)
		if len(responseStr) > 200 {
			fmt.Printf("[printResponseDetails] 📦 Response: %s... (truncated)\n", responseStr[:200])
		} else {
			fmt.Printf("[printResponseDetails] 📦 Response: %s\n", responseStr)
		}
	}

	fmt.Printf("%s\n", strings.Repeat("=", 80))
}

// ===== HELPER FUNCTIONS =====

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
