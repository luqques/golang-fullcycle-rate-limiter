package ratelimiter

import (
	"context"
	"fmt"
	"time"
)

const RejectedMessage = "you have reached the maximum number of requests or actions allowed within a certain time frame"

// Store is the persistence strategy. Implement another Store to replace Redis without changing business or HTTP code.
type Store interface {
	Allow(ctx context.Context, key string, limit int, window, blockDuration time.Duration) (bool, error)
}

type Config struct {
	IPRequestsPerSecond    int
	TokenRequestsPerSecond int
	TokenOverrides         map[string]int
	Window                 time.Duration
	BlockDuration          time.Duration
}

type Limiter struct {
	store  Store
	config Config
}

func New(store Store, config Config) *Limiter {
	return &Limiter{store: store, config: config}
}

func (l *Limiter) Allow(ctx context.Context, ip, token string) (bool, error) {
	key, limit := l.subject(ip, token)
	return l.store.Allow(ctx, key, limit, l.config.Window, l.config.BlockDuration)
}

func (l *Limiter) subject(ip, token string) (string, int) {
	if token != "" {
		if limit, found := l.config.TokenOverrides[token]; found {
			return "token:" + token, limit
		}
		return "token:" + token, l.config.TokenRequestsPerSecond
	}
	return "ip:" + ip, l.config.IPRequestsPerSecond
}

func (l *Limiter) Validate() error {
	if l.store == nil || l.config.IPRequestsPerSecond < 1 || l.config.TokenRequestsPerSecond < 1 || l.config.Window <= 0 || l.config.BlockDuration <= 0 {
		return fmt.Errorf("invalid rate limiter configuration")
	}
	return nil
}
