package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	OpenAIAPIKey    string
	AssistantModel  string
	ModeratorModel  string
	ServerPort      int
	LogLevel        string
	MaxTokens       int
	Temperature     float64
	RateLimitPerMin int
	RequestTimeout  time.Duration
	EnableMetrics   bool
	CacheEnabled    bool
	CacheTTL        time.Duration
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		OpenAIAPIKey:    getEnv("OPENAI_API_KEY", ""),
		AssistantModel:  getEnv("ASSISTANT_MODEL", "gpt-3.5-turbo"),
		ModeratorModel:  getEnv("MODERATOR_MODEL", "gpt-4"),
		ServerPort:      getEnvAsInt("SERVER_PORT", 8080),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		MaxTokens:       getEnvAsInt("MAX_TOKENS", 500),
		Temperature:     getEnvAsFloat("TEMPERATURE", 0.7),
		RateLimitPerMin: getEnvAsInt("RATE_LIMIT_PER_MIN", 60),
		RequestTimeout:  getEnvAsDuration("REQUEST_TIMEOUT", 30*time.Second),
		EnableMetrics:   getEnvAsBool("ENABLE_METRICS", true),
		CacheEnabled:    getEnvAsBool("CACHE_ENABLED", true),
		CacheTTL:        getEnvAsDuration("CACHE_TTL", 1*time.Hour),
	}

	if cfg.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}