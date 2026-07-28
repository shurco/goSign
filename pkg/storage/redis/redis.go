package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var Conn Handler

// Config is ...
type Config struct {
	Address  string `toml:"address"`
	Password string `toml:"password"`
}

// Handler is ...
type Handler interface {
	Close() error
	Ping() error
	Set(key string, value any, expiration time.Duration) error
	Get(key string) *redis.StringCmd
	Delete(key string) (int64, error)

	Client() *redis.Client
}

// rdb is a Redis-backed implementation of the Handler interface.
type rdb struct {
	ctx    context.Context
	client *redis.Client
}

// New creates a new Handler backed by Redis using the given options.
func New(ctx context.Context, addr, password string) Handler {
	Conn = &rdb{
		ctx: ctx,
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			Protocol: 2, // Use RESP2 protocol for compatibility with older Redis versions
		}),
	}
	return Conn
}

// Close closes the underlying Redis client's connections.
func (r *rdb) Close() error {
	return r.client.Close()
}

// Ping provides a way to ping a Redis server.
func (r *rdb) Ping() error {
	_, err := r.client.Ping(r.ctx).Result()
	return err
}

// Set is provides a way to set values in Redis.
func (r *rdb) Set(key string, value any, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

// Get is provides a way to retrieve values from Redis.
func (r *rdb) Get(key string) *redis.StringCmd {
	return r.client.Get(r.ctx, key)
}

// Delete is provides a way to delete values from Redis.
func (r *rdb) Delete(key string) (int64, error) {
	return r.client.Del(r.ctx, key).Result()
}

// Client is provides a way to obtain a Redis client.
func (r *rdb) Client() *redis.Client {
	return r.client
}
