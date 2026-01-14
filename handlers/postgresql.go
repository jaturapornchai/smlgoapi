package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"smlgoapi/models"

	"github.com/gin-gonic/gin"
)

// PgCommandEndpoint godoc
// @Summary Execute PostgreSQL command
// @Description Execute PostgreSQL command via JSON request
// @Tags postgresql
// @Accept json
// @Produce json
// @Param command body models.CommandRequest true "Command to execute"
// @Success 200 {object} models.CommandResponse
// @Router /pgcommand [post]
func (h *APIHandler) PgCommandEndpoint(c *gin.Context) {
	startTime := time.Now()

	// Parse JSON request
	var commandReq models.CommandRequest
	if err := c.ShouldBindJSON(&commandReq); err != nil {
		log.Printf("❌ [pgcommand] JSON bind error: %v", err)
		c.JSON(http.StatusBadRequest, models.CommandResponse{
			Success: false,
			Error:   msgInvalidJSON + ": " + err.Error(),
		})
		return
	}

	log.Printf("🎯 [pgcommand] Executing PostgreSQL command in database '%s': %s", commandReq.DatabaseName, commandReq.Query)

	// Get connection to specific database (Use System Connection for Commands/Writes)
	db, err := h.postgreSQLService.GetSystemDatabaseConnection(commandReq.DatabaseName)
	if err != nil {
		log.Printf("❌ [pgcommand] Database connection failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.CommandResponse{
			Success:  false,
			Database: commandReq.DatabaseName,
			Duration: float64(time.Since(startTime).Nanoseconds()) / 1e6,
			Error:    msgDatabaseError,
		})
		return
	}
	defer db.Close()

	// Execute command
	result, err := db.Exec(commandReq.Query)
	duration := float64(time.Since(startTime).Nanoseconds()) / 1e6

	if err != nil {
		log.Printf("❌ [pgcommand] Execution failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.CommandResponse{
			Success:  false,
			Database: commandReq.DatabaseName,
			Error:    fmt.Sprintf("PostgreSQL command execution failed: %s", err.Error()),
			Command:  commandReq.Query,
			Duration: duration,
		})
		return
	}

	// Get rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("⚠️ [pgcommand] Could not get rows affected: %v", err)
		rowsAffected = -1
	}

	log.Printf("✅ [pgcommand] Execution successful, %d rows affected in %.2fms", rowsAffected, duration)

	c.JSON(http.StatusOK, models.CommandResponse{
		Success: true,
		Message: "PostgreSQL command executed successfully",
		Result: map[string]interface{}{
			"rows_affected": rowsAffected,
			"query":         commandReq.Query,
		},
		Database: commandReq.DatabaseName,
		Command:  commandReq.Query,
		Duration: duration,
	})
}

// PgSelectEndpoint godoc
// @Summary Execute PostgreSQL SELECT query
// @Description Execute PostgreSQL SELECT query and return data via JSON request
// @Tags postgresql
// @Accept json
// @Produce json
// @Param select body models.SelectRequest true "SELECT query to execute"
// @Success 200 {object} models.SelectResponse
// @Router /pgselect [post]
func (h *APIHandler) PgSelectEndpoint(c *gin.Context) {
	startTime := time.Now()

	// Parse JSON request
	var selectReq models.SelectRequest
	if err := c.ShouldBindJSON(&selectReq); err != nil {
		log.Printf("❌ [pgselect] JSON bind error: %v", err)
		c.JSON(http.StatusBadRequest, models.SelectResponse{
			Success: false,
			Error:   msgInvalidJSON + ": " + err.Error(),
		})
		return
	}

	log.Printf("🎯 [pgselect] Executing PostgreSQL query in database '%s': %s", selectReq.DatabaseName, selectReq.Query)

	// Get connection to specific database
	db, err := h.postgreSQLService.GetDatabaseConnection(selectReq.DatabaseName)
	if err != nil {
		log.Printf("❌ [pgselect] Database connection failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.SelectResponse{
			Success:  false,
			Database: selectReq.DatabaseName,
			Duration: float64(time.Since(startTime).Nanoseconds()) / 1e6,
			Error:    msgDatabaseError,
		})
		return
	}
	defer db.Close()

	// Execute select query
	rows, err := db.Query(selectReq.Query)
	if err != nil {
		log.Printf("❌ [pgselect] Query failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.SelectResponse{
			Success:  false,
			Database: selectReq.DatabaseName,
			Error:    fmt.Sprintf("PostgreSQL query execution failed: %s", err.Error()),
			Query:    selectReq.Query,
			Duration: float64(time.Since(startTime).Nanoseconds()) / 1e6,
		})
		return
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		log.Printf("❌ [pgselect] Failed to get columns: %v", err)
		c.JSON(http.StatusInternalServerError, models.SelectResponse{
			Success:  false,
			Database: selectReq.DatabaseName,
			Error:    msgDatabaseError,
			Query:    selectReq.Query,
			Duration: float64(time.Since(startTime).Nanoseconds()) / 1e6,
		})
		return
	}

	// Process rows
	var data []interface{}
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePointers := make([]interface{}, len(columns))
		for i := range values {
			valuePointers[i] = &values[i]
		}

		// Scan the row
		if err := rows.Scan(valuePointers...); err != nil {
			log.Printf("❌ [pgselect] Row scan failed: %v", err)
			c.JSON(http.StatusInternalServerError, models.SelectResponse{
				Success:  false,
				Database: selectReq.DatabaseName,
				Error:    msgDatabaseError,
				Query:    selectReq.Query,
				Duration: float64(time.Since(startTime).Nanoseconds()) / 1e6,
			})
			return
		}

		// Convert to map
		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		data = append(data, row)
	}

	// Check for errors after iteration
	if err := rows.Err(); err != nil {
		log.Printf("❌ [pgselect] Row iteration error: %v", err)
		c.JSON(http.StatusInternalServerError, models.SelectResponse{
			Success:  false,
			Database: selectReq.DatabaseName,
			Error:    msgDatabaseError,
			Query:    selectReq.Query,
			Duration: float64(time.Since(startTime).Nanoseconds()) / 1e6,
		})
		return
	}

	duration := float64(time.Since(startTime).Nanoseconds()) / 1e6
	log.Printf("✅ [pgselect] Query successful, returned %d rows in %.2fms", len(data), duration)

	c.JSON(http.StatusOK, models.SelectResponse{
		Success:    true,
		Message:    "PostgreSQL query executed successfully",
		Data:       data,
		RowCount:   len(data),
		Columns:    columns,
		Database:   selectReq.DatabaseName,
		Query:      selectReq.Query,
		Duration:   duration,
		ServerTime: time.Now().Format(time.RFC3339),
	})
}

// PgSelectGroupEndpoint godoc
// @Summary Execute multiple PostgreSQL SELECT queries
// @Description Execute multiple PostgreSQL SELECT queries in a single request
// @Tags postgresql
// @Accept json
// @Produce json
// @Param selectgroup body models.GroupSelectRequest true "Multiple SELECT queries to execute"
// @Success 200 {object} models.GroupSelectResponse
// @Router /pgselectgroup [post]
func (h *APIHandler) PgSelectGroupEndpoint(c *gin.Context) {
	startTime := time.Now()

	// Parse JSON request
	var groupReq models.GroupSelectRequest
	if err := c.ShouldBindJSON(&groupReq); err != nil {
		log.Printf("❌ [pgselectgroup] JSON bind error: %v", err)
		c.JSON(http.StatusBadRequest, models.GroupSelectResponse{
			Success: false,
			Error:   msgInvalidJSON + ": " + err.Error(),
		})
		return
	}

	log.Printf("🎯 [pgselectgroup] Executing %d PostgreSQL queries in database '%s'", len(groupReq.Queries), groupReq.DatabaseName)

	var results []models.GroupSelectQueryResult

	// Execute each query
	for i, query := range groupReq.Queries {
		queryStartTime := time.Now()

		// Get connection to specific database
		db, err := h.postgreSQLService.GetDatabaseConnection(groupReq.DatabaseName)
		if err != nil {
			log.Printf("❌ [pgselectgroup] Database connection failed for query %d: %v", i+1, err)
			results = append(results, models.GroupSelectQueryResult{
				QueryIndex: i + 1,
				Success:    false,
				Error:      fmt.Sprintf("Database connection failed: %s", err.Error()),
				Duration:   float64(time.Since(queryStartTime).Nanoseconds()) / 1e6,
			})
			continue
		}

		// Execute select query
		rows, err := db.Query(query)
		if err != nil {
			log.Printf("❌ [pgselectgroup] Query %d failed: %v", i+1, err)
			db.Close()
			results = append(results, models.GroupSelectQueryResult{
				QueryIndex: i + 1,
				Success:    false,
				Error:      fmt.Sprintf("PostgreSQL query execution failed: %s", err.Error()),
				Query:      query,
				Duration:   float64(time.Since(queryStartTime).Nanoseconds()) / 1e6,
			})
			continue
		}

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			log.Printf("❌ [pgselectgroup] Failed to get columns for query %d: %v", i+1, err)
			rows.Close()
			db.Close()
			results = append(results, models.GroupSelectQueryResult{
				QueryIndex: i + 1,
				Success:    false,
				Error:      "Failed to get column information",
				Query:      query,
				Duration:   float64(time.Since(queryStartTime).Nanoseconds()) / 1e6,
			})
			continue
		}

		// Process rows
		var data []interface{}
		for rows.Next() {
			// Create a slice of interface{} to hold the values
			values := make([]interface{}, len(columns))
			valuePointers := make([]interface{}, len(columns))
			for j := range values {
				valuePointers[j] = &values[j]
			}

			// Scan the row
			if err := rows.Scan(valuePointers...); err != nil {
				log.Printf("❌ [pgselectgroup] Row scan failed for query %d: %v", i+1, err)
				break
			}

			// Convert to map
			row := make(map[string]interface{})
			for j, col := range columns {
				row[col] = values[j]
			}
			data = append(data, row)
		}

		// Check for errors after iteration
		if err := rows.Err(); err != nil {
			log.Printf("❌ [pgselectgroup] Row iteration error for query %d: %v", i+1, err)
			rows.Close()
			db.Close()
			results = append(results, models.GroupSelectQueryResult{
				QueryIndex: i + 1,
				Success:    false,
				Error:      "Row iteration error",
				Query:      query,
				Duration:   float64(time.Since(queryStartTime).Nanoseconds()) / 1e6,
			})
			continue
		}

		rows.Close()
		db.Close()

		duration := float64(time.Since(queryStartTime).Nanoseconds()) / 1e6
		log.Printf("✅ [pgselectgroup] Query %d successful, returned %d rows in %.2fms", i+1, len(data), duration)

		results = append(results, models.GroupSelectQueryResult{
			QueryIndex: i + 1,
			Success:    true,
			Data:       data,
			RowCount:   len(data),
			Columns:    columns,
			Query:      query,
			Duration:   duration,
		})
	}

	totalDuration := float64(time.Since(startTime).Nanoseconds()) / 1e6

	// Count successful queries
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	log.Printf("✅ [pgselectgroup] Completed %d/%d queries successfully in %.2fms", successCount, len(groupReq.Queries), totalDuration)

	c.JSON(http.StatusOK, models.GroupSelectResponse{
		Success:      true,
		Message:      fmt.Sprintf("PostgreSQL group query completed: %d/%d successful", successCount, len(groupReq.Queries)),
		Results:      results,
		TotalQueries: len(groupReq.Queries),
		Successful:   successCount,
		Failed:       len(groupReq.Queries) - successCount,
		Database:     groupReq.DatabaseName,
		Duration:     totalDuration,
		ServerTime:   time.Now().Format(time.RFC3339),
	})
}

// PgCheckDatabaseEndpoint godoc
// @Summary Check PostgreSQL database existence
// @Description Check if a PostgreSQL database exists and get database information
// @Tags postgresql
// @Accept json
// @Produce json
// @Param checkdb body models.CheckDatabaseRequest true "Database to check"
// @Success 200 {object} models.CheckDatabaseResponse
// @Router /pgcheckdatabase [post]
func (h *APIHandler) PgCheckDatabaseEndpoint(c *gin.Context) {
	startTime := time.Now()

	// Parse JSON request
	var checkReq models.CheckDatabaseRequest
	if err := c.ShouldBindJSON(&checkReq); err != nil {
		log.Printf("❌ [pgcheckdatabase] JSON bind error: %v", err)
		c.JSON(http.StatusBadRequest, models.CheckDatabaseResponse{
			Success: false,
			Error:   msgInvalidJSON + ": " + err.Error(),
		})
		return
	}

	log.Printf("🎯 [pgcheckdatabase] Checking PostgreSQL database: %s", checkReq.DatabaseName)

	// Check if database exists
	exists, err := h.postgreSQLService.CheckDatabaseExists(checkReq.DatabaseName)
	if err != nil {
		log.Printf("❌ [pgcheckdatabase] Database check failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.CheckDatabaseResponse{
			Success:  false,
			Database: checkReq.DatabaseName,
			Error:    fmt.Sprintf("PostgreSQL database check failed: %s", err.Error()),
			Duration: float64(time.Since(startTime).Nanoseconds()) / 1e6,
		})
		return
	}

	var tables []string
	if exists {
		// Get table list if database exists
		tables, err = h.postgreSQLService.GetTableList(checkReq.DatabaseName)
		if err != nil {
			log.Printf("❌ [pgcheckdatabase] Failed to get table list: %v", err)
			c.JSON(http.StatusInternalServerError, models.CheckDatabaseResponse{
				Success:  false,
				Database: checkReq.DatabaseName,
				Error:    fmt.Sprintf("Failed to get table list: %s", err.Error()),
				Duration: float64(time.Since(startTime).Nanoseconds()) / 1e6,
			})
			return
		}
	}

	duration := float64(time.Since(startTime).Nanoseconds()) / 1e6
	log.Printf("✅ [pgcheckdatabase] Database check completed in %.2fms: exists=%v, tables=%d", duration, exists, len(tables))

	c.JSON(http.StatusOK, models.CheckDatabaseResponse{
		Success:    true,
		Message:    fmt.Sprintf("PostgreSQL database check completed: %s", checkReq.DatabaseName),
		Database:   checkReq.DatabaseName,
		Exists:     exists,
		Tables:     tables,
		TableCount: len(tables),
		Duration:   duration,
		ServerTime: time.Now().Format(time.RFC3339),
	})
}

// PgCheckTableEndpoint godoc
// @Summary Check PostgreSQL table existence
// @Description Check if a PostgreSQL table exists in a specific database
// @Tags postgresql
// @Accept json
// @Produce json
// @Param checktable body models.CheckTableRequest true "Database and table to check"
// @Success 200 {object} models.CheckTableResponse
// @Router /pgchecktable [post]
func (h *APIHandler) PgCheckTableEndpoint(c *gin.Context) {
	startTime := time.Now()
	// Read raw body for better debug info (allows us to inspect keys when binding fails)
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("❌ [pgchecktable] Could not read request body: %v", err)
		c.JSON(http.StatusBadRequest, models.CheckTableResponse{Success: false, Error: "cannot read request body"})
		return
	}
	// Restore body for potential downstream readers (not strictly needed here)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	var checkReq models.CheckTableRequest
	bindErr := json.Unmarshal(rawBody, &checkReq)

	// Collect keys for diagnostics
	var keyMap map[string]interface{}
	_ = json.Unmarshal(rawBody, &keyMap)

	// Fallback: support camelCase keys if snake_case missing
	if checkReq.DatabaseName == "" && keyMap != nil {
		if v, ok := keyMap["databaseName"].(string); ok && v != "" {
			checkReq.DatabaseName = v
		}
	}
	if checkReq.TableName == "" && keyMap != nil {
		if v, ok := keyMap["tableName"].(string); ok && v != "" {
			checkReq.TableName = v
		}
	}

	// Manual validation (mirrors binding:"required")
	var validationErrors []string
	if checkReq.DatabaseName == "" {
		validationErrors = append(validationErrors, "database_name (or databaseName) is required")
	}
	if checkReq.TableName == "" {
		validationErrors = append(validationErrors, "table_name (or tableName) is required")
	}

	if bindErr != nil || len(validationErrors) > 0 {
		log.Printf("❌ [pgchecktable] JSON bind/validation error: %v | keys=%v | raw=%s", bindErr, keysFromMap(keyMap), string(rawBody))
		c.JSON(http.StatusBadRequest, models.CheckTableResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid JSON: %v; validation: %v; provided_keys: %v", bindErr, validationErrors, keysFromMap(keyMap)),
		})
		return
	}

	log.Printf("🎯 [pgchecktable] Checking PostgreSQL table '%s' in database '%s'", checkReq.TableName, checkReq.DatabaseName)

	// Check if table exists
	exists, err := h.postgreSQLService.CheckTableExists(checkReq.DatabaseName, checkReq.TableName)
	if err != nil {
		log.Printf("❌ [pgchecktable] Table check failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.CheckTableResponse{
			Success:  false,
			Database: checkReq.DatabaseName,
			Table:    checkReq.TableName,
			Error:    fmt.Sprintf("PostgreSQL table check failed: %s", err.Error()),
			Duration: float64(time.Since(startTime).Nanoseconds()) / 1e6,
		})
		return
	}

	duration := float64(time.Since(startTime).Nanoseconds()) / 1e6
	log.Printf("✅ [pgchecktable] Table check completed in %.2fms: exists=%v", duration, exists)

	c.JSON(http.StatusOK, models.CheckTableResponse{
		Success:    true,
		Message:    fmt.Sprintf("PostgreSQL table check completed: %s.%s", checkReq.DatabaseName, checkReq.TableName),
		Database:   checkReq.DatabaseName,
		Table:      checkReq.TableName,
		Exists:     exists,
		Duration:   duration,
		ServerTime: time.Now().Format(time.RFC3339),
	})
}

// keysFromMap extracts sorted keys for logging (local helper)
func keysFromMap(m map[string]interface{}) []string {
	if m == nil {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
