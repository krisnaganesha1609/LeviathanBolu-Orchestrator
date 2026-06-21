package database

import (
	"context"
	"log"
	"time"

	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/configs"
	"github.com/redis/go-redis/v9"
)

// ConnectRedis opens a connection to Redis and verifies it with a PING
// before returning. Redis is used for ephemeral assistant state: active
// device sessions, conversation/session cache, wake-word debounce, etc.
// (anything that doesn't need PostgreSQL's durability).
func ConnectRedis(cfg *configs.OrchestratorConfig) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,

		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,

		PoolSize:     10,
		MinIdleConns: 2,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("[database] failed to connect to Redis: %v", err)
	}

	log.Println("[database] connected to Redis")
	return rdb
}
