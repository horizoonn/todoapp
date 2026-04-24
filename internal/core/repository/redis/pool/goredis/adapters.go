package core_goredis_pool

import (
	"errors"

	core_redis_pool "github.com/horizoonn/todoapp/internal/core/repository/redis/pool"
	"github.com/redis/go-redis/v9"
)

type goredisStringCmd struct {
	*redis.StringCmd
}

func (c goredisStringCmd) Bytes() ([]byte, error) {
	data, err := c.StringCmd.Bytes()
	if err != nil {
		return nil, mapError(err)
	}

	return data, nil
}

type goredisStatusCmd struct {
	*redis.StatusCmd
}

func (c goredisStatusCmd) Err() error {
	return mapError(c.StatusCmd.Err())
}

type goredisIntCmd struct {
	*redis.IntCmd
}

func (c goredisIntCmd) Err() error {
	return mapError(c.IntCmd.Err())
}

func (c goredisIntCmd) Val() int64 {
	return c.IntCmd.Val()
}

func mapError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, redis.Nil) {
		return core_redis_pool.NotFound
	}

	return err
}
