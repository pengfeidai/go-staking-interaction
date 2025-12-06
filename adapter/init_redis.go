package adapter

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"staking-interaction/common/config"
	"time"
)

func NewRedisClientWithRetry() (*redis.Client, error) {
	var client *redis.Client
	var err error
	conf := config.Get()
	redisConfig := conf.RedisConfig
	maxRetries := conf.Sync.RetryAttempts

	for i := 0; i <= maxRetries; i++ {
		client, err = NewRedisClient(redisConfig)
		if err == nil {
			return client, nil
		}
		if i < maxRetries {
			waitTime := time.Duration(i) * 2 * time.Second
			fmt.Printf("Retrying in %v...\n", waitTime)
			time.Sleep(waitTime)
		}
	}
	return nil, fmt.Errorf("redis connection failed: %v, tried %d times", err, maxRetries)
}

func NewRedisClient(redisConfig config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password:     redisConfig.Password,
		DB:           redisConfig.DatabaseConfig,
		PoolSize:     redisConfig.PoolSize,
		MinIdleConns: redisConfig.MinIdleConns,
		DialTimeout:  redisConfig.DialTimeout,
		ReadTimeout:  redisConfig.ReadTimeout,
		WriteTimeout: redisConfig.WriteTimeout,
		IdleTimeout:  redisConfig.IdleTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %v", err)
	}

	fmt.Println("Redis connection success.")
	return client, nil
}
