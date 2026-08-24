package ratelimiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisKeyPrefix = "rate-limiter:"

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(address, password string, db int) *RedisStore {
	return &RedisStore{client: redis.NewClient(&redis.Options{Addr: address, Password: password, DB: db})}
}

func (s *RedisStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}

var allowScript = redis.NewScript(`
local blocked = KEYS[1] .. ':blocked'
if redis.call('EXISTS', blocked) == 1 then return 0 end
local count = redis.call('INCR', KEYS[1])
if count == 1 then redis.call('PEXPIRE', KEYS[1], ARGV[1]) end
if count > tonumber(ARGV[2]) then
  redis.call('SET', blocked, '1', 'PX', ARGV[3])
  return 0
end
return 1`)

func (s *RedisStore) Allow(ctx context.Context, key string, limit int, window, blockDuration time.Duration) (bool, error) {
	result, err := allowScript.Run(ctx, s.client, []string{redisKeyPrefix + key}, window.Milliseconds(), limit, blockDuration.Milliseconds()).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}
