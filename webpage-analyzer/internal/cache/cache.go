package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"webpage-analyzer/internal/models"
)

// ResultCache stores analysis results.
type ResultCache interface {
	Get(ctx context.Context, url string) (*models.AnalysisResult, bool, error)
	Set(ctx context.Context, url string, result *models.AnalysisResult, ttl time.Duration) error
}

type RedisResultCache struct {
	client *redis.Client
	prefix string
}

func NewRedisResultCache(addr, password string, db int) *RedisResultCache {
	return &RedisResultCache{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
		prefix: "analysis",
	}
}

func (c *RedisResultCache) Get(ctx context.Context, url string) (*models.AnalysisResult, bool, error) {
	value, err := c.client.Get(ctx, c.key(url)).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var result models.AnalysisResult
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, false, err
	}

	return &result, true, nil
}

func (c *RedisResultCache) Set(ctx context.Context, url string, result *models.AnalysisResult, ttl time.Duration) error {
	payload, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, c.key(url), payload, ttl).Err()
}

func (c *RedisResultCache) key(url string) string {
	hash := sha256.Sum256([]byte(url))
	return fmt.Sprintf("%s:%s", c.prefix, hex.EncodeToString(hash[:]))
}
