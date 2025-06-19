package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	Server struct {
		Port string `json:"port"`
		Host string `json:"host"`
	} `json:"server"`
	ClickHouse struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
		Secure   bool   `json:"secure"`
	} `json:"clickhouse"`
	PostgreSQL struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
		SSLMode  string `json:"sslmode"`
	} `json:"postgresql"`
}

// JSONConfig represents the structure of smlgoapi.json
type JSONConfig struct {
	Server struct {
		Host string `json:"host"`
		Port string `json:"port"`
	} `json:"server"`
	ClickHouse struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
		Secure   bool   `json:"secure"`
	} `json:"clickhouse"`
	PostgreSQL struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
		SSLMode  string `json:"sslmode"`
	} `json:"postgresql"`
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
		config.ClickHouse.Host = jsonConfig.ClickHouse.Host
		config.ClickHouse.Port = jsonConfig.ClickHouse.Port
		config.ClickHouse.User = jsonConfig.ClickHouse.User
		config.ClickHouse.Password = jsonConfig.ClickHouse.Password
		config.ClickHouse.Database = jsonConfig.ClickHouse.Database
		config.ClickHouse.Secure = jsonConfig.ClickHouse.Secure

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
	// ClickHouse configuration
	config.ClickHouse.Host = getEnv("CLICKHOUSE_HOST", "localhost")
	config.ClickHouse.Port = getEnv("CLICKHOUSE_PORT", "9000")
	config.ClickHouse.User = getEnv("CLICKHOUSE_USER", "default")
	config.ClickHouse.Password = getEnv("CLICKHOUSE_PASSWORD", "")
	config.ClickHouse.Database = getEnv("CLICKHOUSE_DATABASE", "default")
	config.ClickHouse.Secure = getEnv("CLICKHOUSE_SECURE", "false") == "true"

	// PostgreSQL configuration
	config.PostgreSQL.Host = getEnv("POSTGRESQL_HOST", "localhost")
	config.PostgreSQL.Port = getEnv("POSTGRESQL_PORT", "5432")
	config.PostgreSQL.User = getEnv("POSTGRESQL_USER", "postgres")
	config.PostgreSQL.Password = getEnv("POSTGRESQL_PASSWORD", "")
	config.PostgreSQL.Database = getEnv("POSTGRESQL_DATABASE", "postgres")
	config.PostgreSQL.SSLMode = getEnv("POSTGRESQL_SSLMODE", "disable")

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
		data, err := ioutil.ReadFile(configPath)
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

func (c *Config) GetClickHouseDSN() string {
	return fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s?secure=%t",
		c.ClickHouse.User,
		c.ClickHouse.Password,
		c.ClickHouse.Host,
		c.ClickHouse.Port,
		c.ClickHouse.Database,
		c.ClickHouse.Secure,
	)
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

func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
