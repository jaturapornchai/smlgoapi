package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"smlgoapi/models"

	"github.com/gin-gonic/gin"
)

// ChCommandEndpoint godoc
// @Summary Execute ClickHouse command
// @Description Execute ClickHouse command via JSON request
// @Tags clickhouse
// @Accept json
// @Produce json
// @Param command body models.CommandRequest true "Command to execute"
// @Success 200 {object} models.CommandResponse
// @Router /chcommand [post]
func (h *APIHandler) ChCommandEndpoint(c *gin.Context) {
	startTime := time.Now()

	// Check if ClickHouse service is available
	if h.clickHouseService == nil {
		log.Printf("❌ [chcommand] ClickHouse service is not available")
		c.JSON(http.StatusServiceUnavailable, models.CommandResponse{
			Success: false,
			Error:   "ClickHouse service is not available",
		})
		return
	}

	// Parse JSON request
	var commandReq models.CommandRequest
	if err := c.ShouldBindJSON(&commandReq); err != nil {
		log.Printf("❌ [chcommand] JSON bind error: %v", err)
		c.JSON(http.StatusBadRequest, models.CommandResponse{
			Success: false,
			Error:   msgInvalidJSON + ": " + err.Error(),
		})
		return
	}

	log.Printf("🎯 [chcommand] Executing ClickHouse command in database '%s': %s", commandReq.DatabaseName, commandReq.Query)

	ctx := c.Request.Context()

	// Get connection to specific database
	db, err := h.clickHouseService.GetDatabaseConnection(commandReq.DatabaseName)
	if err != nil {
		log.Printf("❌ [chcommand] Database connection failed: %v", err)
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
	result, err := db.ExecContext(ctx, commandReq.Query)
	duration := float64(time.Since(startTime).Nanoseconds()) / 1e6

	if err != nil {
		log.Printf("❌ [chcommand] Execution failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.CommandResponse{
			Success:  false,
			Database: commandReq.DatabaseName,
			Error:    fmt.Sprintf("ClickHouse command execution failed: %s", err.Error()),
			Command:  commandReq.Query,
			Duration: duration,
		})
		return
	}

	// Get rows affected if possible
	rowsAffected, err := result.RowsAffected()
	var resultData interface{}
	if err != nil {
		// Some commands might not return rows affected, return basic success
		resultData = map[string]interface{}{
			"status": "success",
			"query":  commandReq.Query,
		}
	} else {
		resultData = map[string]interface{}{
			"status":        "success",
			"rows_affected": rowsAffected,
			"query":         commandReq.Query,
		}
	}

	log.Printf("✅ [chcommand] Execution successful in %.2fms", duration)

	c.JSON(http.StatusOK, models.CommandResponse{
		Success:  true,
		Message:  "ClickHouse command executed successfully",
		Result:   resultData,
		Database: commandReq.DatabaseName,
		Command:  commandReq.Query,
		Duration: duration,
	})
}

// ChSelectEndpoint godoc
// @Summary Execute ClickHouse SELECT query
// @Description Execute ClickHouse SELECT query and return data via JSON request
// @Tags clickhouse
// @Accept json
// @Produce json
// @Param select body models.SelectRequest true "SELECT query to execute"
// @Success 200 {object} models.SelectResponse
// @Router /chselect [post]
func (h *APIHandler) ChSelectEndpoint(c *gin.Context) {
	startTime := time.Now()

	// Check if ClickHouse service is available
	if h.clickHouseService == nil {
		log.Printf("❌ [chselect] ClickHouse service is not available")
		c.JSON(http.StatusServiceUnavailable, models.SelectResponse{
			Success: false,
			Error:   "ClickHouse service is not available",
		})
		return
	}

	// Parse JSON request
	var selectReq models.SelectRequest
	if err := c.ShouldBindJSON(&selectReq); err != nil {
		log.Printf("❌ [chselect] JSON bind error: %v", err)
		c.JSON(http.StatusBadRequest, models.SelectResponse{
			Success: false,
			Error:   msgInvalidJSON + ": " + err.Error(),
		})
		return
	}

	log.Printf("🎯 [chselect] Executing ClickHouse query in database '%s': %s", selectReq.DatabaseName, selectReq.Query)

	ctx := c.Request.Context()

	// Get connection to specific database
	db, err := h.clickHouseService.GetDatabaseConnection(selectReq.DatabaseName)
	if err != nil {
		log.Printf("❌ [chselect] Database connection failed: %v", err)
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
	rows, err := db.QueryContext(ctx, selectReq.Query)
	if err != nil {
		log.Printf("❌ [chselect] Query failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.SelectResponse{
			Success:  false,
			Database: selectReq.DatabaseName,
			Error:    fmt.Sprintf("ClickHouse query execution failed: %s", err.Error()),
			Query:    selectReq.Query,
			Duration: float64(time.Since(startTime).Nanoseconds()) / 1e6,
		})
		return
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		log.Printf("❌ [chselect] Failed to get columns: %v", err)
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
			log.Printf("❌ [chselect] Row scan failed: %v", err)
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
		log.Printf("❌ [chselect] Row iteration error: %v", err)
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
	log.Printf("✅ [chselect] Query successful, returned %d rows in %.2fms", len(data), duration)

	c.JSON(http.StatusOK, models.SelectResponse{
		Success:    true,
		Message:    "ClickHouse query executed successfully",
		Data:       data,
		RowCount:   len(data),
		Columns:    columns,
		Database:   selectReq.DatabaseName,
		Query:      selectReq.Query,
		Duration:   duration,
		ServerTime: time.Now().Format(time.RFC3339),
	})
}

// ChSelectGroupEndpoint godoc
// @Summary Execute multiple ClickHouse SELECT queries
// @Description Execute multiple ClickHouse SELECT queries in a single request
// @Tags clickhouse
// @Accept json
// @Produce json
// @Param selectgroup body models.GroupSelectRequest true "Multiple SELECT queries to execute"
// @Success 200 {object} models.GroupSelectResponse
// @Router /chselectgroup [post]
func (h *APIHandler) ChSelectGroupEndpoint(c *gin.Context) {
	startTime := time.Now()

	// Check if ClickHouse service is available
	if h.clickHouseService == nil {
		log.Printf("❌ [chselectgroup] ClickHouse service is not available")
		c.JSON(http.StatusServiceUnavailable, models.GroupSelectResponse{
			Success: false,
			Error:   "ClickHouse service is not available",
		})
		return
	}

	// Parse JSON request
	var groupReq models.GroupSelectRequest
	if err := c.ShouldBindJSON(&groupReq); err != nil {
		log.Printf("❌ [chselectgroup] JSON bind error: %v", err)
		c.JSON(http.StatusBadRequest, models.GroupSelectResponse{
			Success: false,
			Error:   msgInvalidJSON + ": " + err.Error(),
		})
		return
	}

	log.Printf("🎯 [chselectgroup] Executing %d ClickHouse queries in database '%s'", len(groupReq.Queries), groupReq.DatabaseName)

	var results []models.GroupSelectQueryResult

	// Execute each query
	for i, query := range groupReq.Queries {
		queryStartTime := time.Now()

		// Get connection to specific database
		db, err := h.clickHouseService.GetDatabaseConnection(groupReq.DatabaseName)
		if err != nil {
			log.Printf("❌ [chselectgroup] Database connection failed for query %d: %v", i+1, err)
			results = append(results, models.GroupSelectQueryResult{
				QueryIndex: i + 1,
				Success:    false,
				Error:      fmt.Sprintf("Database connection failed: %s", err.Error()),
				Duration:   float64(time.Since(queryStartTime).Nanoseconds()) / 1e6,
			})
			continue
		}

		// Execute select query
		ctx := c.Request.Context()
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			log.Printf("❌ [chselectgroup] Query %d failed: %v", i+1, err)
			db.Close()
			results = append(results, models.GroupSelectQueryResult{
				QueryIndex: i + 1,
				Success:    false,
				Error:      fmt.Sprintf("ClickHouse query execution failed: %s", err.Error()),
				Query:      query,
				Duration:   float64(time.Since(queryStartTime).Nanoseconds()) / 1e6,
			})
			continue
		}

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			log.Printf("❌ [chselectgroup] Failed to get columns for query %d: %v", i+1, err)
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
				log.Printf("❌ [chselectgroup] Row scan failed for query %d: %v", i+1, err)
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
			log.Printf("❌ [chselectgroup] Row iteration error for query %d: %v", i+1, err)
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
		log.Printf("✅ [chselectgroup] Query %d successful, returned %d rows in %.2fms", i+1, len(data), duration)

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

	log.Printf("✅ [chselectgroup] Completed %d/%d queries successfully in %.2fms", successCount, len(groupReq.Queries), totalDuration)

	c.JSON(http.StatusOK, models.GroupSelectResponse{
		Success:      true,
		Message:      fmt.Sprintf("ClickHouse group query completed: %d/%d successful", successCount, len(groupReq.Queries)),
		Results:      results,
		TotalQueries: len(groupReq.Queries),
		Successful:   successCount,
		Failed:       len(groupReq.Queries) - successCount,
		Database:     groupReq.DatabaseName,
		Duration:     totalDuration,
		ServerTime:   time.Now().Format(time.RFC3339),
	})
}

// ChCheckDatabaseEndpoint godoc
// @Summary Check ClickHouse database existence
// @Description Check if a ClickHouse database exists and get database information
// @Tags clickhouse
// @Accept json
// @Produce json
// @Param checkdb body models.CheckDatabaseRequest true "Database to check"
// @Success 200 {object} models.CheckDatabaseResponse
// @Router /chcheckdatabase [post]
func (h *APIHandler) ChCheckDatabaseEndpoint(c *gin.Context) {
	startTime := time.Now()

	// Check if ClickHouse service is available
	if h.clickHouseService == nil {
		log.Printf("❌ [chcheckdatabase] ClickHouse service is not available")
		c.JSON(http.StatusServiceUnavailable, models.CheckDatabaseResponse{
			Success: false,
			Error:   "ClickHouse service is not available",
		})
		return
	}

	// Parse JSON request
	var checkReq models.CheckDatabaseRequest
	if err := c.ShouldBindJSON(&checkReq); err != nil {
		log.Printf("❌ [chcheckdatabase] JSON bind error: %v", err)
		c.JSON(http.StatusBadRequest, models.CheckDatabaseResponse{
			Success: false,
			Error:   msgInvalidJSON + ": " + err.Error(),
		})
		return
	}

	log.Printf("🎯 [chcheckdatabase] Checking ClickHouse database: %s", checkReq.DatabaseName)

	// Check if database exists
	exists, err := h.clickHouseService.CheckDatabaseExists(checkReq.DatabaseName)
	if err != nil {
		log.Printf("❌ [chcheckdatabase] Database check failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.CheckDatabaseResponse{
			Success:  false,
			Database: checkReq.DatabaseName,
			Error:    fmt.Sprintf("ClickHouse database check failed: %s", err.Error()),
			Duration: float64(time.Since(startTime).Nanoseconds()) / 1e6,
		})
		return
	}

	var tables []string
	if exists {
		// Get table list if database exists
		tables, err = h.clickHouseService.GetTableList(checkReq.DatabaseName)
		if err != nil {
			log.Printf("❌ [chcheckdatabase] Failed to get table list: %v", err)
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
	log.Printf("✅ [chcheckdatabase] Database check completed in %.2fms: exists=%v, tables=%d", duration, exists, len(tables))

	c.JSON(http.StatusOK, models.CheckDatabaseResponse{
		Success:    true,
		Message:    fmt.Sprintf("ClickHouse database check completed: %s", checkReq.DatabaseName),
		Database:   checkReq.DatabaseName,
		Exists:     exists,
		Tables:     tables,
		TableCount: len(tables),
		Duration:   duration,
		ServerTime: time.Now().Format(time.RFC3339),
	})
}

// ChCheckTableEndpoint godoc
// @Summary Check ClickHouse table existence
// @Description Check if a ClickHouse table exists in a specific database
// @Tags clickhouse
// @Accept json
// @Produce json
// @Param checktable body models.CheckTableRequest true "Database and table to check"
// @Success 200 {object} models.CheckTableResponse
// @Router /chchecktable [post]
func (h *APIHandler) ChCheckTableEndpoint(c *gin.Context) {
	startTime := time.Now()

	// Check if ClickHouse service is available
	if h.clickHouseService == nil {
		log.Printf("❌ [chchecktable] ClickHouse service is not available")
		c.JSON(http.StatusServiceUnavailable, models.CheckTableResponse{
			Success: false,
			Error:   "ClickHouse service is not available",
		})
		return
	}

	// Parse JSON request
	var checkReq models.CheckTableRequest
	if err := c.ShouldBindJSON(&checkReq); err != nil {
		log.Printf("❌ [chchecktable] JSON bind error: %v", err)
		c.JSON(http.StatusBadRequest, models.CheckTableResponse{
			Success: false,
			Error:   msgInvalidJSON + ": " + err.Error(),
		})
		return
	}

	log.Printf("🎯 [chchecktable] Checking ClickHouse table '%s' in database '%s'", checkReq.TableName, checkReq.DatabaseName)

	// Check if table exists
	exists, err := h.clickHouseService.CheckTableExists(checkReq.DatabaseName, checkReq.TableName)
	if err != nil {
		log.Printf("❌ [chchecktable] Table check failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.CheckTableResponse{
			Success:  false,
			Database: checkReq.DatabaseName,
			Table:    checkReq.TableName,
			Error:    fmt.Sprintf("ClickHouse table check failed: %s", err.Error()),
			Duration: float64(time.Since(startTime).Nanoseconds()) / 1e6,
		})
		return
	}

	duration := float64(time.Since(startTime).Nanoseconds()) / 1e6
	log.Printf("✅ [chchecktable] Table check completed in %.2fms: exists=%v", duration, exists)

	c.JSON(http.StatusOK, models.CheckTableResponse{
		Success:    true,
		Message:    fmt.Sprintf("ClickHouse table check completed: %s.%s", checkReq.DatabaseName, checkReq.TableName),
		Database:   checkReq.DatabaseName,
		Table:      checkReq.TableName,
		Exists:     exists,
		Duration:   duration,
		ServerTime: time.Now().Format(time.RFC3339),
	})
}
