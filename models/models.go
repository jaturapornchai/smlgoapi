package models

import "time"

// Table represents a database table
type Table struct {
	Name string `json:"name" db:"name"`
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
}

// SearchRequest represents a vector search request (for backward compatibility)
type SearchRequest struct {
	Query  string `json:"query" form:"query" binding:"required" example:"aGVsbG8gd29ybGQ="` // base64 encoded query
	Limit  int    `json:"limit" form:"limit" example:"10"`
	Offset int    `json:"offset" form:"offset" example:"0"`
}
