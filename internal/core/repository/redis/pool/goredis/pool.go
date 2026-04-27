package core_goredis_pool

import (
	"context"
	"fmt"
	"time"

	core_redis_pool "github.com/horizoonn/todoapp/internal/core/repository/redis/pool"
	"github.com/redis/go-redis/v9"
)

type Pool struct {
	client *redis.Client
	ttl    time.Duration
}

func NewPool(ctx context.Context, cfg Config) (*Pool, error) {
	options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	client := redis.NewClient(options)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Pool{
		client: client,
		ttl:    cfg.TTL,
	}, nil
}

func (p *Pool) Get(
	ctx context.Context,
	key string,
) core_redis_pool.StringCmd {
	cmd := p.client.Get(ctx, key)

	return goredisStringCmd{cmd}
}

func (p *Pool) Set(
	ctx context.Context,
	key string,
	value any,
	ttl time.Duration,
) core_redis_pool.StatusCmd {
	cmd := p.client.Set(ctx, key, value, ttl)

	return goredisStatusCmd{cmd}
}

func (p *Pool) Del(
	ctx context.Context,
	keys ...string,
) core_redis_pool.IntCmd {
	cmd := p.client.Del(ctx, keys...)

	return goredisIntCmd{cmd}
}

func (p *Pool) Incr(
	ctx context.Context,
	key string,
) core_redis_pool.IntCmd {
	cmd := p.client.Incr(ctx, key)

	return goredisIntCmd{cmd}
}

func (p *Pool) HGet(
	ctx context.Context,
	key string,
	field string,
) core_redis_pool.StringCmd {
	cmd := p.client.HGet(ctx, key, field)

	return goredisStringCmd{cmd}
}

func (p *Pool) HSet(
	ctx context.Context,
	key string,
	values ...any,
) core_redis_pool.IntCmd {
	cmd := p.client.HSet(ctx, key, values...)

	return goredisIntCmd{cmd}
}

func (p *Pool) FlushDB(ctx context.Context) core_redis_pool.StatusCmd {
	cmd := p.client.FlushDB(ctx)

	return goredisStatusCmd{cmd}
}

func (p *Pool) Close() error {
	return p.client.Close()
}

func (p *Pool) TTL() time.Duration {
	return p.ttl
}
