package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/lucas/go-rate-limiter/internal/config"
	"github.com/lucas/go-rate-limiter/internal/ratelimiter"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	store := ratelimiter.NewRedisStore(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	defer store.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := store.Ping(ctx); err != nil {
		log.Fatalf("redis is unavailable: %v", err)
	}

	limiter := ratelimiter.New(store, cfg.RateLimit)
	if err := limiter.Validate(); err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"request accepted"}`))
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           ratelimiter.Middleware(limiter)(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("listening on :%s", cfg.Port)
	log.Fatal(server.ListenAndServe())
}
