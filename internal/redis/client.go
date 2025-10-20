package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
}

func NewClient(addr, password string, db int) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &Client{rdb}
}

func (c *Client) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	err := c.Get(ctx, jti).Err()
	if err == redis.Nil {
		return false, nil // Non in blacklist
	}
	if err != nil {
		return false, err // Errore di Redis
	}
	return true, nil // È in blacklist
}

func (c *Client) BlacklistToken(ctx context.Context, jti string, expiration time.Duration) error {
	return c.Set(ctx, jti, "revoked", expiration).Err()
}