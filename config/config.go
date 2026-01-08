package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// Default configuration values
const (
	defaultHost    = "0.0.0.0"
	defaultPort    = "8008"
	defaultDBHost  = "localhost"
	defaultDBPort  = "5432"
	defaultDBUser  = "postgres"
	defaultDBName  = "postgres"
	defaultSSLMode = "disable"
	configFileName = "smlgoapi.json"
	envFileName    = ".env"
)

// Log messages
const (
	msgConfigNotFound  = "📄 %s not found, falling back to environment variables"
	msgLoadingFromEnv  = "📄 Loading configuration from environment variables"
	msgLoadingFromJSON = "📄 Loading configuration from %s"
	msgEnvLoadError    = "❌ Error loading .env file: %v"
	msgJSONReadError   = "❌ Error reading JSON config: %v"
	msgJSONParseError  = "❌ Error parsing JSON config: %v"
)

type Config struct {
	Server struct {
		Port string `json:"port"`
		Host string `json:"host"`
	} `json:"server"`
	PostgreSQL struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
		SSLMode  string `json:"sslmode"`
	} `json:"postgresql"`
	ClickHouse struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
	} `json:"clickhouse"`
	// SMTP configuration (replaces Email.APIKey)
	SMTP struct {
		Host      string `json:"host"`
		Port      int    `json:"port"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		FromEmail string `json:"from_email"`
		FromName  string `json:"from_name"`
	} `json:"smtp"`
}

// JSONConfig represents the structure of smlgoapi.json
type JSONConfig struct {
	Server struct {
		Host string `json:"host"`
		Port string `json:"port"`
	} `json:"server"`
	PostgreSQL struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
		SSLMode  string `json:"sslmode"`
	} `json:"postgresql"`
	ClickHouse struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
	} `json:"clickhouse"`
	SMTP struct {
		Host      string `json:"host"`
		Port      int    `json:"port"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		FromEmail string `json:"from_email"`
		FromName  string `json:"from_name"`
	} `json:"smtp"`
	// Alternative field name for backward compatibility
	Postgres struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
		Secure   bool   `json:"secure"`
	} `json:"postgres"`
}

func LoadConfig() *Config {
	config := &Config{}
	// Try to load from smlgoapi.json first
	if jsonConfig := loadJSONConfig(); jsonConfig != nil {
		log.Println("📄 Loading configuration from smlgoapi.json")
		config.Server.Host = jsonConfig.Server.Host
		config.Server.Port = jsonConfig.Server.Port

		// Support both "postgresql" and "postgres" field names
		if jsonConfig.PostgreSQL.Host != "" {
			config.PostgreSQL.Host = jsonConfig.PostgreSQL.Host
			config.PostgreSQL.Port = jsonConfig.PostgreSQL.Port
			config.PostgreSQL.User = jsonConfig.PostgreSQL.User
			config.PostgreSQL.Password = jsonConfig.PostgreSQL.Password
			config.PostgreSQL.Database = jsonConfig.PostgreSQL.Database
			config.PostgreSQL.SSLMode = jsonConfig.PostgreSQL.SSLMode
		} else if jsonConfig.Postgres.Host != "" {
			config.PostgreSQL.Host = jsonConfig.Postgres.Host
			config.PostgreSQL.Port = jsonConfig.Postgres.Port
			config.PostgreSQL.User = jsonConfig.Postgres.User
			config.PostgreSQL.Password = jsonConfig.Postgres.Password
			config.PostgreSQL.Database = jsonConfig.Postgres.Database
			if jsonConfig.Postgres.Secure {
				config.PostgreSQL.SSLMode = "require"
			} else {
				config.PostgreSQL.SSLMode = "disable"
			}
		}

		// ClickHouse configuration
		config.ClickHouse.Host = jsonConfig.ClickHouse.Host
		config.ClickHouse.Port = jsonConfig.ClickHouse.Port
		config.ClickHouse.User = jsonConfig.ClickHouse.User
		config.ClickHouse.Password = jsonConfig.ClickHouse.Password
		config.ClickHouse.Database = jsonConfig.ClickHouse.Database

		// SMTP configuration
		config.SMTP.Host = jsonConfig.SMTP.Host
		config.SMTP.Port = jsonConfig.SMTP.Port
		config.SMTP.Username = jsonConfig.SMTP.Username
		config.SMTP.Password = jsonConfig.SMTP.Password
		config.SMTP.FromEmail = jsonConfig.SMTP.FromEmail
		config.SMTP.FromName = jsonConfig.SMTP.FromName

		return config
	}

	// Fallback to environment variables and .env file
	log.Println("📄 Loading configuration from environment variables")
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Server configuration
	config.Server.Port = getEnv("SERVER_PORT", "8080")
	config.Server.Host = getEnv("SERVER_HOST", "localhost")

	// PostgreSQL configuration
	config.PostgreSQL.Host = getEnv("POSTGRESQL_HOST", "localhost")
	config.PostgreSQL.Port = getEnv("POSTGRESQL_PORT", "5432")
	config.PostgreSQL.User = getEnv("POSTGRESQL_USER", "postgres")
	config.PostgreSQL.Password = getEnv("POSTGRESQL_PASSWORD", "")
	config.PostgreSQL.Database = getEnv("POSTGRESQL_DATABASE", "postgres")
	config.PostgreSQL.SSLMode = getEnv("POSTGRESQL_SSLMODE", "disable")

	// ClickHouse configuration
	config.ClickHouse.Host = getEnv("CLICKHOUSE_HOST", "localhost")
	config.ClickHouse.Port = getEnv("CLICKHOUSE_PORT", "9000")
	config.ClickHouse.User = getEnv("CLICKHOUSE_USER", "default")
	config.ClickHouse.Password = getEnv("CLICKHOUSE_PASSWORD", "")
	config.ClickHouse.Database = getEnv("CLICKHOUSE_DATABASE", "default")

	// SMTP configuration
	config.SMTP.Host = getEnv("SMTP_HOST", "smtp.brevo.com")
	config.SMTP.Port = getEnvAsInt("SMTP_PORT", 587)
	config.SMTP.Username = getEnv("SMTP_USERNAME", "")
	config.SMTP.Password = getEnv("SMTP_PASSWORD", "")
	config.SMTP.FromEmail = getEnv("SMTP_FROM_EMAIL", "noreply@bcaccount.com")
	config.SMTP.FromName = getEnv("SMTP_FROM_NAME", "SML Email Service")

	return config
}

// loadJSONConfig attempts to load configuration from smlgoapi.json
func loadJSONConfig() *JSONConfig {
	// Try multiple possible locations for the config file
	possiblePaths := []string{
		"smlgoapi.json",
		"./smlgoapi.json",
		filepath.Join(".", "smlgoapi.json"),
	}

	for _, configPath := range possiblePaths {
		data, err := os.ReadFile(configPath)
		if err != nil {
			continue // Try next path
		}

		var jsonConfig JSONConfig
		if err := json.Unmarshal(data, &jsonConfig); err != nil {
			log.Printf("Warning: Error parsing %s: %v", configPath, err)
			continue
		}

		log.Printf("✅ Successfully loaded configuration from %s", configPath)
		return &jsonConfig
	}

	log.Println("📄 smlgoapi.json not found, falling back to environment variables")
	return nil
}

func (c *Config) GetPostgreSQLDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.PostgreSQL.User,
		c.PostgreSQL.Password,
		c.PostgreSQL.Host,
		c.PostgreSQL.Port,
		c.PostgreSQL.Database,
		c.PostgreSQL.SSLMode,
	)
}

func (c *Config) GetClickHouseDSN() string {
	return fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s",
		c.ClickHouse.User,
		c.ClickHouse.Password,
		c.ClickHouse.Host,
		c.ClickHouse.Port,
		c.ClickHouse.Database,
	)
}

func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
