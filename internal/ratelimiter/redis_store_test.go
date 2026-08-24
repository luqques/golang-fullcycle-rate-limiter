package ratelimiter

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRedisStoreBlocksAfterLimit(t *testing.T) {
	address := os.Getenv("REDIS_ADDR")
	if address == "" {
		t.Skip("requires REDIS_ADDR (provided by docker compose test)")
	}
	store := NewRedisStore(address, "", 0)
	defer store.Close()
	ctx := context.Background()
	key := "test:" + time.Now().Format("20060102150405.000000000")
	for i := 0; i < 2; i++ {
		allowed, err := store.Allow(ctx, key, 2, time.Second, time.Minute)
		if err != nil || !allowed {
			t.Fatalf("request %d should be allowed: %v", i+1, err)
		}
	}
	allowed, err := store.Allow(ctx, key, 2, time.Second, time.Minute)
	if err != nil || allowed {
		t.Fatalf("third request should be blocked; allowed=%v err=%v", allowed, err)
	}
	allowed, err = store.Allow(ctx, key, 2, time.Second, time.Minute)
	if err != nil || allowed {
		t.Fatalf("blocked key must remain rejected; allowed=%v err=%v", allowed, err)
	}
}
