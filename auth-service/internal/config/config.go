package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppPort         string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	DBMaxConns      int32
	DBMinConns      int32
	DBMaxConnIdle   time.Duration
	DBMaxConnLife   time.Duration
	RedisHost       string
	RedisPort       string
	RedisPassword   string
	RSAPrivateKeyPath string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		AppPort:           getEnv("APP_PORT", "8080"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        os.Getenv("DB_PASSWORD"),
		DBName:            getEnv("DB_NAME", "auth_db"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		DBMaxConns:        getEnvAsInt32("DB_MAX_CONNS", 25),
		DBMinConns:        getEnvAsInt32("DB_MIN_CONNS", 5),
		DBMaxConnIdle:     getEnvAsDuration("DB_MAX_CONN_IDLE", 15*time.Minute),
		DBMaxConnLife:     getEnvAsDuration("DB_MAX_CONN_LIFE", 1*time.Hour),
		RedisHost:         getEnv("REDIS_HOST", "localhost"),
		RedisPort:         getEnv("REDIS_PORT", "6379"),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		RSAPrivateKeyPath: os.Getenv("RSA_PRIVATE_KEY_PATH"),
		AccessTokenTTL:    15 * time.Minute,
		RefreshTokenTTL:   7 * 24 * time.Hour,
	}

	if cfg.DBPassword == "" {
		return nil, errors.New("DB_PASSWORD environment variable is required")
	}

	return cfg, nil
}

func (c *Config) PostgresDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt32(key string, defaultVal int32) int32 {
	valStr := os.Getenv(key)
	if val, err := strconv.ParseInt(valStr, 10, 32); err == nil {
		return int32(val)
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valStr := os.Getenv(key)
	if val, err := time.ParseDuration(valStr); err == nil {
		return val
	}
	return defaultVal
}