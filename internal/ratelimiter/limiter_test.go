package ratelimiter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type call struct {
	key   string
	limit int
}

type fakeStore struct {
	calls   []call
	allowed bool
}

func (s *fakeStore) Allow(_ context.Context, key string, limit int, _, _ time.Duration) (bool, error) {
	s.calls = append(s.calls, call{key: key, limit: limit})
	return s.allowed, nil
}

func TestTokenTakesPrecedenceOverIP(t *testing.T) {
	store := &fakeStore{allowed: true}
	limiter := New(store, Config{IPRequestsPerSecond: 10, TokenRequestsPerSecond: 100, Window: time.Second, BlockDuration: time.Minute})
	allowed, err := limiter.Allow(context.Background(), "192.0.2.10", "client-token")
	if err != nil || !allowed {
		t.Fatalf("expected request to be allowed, err=%v", err)
	}
	if got := store.calls[0]; got.key != "token:client-token" || got.limit != 100 {
		t.Fatalf("token must override IP; got %+v", got)
	}
}

func TestTokenOverrideHasPriorityOverDefaultTokenLimit(t *testing.T) {
	store := &fakeStore{allowed: true}
	limiter := New(store, Config{IPRequestsPerSecond: 10, TokenRequestsPerSecond: 100, TokenOverrides: map[string]int{"vip": 500}, Window: time.Second, BlockDuration: time.Minute})
	_, _ = limiter.Allow(context.Background(), "192.0.2.10", "vip")
	if got := store.calls[0]; got.limit != 500 {
		t.Fatalf("expected override 500, got %d", got.limit)
	}
}

func TestMiddlewareReturnsExactMessageWhenBlocked(t *testing.T) {
	store := &fakeStore{allowed: false}
	limiter := New(store, Config{IPRequestsPerSecond: 1, TokenRequestsPerSecond: 1, Window: time.Second, BlockDuration: time.Minute})
	handler := Middleware(limiter)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.3:1234"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", response.Code)
	}
	if response.Body.String() != RejectedMessage {
		t.Fatalf("unexpected response body: %q", response.Body.String())
	}
}
