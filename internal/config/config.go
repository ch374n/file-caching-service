package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// RedisMode defines how Redis is configured
type RedisMode string

const (
	RedisModeDisabled RedisMode = "disabled" // No caching
	RedisModeEnabled  RedisMode = "enabled"  // Redis caching enabled
)

type Config struct {
	Port     string
	LogLevel string
	Redis    RedisConfig
	R2       R2Config
}

type RedisConfig struct {
	Mode     RedisMode
	Addr     string
	Password string
	DB       int
	CacheTTL time.Duration

	// Timeout settings (optimized for in-cluster Redis)
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
}

func Load() *Config {
	redisMode := parseRedisMode(getEnv("REDIS_MODE", "enabled"))

	return &Config{
		Port:     getEnv("PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		Redis: RedisConfig{
			Mode:         redisMode,
			Addr:         getEnv("REDIS_ADDR", "localhost:6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvAsInt("REDIS_DB", 0),
			CacheTTL:     getEnvAsDuration("CACHE_TTL", 5*time.Minute),
			DialTimeout:  getEnvAsDuration("REDIS_DIAL_TIMEOUT", 2*time.Second),
			ReadTimeout:  getEnvAsDuration("REDIS_READ_TIMEOUT", 5*time.Second),
			WriteTimeout: getEnvAsDuration("REDIS_WRITE_TIMEOUT", 5*time.Second),
		},
		R2: R2Config{
			AccountID:       getEnv("R2_ACCOUNT_ID", ""),
			AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
			BucketName:      getEnv("R2_BUCKET_NAME", ""),
		},
	}
}

func parseRedisMode(mode string) RedisMode {
	switch strings.ToLower(mode) {
	case "disabled", "none", "off", "false":
		return RedisModeDisabled
	default:
		return RedisModeEnabled
	}
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

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
