package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server struct {
		Port string
		Host string
	}
	ClickHouse struct {
		Host     string
		Port     string
		User     string
		Password string
		Database string
		Secure   bool
	}
}

func LoadConfig() *Config {
	// โหลด .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	config := &Config{}

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

	return config
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

func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
