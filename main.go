package main

import (
	"context"
	"database/sql"
	"encoding/json" // Added this line
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
)

// ===== CONFIGURATION =====
type Config struct {
	ServerPort          string
	ClickHouseHost      string
	ClickHousePort      string
	ClickHouseUser      string
	ClickHousePass      string
	ClickHouseDB        string
	ClickHouseSecure    bool
	CacheEnabled        bool
	MaxWorkers          int
	CacheTTL            time.Duration
	SimilarityThreshold float64
	// Debug configuration
	DebugMode          bool
	DebugLevel         int // 0=OFF, 1=ERROR, 2=WARN, 3=INFO, 4=DEBUG, 5=TRACE
	LogStepByStep      bool
	LogSQLExecution    bool
	LogRequestResponse bool
	LogPerformance     bool
	// Timeout configuration
	RequestTimeout  time.Duration
	SQLTimeout      time.Duration
	DatabaseTimeout time.Duration
	HttpTimeout     time.Duration
}

var config Config

// Debug levels
const (
	DEBUG_OFF   = 0
	DEBUG_ERROR = 1
	DEBUG_WARN  = 2
	DEBUG_INFO  = 3
	DEBUG_DEBUG = 4
	DEBUG_TRACE = 5
)

// Debug step tracking
type DebugStep struct {
	StepNumber int                    `json:"step_number"`
	StepName   string                 `json:"step_name"`
	Status     string                 `json:"status"` // STARTED, SUCCESS, ERROR, SKIPPED
	StartTime  time.Time              `json:"start_time"`
	EndTime    *time.Time             `json:"end_time,omitempty"`
	Duration   string                 `json:"duration,omitempty"`
	Input      interface{}            `json:"input,omitempty"`
	Output     interface{}            `json:"output,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

type DebugTrace struct {
	RequestID   int64       `json:"request_id"`
	Method      string      `json:"method"`
	Endpoint    string      `json:"endpoint"`
	StartTime   time.Time   `json:"start_time"`
	Steps       []DebugStep `json:"steps"`
	TotalSteps  int         `json:"total_steps"`
	Completed   bool        `json:"completed"`
	FinalStatus string      `json:"final_status"`
	TotalTime   string      `json:"total_time"`
}

// ===== GLOBAL VARIABLES =====
var (
	clickhouseDB   *sql.DB
	localCache     *cache.Cache // Re-added
	requestCounter int64
	requestMutex   sync.Mutex // Re-added
	workerPool     chan bool  // Re-added
	stats          *ServerStats
	vectorDB       *TFIDFVectorDatabase
	// Debug tracking
	debugTraces map[int64]*DebugTrace
	debugMutex  sync.RWMutex
)

// ===== DATA STRUCTURES =====

// Request/Response Models
type SearchRequest struct {
	Query  string `json:"query" binding:"required"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

type SearchResponse struct {
	TotalCount int           `json:"total_count"`
	Data       []interface{} `json:"data"`
	Offset     int           `json:"offset"`
	Limit      int           `json:"limit"`
}

type CommandGetRequest struct {
	QueryBase64 string `json:"query_base64" binding:"required"`
}

type CommandPostRequest struct {
	QueryBase64 string `json:"query_base64" binding:"required"`
}

type CommandResponse struct {
	Result     interface{} `json:"result"`
	Command    string      `json:"command"`
	DecodedSQL string      `json:"decoded_sql"`
	Method     string      `json:"method"`
}

type ImageUploadRequest struct {
	Barcode      string `json:"barcode" binding:"required"`
	ImageNumber  int    `json:"imagenumber"`
	ImageData    string `json:"image_data" binding:"required"`
	UseMultiView bool   `json:"use_multi_view"`
}

type ImageUploadResponse struct {
	Status              string  `json:"status"`
	Message             string  `json:"message"`
	Barcode             string  `json:"barcode"`
	ImageNumber         int     `json:"imagenumber"`
	TotalViewsGenerated int     `json:"total_views_generated,omitempty"`
	TotalVectorsStored  int     `json:"total_vectors_stored,omitempty"`
	VectorSize          int     `json:"vector_size,omitempty"`
	ProcessingTimeMS    float64 `json:"processing_time_ms"`
}

type ImageSearchRequest struct {
	ImageData           string  `json:"image_data" binding:"required"`
	SimilarityThreshold float64 `json:"similarity_threshold"`
	Limit               int     `json:"limit"`
	UseMultiView        bool    `json:"use_multi_view"`
}

type ImageSearchResult struct {
	Barcode         string  `json:"barcode"`
	ImageNumber     int     `json:"imagenumber"`
	SimilarityScore float64 `json:"similarity_score"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
}

type ImageSearchResponse struct {
	TotalFound       int                 `json:"total_found"`
	Results          []ImageSearchResult `json:"results"`
	QueryVectorSize  int                 `json:"query_vector_size"`
	ProcessingTimeMS float64             `json:"processing_time_ms"`
}

// Server Stats
type Performance struct {
	RequestsPerSecond   float64 `json:"requests_per_second"`
	AverageResponseTime float64 `json:"average_response_time_ms"`
	TotalRequests       int64   `json:"total_requests"`
	ErrorRate           float64 `json:"error_rate"`
}

type CacheInfo struct {
	HitRate     float64 `json:"hit_rate"`
	MissRate    float64 `json:"miss_rate"`
	TotalHits   int64   `json:"total_hits"`
	TotalMisses int64   `json:"total_misses"`
	ItemCount   int     `json:"item_count"`
}

type ServerStats struct {
	StartTime         time.Time   `json:"start_time"`
	Uptime            string      `json:"uptime"`
	TotalRequests     int64       `json:"total_requests"`
	ActiveConnections int         `json:"active_connections"`
	Performance       Performance `json:"performance"`
	Cache             CacheInfo   `json:"cache"`
	Memory            string      `json:"memory_usage"`
}

type HealthResponse struct {
	Status       string `json:"status"`
	Message      string `json:"message"`
	Timestamp    string `json:"timestamp"`
	RequestCount int64  `json:"request_count"`
}

// Vector Database Mock
type TFIDFVectorDatabase struct {
	mu sync.RWMutex
}

func NewTFIDFVectorDatabase() *TFIDFVectorDatabase {
	return &TFIDFVectorDatabase{}
}

func (db *TFIDFVectorDatabase) CreateVectorTable() error {
	log.Println("✅ [TFIDFVectorDatabase] Vector table created")
	return nil
}

func (db *TFIDFVectorDatabase) BuildVectorIndex() error {
	log.Println("✅ [TFIDFVectorDatabase] Vector index built")
	return nil
}

func (db *TFIDFVectorDatabase) SearchProducts(query string, limit, offset int) (string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Mock search results
	results := map[string]interface{}{
		"total_count": 2,
		"data": []map[string]interface{}{
			{
				"barcode":          "123456789",
				"name":             fmt.Sprintf("Product matching '%s' - Result 1", query),
				"description":      "High quality product with excellent features",
				"similarity_score": 0.95,
				"balance_qty":      100,
			},
			{
				"barcode":          "987654321",
				"name":             fmt.Sprintf("Product matching '%s' - Result 2", query),
				"description":      "Another great product option",
				"similarity_score": 0.87,
				"balance_qty":      50,
			},
		},
		"offset": offset,
		"limit":  limit,
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (db *TFIDFVectorDatabase) Close() {
	log.Println("🔌 [TFIDFVectorDatabase] Connection closed")
}

// ===== UTILITY FUNCTIONS =====
func loadConfig() {
	config = Config{
		ServerPort:          getEnv("PORT", "8008"),
		ClickHouseHost:      getEnv("CLICKHOUSE_HOST", "localhost"),
		ClickHousePort:      getEnv("CLICKHOUSE_PORT", "9000"),
		ClickHouseUser:      getEnv("CLICKHOUSE_USER", "default"),
		ClickHousePass:      getEnv("CLICKHOUSE_PASSWORD", ""),
		ClickHouseDB:        getEnv("CLICKHOUSE_DATABASE", "default"),
		ClickHouseSecure:    getEnv("CLICKHOUSE_SECURE", "false") == "true",
		CacheEnabled:        getEnv("CACHE_ENABLED", "true") == "true",
		MaxWorkers:          parseInt(getEnv("MAX_WORKERS", "100")),
		CacheTTL:            time.Duration(parseInt(getEnv("CACHE_TTL_MINUTES", "15"))) * time.Minute,
		SimilarityThreshold: parseFloat(getEnv("SIMILARITY_THRESHOLD", "0.25")),
		// Debug configuration
		DebugMode:          getEnv("DEBUG_MODE", "true") == "true",
		DebugLevel:         parseInt(getEnv("DEBUG_LEVEL", "4")), // Default to DEBUG level
		LogStepByStep:      getEnv("LOG_STEP_BY_STEP", "true") == "true",
		LogSQLExecution:    getEnv("LOG_SQL_EXECUTION", "true") == "true",
		LogRequestResponse: getEnv("LOG_REQUEST_RESPONSE", "true") == "true",
		LogPerformance:     getEnv("LOG_PERFORMANCE", "true") == "true",
		// Timeout configuration
		RequestTimeout:  time.Duration(parseInt(getEnv("REQUEST_TIMEOUT_SECONDS", "30"))) * time.Second,
		SQLTimeout:      time.Duration(parseInt(getEnv("SQL_TIMEOUT_SECONDS", "60"))) * time.Second,
		DatabaseTimeout: time.Duration(parseInt(getEnv("DATABASE_TIMEOUT_SECONDS", "10"))) * time.Second,
		HttpTimeout:     time.Duration(parseInt(getEnv("HTTP_TIMEOUT_SECONDS", "120"))) * time.Second,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return 0
}

func parseFloat(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return 0.0
}

// ===== UTILITY FUNCTIONS MOVED TO handler_utils.go =====

// ===== DEBUG UTILITY FUNCTIONS =====

func initDebug() {
	debugTraces = make(map[int64]*DebugTrace)

	if config.DebugMode {
		logDebug(DEBUG_INFO, "Debug mode enabled", map[string]interface{}{
			"debug_level":              config.DebugLevel,
			"step_logging":             config.LogStepByStep,
			"sql_logging":              config.LogSQLExecution,
			"request_response_logging": config.LogRequestResponse,
			"performance_logging":      config.LogPerformance,
		})
	}
}

func logDebug(level int, message string, details map[string]interface{}) {
	if !config.DebugMode || level > config.DebugLevel {
		return
	}

	levelNames := []string{"OFF", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}
	levelName := "UNKNOWN"
	if level < len(levelNames) {
		levelName = levelNames[level]
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	fmt.Printf("\n🔍 [DEBUG-%s] %s | %s\n", levelName, timestamp, message)
	for key, value := range details {
		fmt.Printf("    [DEBUG] %s: %v\n", key, value)
	}
}

// ===== DEBUG UTILITY FUNCTIONS MOVED TO handler_utils.go =====
// Functions addDebugStep, completeDebugStep, completeDebugTrace, getDebugTrace, logSQLExecution, logPerformanceMetrics were removed from here.

// ===== DATABASE FUNCTIONS =====
func initClickHouse() error {
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s?dial_timeout=10s&max_execution_time=60",
		config.ClickHouseUser,
		config.ClickHousePass,
		config.ClickHouseHost,
		config.ClickHousePort,
		config.ClickHouseDB,
	)

	var err error
	clickhouseDB, err = sql.Open("clickhouse", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := clickhouseDB.PingContext(ctx); err != nil {
		log.Printf("⚠️ Warning: ClickHouse ping failed: %v", err)
		// Continue without failing - will use mock data
	} else {
		log.Println("✅ ClickHouse connection established")
	}

	if err := initTables(); err != nil {
		log.Printf("⚠️ Warning: Failed to initialize tables: %v", err)
	}

	return nil
}

func initTables() error {
	if clickhouseDB == nil {
		return fmt.Errorf("clickhouse connection not established")
	}

	tables := []string{
		`CREATE TABLE IF NOT EXISTS product_vectors (
			barcode String,
			name String,
			description String,
			imagenumber Int32,
			image_vector String,
			created_at DateTime,
			updated_at DateTime
		) ENGINE = MergeTree() ORDER BY (barcode, imagenumber)`,

		`CREATE TABLE IF NOT EXISTS product_multi_view_vectors (
			barcode String,
			imagenumber Int32,
			view_type String,
			vector Array(Float32),
			image_hash String,
			is_primary Bool,
			quality_score Float32,
			view_weight Float32,
			metadata String
		) ENGINE = MergeTree() ORDER BY (barcode, imagenumber, view_type)`,

		`CREATE TABLE IF NOT EXISTS commands_log (
			id UInt64,
			command String,
			result String,
			executed_at DateTime
		) ENGINE = MergeTree() ORDER BY id`,
	}

	for _, table := range tables {
		if _, err := clickhouseDB.Exec(table); err != nil {
			log.Printf("Warning: Failed to create table: %v", err)
		}
	}

	log.Println("✅ Database tables initialized")
	return nil
}

// ===== VECTOR FUNCTIONS =====

// ===== HTTP HANDLERS =====
// handleRoot function moved to handler_root.go

// handleHealth function moved to handler_health.go

// handleSearch function moved to handler_search.go

// handleCommandGet function moved to handler_command_get.go

// handleCommandPost function moved to handler_command_post.go

// handleImageUpload function moved to handler_image_upload.go

// handleImageSearch function moved to handler_image_search.go

// Get local IP address
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// ===== MIDDLEWARE =====

func requestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Add timeout context
		ctx, cancel := context.WithTimeout(c.Request.Context(), config.RequestTimeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		// Continue with request
		c.Next()

		// Log request completion
		duration := time.Since(startTime)
		if config.LogPerformance {
			logDebug(DEBUG_INFO, fmt.Sprintf("Request completed: %s %s", c.Request.Method, c.Request.URL.Path), map[string]interface{}{
				"duration":  duration.String(),
				"status":    c.Writer.Status(),
				"client_ip": c.ClientIP(),
			})
		}
	}
}

func corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// ===== MAIN FUNCTION =====
func main() {
	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Warning: .env file not found")
	}

	loadConfig()

	// Initialize debug system
	initDebug()

	localIP := getLocalIP()
	fmt.Printf("%s\n", strings.Repeat("=", 80))
	fmt.Printf("[main] 🚀 ENHANCED VECTOR SEARCH API - TEST MODE WITH MULTI-VIEW\n")
	fmt.Printf("%s\n", strings.Repeat("=", 80))
	fmt.Printf("[main] 📅 Started: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("[main] 🌐 Host: 0.0.0.0:%s\n", config.ServerPort)
	fmt.Printf("[main] 🖥️  Local IP: %s\n", localIP)
	fmt.Printf("[main] 📍 Health: http://%s:%s/health\n", localIP, config.ServerPort)
	fmt.Printf("[main] 📚 Help: http://%s:%s/help\n", localIP, config.ServerPort)
	fmt.Printf("[main] 🔍 Search: http://%s:%s/search\n", localIP, config.ServerPort)
	fmt.Printf("[main] 💻 Command GET: http://%s:%s/commandget\n", localIP, config.ServerPort)
	fmt.Printf("[main] 💻 Command POST: http://%s:%s/commandpost\n", localIP, config.ServerPort)
	fmt.Printf("[main] 📸 Image Upload: http://%s:%s/imgupload\n", localIP, config.ServerPort)
	fmt.Printf("[main] 🖼️ Image Search: http://%s:%s/imgsearch\n", localIP, config.ServerPort)
	fmt.Printf("%s\n", strings.Repeat("=", 80))

	// Debug configuration info
	if config.DebugMode {
		fmt.Printf("[main] 🔍 DEBUG MODE ENABLED\n")
		fmt.Printf("[main] 📊 Debug Level: %d\n", config.DebugLevel)
		fmt.Printf("[main] 📝 Step-by-step Logging: %t\n", config.LogStepByStep)
		fmt.Printf("[main] 🗄️ SQL Execution Logging: %t\n", config.LogSQLExecution)
		fmt.Printf("[main] 📤 Request/Response Logging: %t\n", config.LogRequestResponse)
		fmt.Printf("[main] 📊 Performance Logging: %t\n", config.LogPerformance)
		fmt.Printf("%s\n", strings.Repeat("=", 80))
	}

	fmt.Printf("[main] 💡 TIP: Enhanced with Multi-View capabilities for better accuracy\n")
	fmt.Printf("[main] 🔧 Set use_multi_view=false in requests for legacy single-view mode\n")
	fmt.Printf("[main] 🛑 Press Ctrl+C to stop the server\n")
	fmt.Printf("%s\n", strings.Repeat("=", 80))

	// Initialize stats
	stats = &ServerStats{
		StartTime: time.Now(),
		Performance: Performance{
			RequestsPerSecond:   0,
			AverageResponseTime: 0,
			TotalRequests:       0,
			ErrorRate:           0,
		},
		Cache: CacheInfo{
			HitRate:     0,
			MissRate:    0,
			TotalHits:   0,
			TotalMisses: 0,
			ItemCount:   0,
		},
	}

	// Initialize cache
	if config.CacheEnabled {
		localCache = cache.New(config.CacheTTL, config.CacheTTL*2)
		log.Println("✅ Local cache initialized")
	}

	// Initialize vector database
	vectorDB = NewTFIDFVectorDatabase()
	log.Println("⏳ กำลังเริ่มต้น API server...")
	time.Sleep(2 * time.Second) // Wait for ClickHouse to be ready

	vectorDB.CreateVectorTable()
	vectorDB.BuildVectorIndex()
	log.Println("✅ API Server เริ่มต้นสำเร็จ!")

	// Initialize database
	if err := initClickHouse(); err != nil {
		log.Printf("⚠️ Warning: ClickHouse initialization failed: %v", err)
	}

	// Setup Gin
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(requestMiddleware()) // Routes
	r.GET("/", handleRoot)
	r.GET("/health", handleHealth)
	r.GET("/help", handleHelp)
	r.POST("/search", handleSearch)
	r.GET("/commandget", handleCommandGet)
	r.POST("/commandpost", handleCommandPost)
	r.POST("/imgupload", handleImageUpload)
	r.POST("/imgsearch", handleImageSearch)

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("[main] 📊 REAL-TIME API MONITORING\n")
	fmt.Printf("%s\n", strings.Repeat("=", 80))

	// Start server
	log.Printf("🚀 Server starting on port %s", config.ServerPort)

	if err := r.Run(":" + config.ServerPort); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
