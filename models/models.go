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

// CommandRequest represents a command request for executing SQL commands
type CommandRequest struct {
	Query string `json:"query" binding:"required"` // SQL command to execute
}

// CommandResponse represents the response from command execution
type CommandResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message,omitempty"`
	Result   interface{} `json:"result,omitempty"`
	Command  string      `json:"command,omitempty"`
	Duration float64     `json:"duration_ms"`
	Error    string      `json:"error,omitempty"`
}

// SelectRequest represents a select query request
type SelectRequest struct {
	Query string `json:"query" binding:"required"` // SELECT query to execute
}

// SelectResponse represents the response from select query
type SelectResponse struct {
	Success  bool          `json:"success"`
	Message  string        `json:"message,omitempty"`
	Data     []interface{} `json:"data,omitempty"`
	Query    string        `json:"query,omitempty"`
	RowCount int           `json:"row_count"`
	Duration float64       `json:"duration_ms"`
	Error    string        `json:"error,omitempty"`
}

// Thai Administrative Data Models

// Province represents a Thai province
type Province struct {
	ID          int     `json:"id"`
	NameTh      string  `json:"name_th"`
	NameEn      string  `json:"name_en"`
	GeographyID int     `json:"geography_id,omitempty"`
	CreatedAt   string  `json:"created_at,omitempty"`
	UpdatedAt   string  `json:"updated_at,omitempty"`
	DeletedAt   *string `json:"deleted_at,omitempty"`
}

// Amphure represents a Thai district (amphure)
type Amphure struct {
	ID         int     `json:"id"`
	NameTh     string  `json:"name_th"`
	NameEn     string  `json:"name_en"`
	ProvinceID int     `json:"province_id"`
	CreatedAt  string  `json:"created_at,omitempty"`
	UpdatedAt  string  `json:"updated_at,omitempty"`
	DeletedAt  *string `json:"deleted_at,omitempty"`
}

// Tambon represents a Thai sub-district (tambon)
type Tambon struct {
	ID        int     `json:"id"`
	NameTh    string  `json:"name_th"`
	NameEn    string  `json:"name_en"`
	ZipCode   int     `json:"zip_code,omitempty"`
	AmphureID int     `json:"amphure_id"`
	CreatedAt string  `json:"created_at,omitempty"`
	UpdatedAt string  `json:"updated_at,omitempty"`
	DeletedAt *string `json:"deleted_at,omitempty"`
}

// ProvinceRequest represents a request for province data
type ProvinceRequest struct {
	// Empty for now, but can be extended later
}

// AmphureRequest represents a request for amphure data
type AmphureRequest struct {
	ProvinceID int `json:"province_id" binding:"required"`
}

// TambonRequest represents a request for tambon data
type TambonRequest struct {
	AmphureID  int `json:"amphure_id" binding:"required"`
	ProvinceID int `json:"province_id" binding:"required"`
}

// ZipCodeRequest represents a request to find location by zip code
type ZipCodeRequest struct {
	ZipCode int `json:"zip_code" binding:"required"`
}

// CompleteLocationData represents complete location information with nested structure
type CompleteLocationData struct {
	Province Province `json:"province"`
	Amphure  Amphure  `json:"amphure"`
	Tambon   Tambon   `json:"tambon"`
}

// TambonWithNested represents the structure of tambon data from the JSON file
type TambonWithNested struct {
	ID        int               `json:"id"`
	ZipCode   int               `json:"zip_code"`
	NameTh    string            `json:"name_th"`
	NameEn    string            `json:"name_en"`
	AmphureID int               `json:"amphure_id"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
	DeletedAt *string           `json:"deleted_at"`
	Amphure   AmphureWithNested `json:"amphure"`
}

// AmphureWithNested represents the nested amphure structure
type AmphureWithNested struct {
	ID         int            `json:"id"`
	NameTh     string         `json:"name_th"`
	NameEn     string         `json:"name_en"`
	ProvinceID int            `json:"province_id"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
	DeletedAt  *string        `json:"deleted_at"`
	Province   ProvinceNested `json:"province"`
}

// ProvinceNested represents the nested province structure
type ProvinceNested struct {
	ID          int     `json:"id"`
	NameTh      string  `json:"name_th"`
	NameEn      string  `json:"name_en"`
	GeographyID int     `json:"geography_id"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	DeletedAt   *string `json:"deleted_at"`
}
