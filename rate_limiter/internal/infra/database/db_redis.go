package database

import (
	"sync"
	"time"

	"github.com/go-redis/redis"
)

type DBRedis struct {
	client *redis.Client
	mu     sync.Mutex
}

// NewDBRedis returns a new DBRedis with the given Redis client.
// It is used to create a new DBInterface for Redis.
func NewDBRedis(client *redis.Client) DBInterface {
	return &DBRedis{client: client}
}

// Get gets the value for the given key. It returns the value and an error.
func (r *DBRedis) Get(key string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.client.Get(key).Result()
}

// Set sets the value for the given key. It returns an error if something goes wrong.
func (r *DBRedis) Set(key string, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.client.Set(key, value, time.Second*10).Err()
}
