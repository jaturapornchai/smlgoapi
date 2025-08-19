package models

import "time"

// Table represents a database table
type Table struct {
	Name string `json:"name" db:"name"`
}

// ICUnitUse represents the ic_unit_use table structure
type ICUnitUse struct {
	ICCode      string   `json:"ic_code" db:"ic_code"`
	Code        string   `json:"code" db:"code"`
	Name1       *string  `json:"name_1,omitempty" db:"name_1"`
	LineNumber  *int     `json:"line_number,omitempty" db:"line_number"`
	StandValue  *float64 `json:"stand_value,omitempty" db:"stand_value"`
	DivideValue *float64 `json:"divide_value,omitempty" db:"divide_value"`
	RowOrderRef *int     `json:"row_order_ref,omitempty" db:"row_order_ref"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
	Database  string    `json:"database"`
}

// APIResponse represents a generic API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SearchParameters represents all search parameters in JSON format
type SearchParameters struct {
	Query  string `json:"query" binding:"required"` // actual search text (not base64)
	Limit  int    `json:"limit,omitempty"`          // number of results
	Offset int    `json:"offset,omitempty"`         // pagination offset
	AI     int    `json:"ai,omitempty"`             // AI mode: 0=no AI, 1=use AI to enhance query
}

// SearchResult represents a search result item
type SearchResult struct {
	ID              interface{} `json:"id"`
	Code            string      `json:"code,omitempty"`
	Name            string      `json:"name"`
	Brand           string      `json:"brand,omitempty"`
	Category        string      `json:"category,omitempty"`
	Description     string      `json:"description,omitempty"`
	Barcode         string      `json:"barcode,omitempty"`
	Unit            string      `json:"unit,omitempty"`
	ImgURL          string      `json:"img_url,omitempty"`
	ImagePath       string      `json:"image_path,omitempty"`
	SimilarityScore float64     `json:"similarity_score"`
}

// VectorSearchResponse represents the response from vector search
type VectorSearchResponse struct {
	Success      bool           `json:"success"`
	Message      string         `json:"message,omitempty"`
	Data         []SearchResult `json:"data"`
	TotalCount   int            `json:"total_count"`
	Query        string         `json:"query"`
	Duration     float64        `json:"duration_ms"`
	SearchMethod string         `json:"search_method,omitempty"`
	Error        string         `json:"error,omitempty"`
}

// CommandRequest represents a command request for executing SQL commands
type CommandRequest struct {
	DatabaseName string `json:"database_name" binding:"required"` // Target database name
	Query        string `json:"query" binding:"required"`         // SQL command to execute
}

// CommandResponse represents the response from command execution
type CommandResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message,omitempty"`
	Result   interface{} `json:"result,omitempty"`
	Database string      `json:"database,omitempty"`
	Command  string      `json:"command,omitempty"`
	Duration float64     `json:"duration_ms"`
	Error    string      `json:"error,omitempty"`
}

// SelectRequest represents a select query request
type SelectRequest struct {
	DatabaseName string `json:"database_name" binding:"required"` // Target database name
	Query        string `json:"query" binding:"required"`         // SELECT query to execute
}

// SelectResponse represents the response from select query
type SelectResponse struct {
	Success    bool          `json:"success"`
	Message    string        `json:"message,omitempty"`
	Data       []interface{} `json:"data,omitempty"`
	Database   string        `json:"database,omitempty"`
	Query      string        `json:"query,omitempty"`
	RowCount   int           `json:"row_count"`
	Columns    []string      `json:"columns,omitempty"`
	Duration   float64       `json:"duration_ms"`
	ServerTime string        `json:"server_time,omitempty"`
	Error      string        `json:"error,omitempty"`
}

// GroupSelectRequest represents a batch of select queries to run in a single request
type GroupSelectRequest struct {
	DatabaseName string   `json:"database_name" binding:"required"` // Target database name
	Queries      []string `json:"queries" binding:"required"`       // List of SELECT queries to execute
}

// GroupSelectQueryResult holds result for a single query in the batch
type GroupSelectQueryResult struct {
	QueryIndex int           `json:"query_index"`
	Query      string        `json:"query"`
	Success    bool          `json:"success"`
	RowCount   int           `json:"row_count"`
	Columns    []string      `json:"columns,omitempty"`
	Duration   float64       `json:"duration_ms"`
	Data       []interface{} `json:"data,omitempty"`
	Error      string        `json:"error,omitempty"`
}

// GroupSelectResponse represents the aggregated response for batch select
type GroupSelectResponse struct {
	Success      bool                     `json:"success"`
	Message      string                   `json:"message,omitempty"`
	Database     string                   `json:"database"`
	TotalQueries int                      `json:"total_queries"`
	Successful   int                      `json:"successful"`
	Failed       int                      `json:"failed"`
	Duration     float64                  `json:"duration_ms"`
	ServerTime   string                   `json:"server_time,omitempty"`
	Results      []GroupSelectQueryResult `json:"results"`
	Error        string                   `json:"error,omitempty"`
	Partial      bool                     `json:"partial,omitempty"` // true if some failed
}

// DatabaseRequest represents a request to check/create database
type DatabaseRequest struct {
	DatabaseName string `json:"database_name" binding:"required"`
}

// DatabaseResponse represents the response from database check/create operation
type DatabaseResponse struct {
	Success  bool    `json:"success"`
	Status   string  `json:"status"` // "exists" or "created"
	Database string  `json:"database"`
	Message  string  `json:"message"`
	Duration float64 `json:"duration_ms"`
	Error    string  `json:"error,omitempty"`
}

// TableCheckRequest represents a request to check/create table from SQL scripts
type TableCheckRequest struct {
	DatabaseName string `json:"database_name" binding:"required"`
	TableName    string `json:"table_name"` // Optional: if empty, scan all SQL files
}

// TableCheckResponse represents the response from table check/create operation
type TableCheckResponse struct {
	Success     bool     `json:"success"`
	Status      string   `json:"status"` // "exists", "created", "script_not_found"
	TableName   string   `json:"table_name"`
	Database    string   `json:"database"`
	Message     string   `json:"message"`
	ScriptPath  string   `json:"script_path,omitempty"`
	TablesFound []string `json:"tables_found,omitempty"`
	Duration    float64  `json:"duration_ms"`
	Error       string   `json:"error,omitempty"`
}

// CheckDatabaseRequest represents a request to check database existence
type CheckDatabaseRequest struct {
	DatabaseName string `json:"database_name" binding:"required"`
}

// CheckDatabaseResponse represents the response from database check operation
type CheckDatabaseResponse struct {
	Success    bool     `json:"success"`
	Message    string   `json:"message,omitempty"`
	Database   string   `json:"database"`
	Exists     bool     `json:"exists"`
	Tables     []string `json:"tables,omitempty"`
	TableCount int      `json:"table_count"`
	Duration   float64  `json:"duration_ms"`
	ServerTime string   `json:"server_time,omitempty"`
	Error      string   `json:"error,omitempty"`
}

// CheckTableRequest represents a request to check table existence
type CheckTableRequest struct {
	DatabaseName string `json:"database_name" binding:"required"`
	TableName    string `json:"table_name" binding:"required"`
}

// CheckTableResponse represents the response from table check operation
type CheckTableResponse struct {
	Success    bool    `json:"success"`
	Message    string  `json:"message,omitempty"`
	Database   string  `json:"database"`
	Table      string  `json:"table"`
	Exists     bool    `json:"exists"`
	Duration   float64 `json:"duration_ms"`
	ServerTime string  `json:"server_time,omitempty"`
	Error      string  `json:"error,omitempty"`
}
