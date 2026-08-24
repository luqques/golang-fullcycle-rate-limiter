package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lucas/go-rate-limiter/internal/ratelimiter"
)

type Config struct {
	Port          string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RateLimit     ratelimiter.Config
}

func Load() (Config, error) {
	ipLimit, err := positiveInt("RATE_LIMIT_IP_REQUESTS_PER_SECOND", 10)
	if err != nil {
		return Config{}, err
	}
	tokenLimit, err := positiveInt("RATE_LIMIT_TOKEN_REQUESTS_PER_SECOND", 100)
	if err != nil {
		return Config{}, err
	}
	block, err := duration("RATE_LIMIT_BLOCK_DURATION", "5m")
	if err != nil {
		return Config{}, err
	}
	window, err := duration("RATE_LIMIT_WINDOW", "1s")
	if err != nil {
		return Config{}, err
	}
	db, err := intValue("REDIS_DB", 0)
	if err != nil {
		return Config{}, err
	}
	overrides, err := parseTokenOverrides(os.Getenv("RATE_LIMIT_TOKEN_OVERRIDES"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		Port:          env("PORT", "8080"),
		RedisAddr:     env("REDIS_ADDR", "localhost:6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       db,
		RateLimit: ratelimiter.Config{IPRequestsPerSecond: ipLimit, TokenRequestsPerSecond: tokenLimit, TokenOverrides: overrides, BlockDuration: block, Window: window},
	}, nil
}

func parseTokenOverrides(value string) (map[string]int, error) {
	result := map[string]int{}
	if strings.TrimSpace(value) == "" {
		return result, nil
	}
	for _, item := range strings.Split(value, ",") {
		pair := strings.SplitN(strings.TrimSpace(item), ":", 2)
		if len(pair) != 2 || pair[0] == "" {
			return nil, fmt.Errorf("invalid RATE_LIMIT_TOKEN_OVERRIDES entry %q; expected token:limit", item)
		}
		limit, err := strconv.Atoi(pair[1])
		if err != nil || limit < 1 {
			return nil, fmt.Errorf("invalid token limit in %q", item)
		}
		result[pair[0]] = limit
	}
	return result, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func positiveInt(key string, fallback int) (int, error) {
	value, err := intValue(key, fallback)
	if err != nil || value < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return value, nil
}

func intValue(key string, fallback int) (int, error) {
	value := env(key, strconv.Itoa(fallback))
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return parsed, nil
}

func duration(key, fallback string) (time.Duration, error) {
	value := env(key, fallback)
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}
	return parsed, nil
}
