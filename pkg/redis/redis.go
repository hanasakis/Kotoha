package redis

import (
	"context"
	"time"

	klog "github.com/hanasakis/kotoha/pkg/log"
	goredis "github.com/redis/go-redis/v9"
)

type Client struct {
	RDB *goredis.Client
}

func New(addr, password string, db int) (*Client, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	klog.Info("Redis connected", nil)
	return &Client{RDB: rdb}, nil
}

func (c *Client) Close() error {
	return c.RDB.Close()
}
