package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"crud-api/internal/domain/user"
	"crud-api/internal/ports"
)

type userCache struct {
	client *redis.Client
}

func NewUserCache(client *redis.Client) ports.UserCache {
	return &userCache{client: client}
}

func (c *userCache) Set(ctx context.Context, u *user.User, ttl time.Duration) error {
	b, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("marshal user: %w", err)
	}

	return c.client.Set(ctx, c.key(u.ID), string(b), ttl).Err()
}

func (c *userCache) Get(ctx context.Context, id string) (*user.User, error) {
	val, err := c.client.Get(ctx, c.key(id)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var u user.User
	if err := json.Unmarshal([]byte(val), &u); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}

	return &u, nil
}

func (c *userCache) Delete(ctx context.Context, id string) error {
	return c.client.Del(ctx, c.key(id)).Err()
}

func (c *userCache) key(id string) string {
	return fmt.Sprintf("user:{%s}", id)
}
