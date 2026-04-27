//go:build integration

package integration

import (
	"context"
	"net/url"
	"testing"
	"time"

	core_goredis_pool "github.com/horizoonn/todoapp/internal/core/repository/redis/pool/goredis"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

const redisImage = "redis:8.6.1"

func NewRedisPool(t *testing.T) *core_goredis_pool.Pool {
	t.Helper()

	ctx := context.Background()
	container, err := tcredis.Run(ctx, redisImage)
	if err != nil {
		t.Fatalf("start redis container: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Fatalf("terminate redis container: %v", err)
		}
	})

	connectionString, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("get redis connection string: %v", err)
	}

	cfg := redisConfigFromURL(t, connectionString)
	pool, err := core_goredis_pool.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("create redis pool: %v", err)
	}
	t.Cleanup(func() {
		if err := pool.Close(); err != nil {
			t.Fatalf("close redis pool: %v", err)
		}
	})

	return pool
}

func FlushRedis(t *testing.T, pool *core_goredis_pool.Pool) {
	t.Helper()

	ctx := context.Background()
	if err := pool.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("flush redis db: %v", err)
	}
}

func redisConfigFromURL(t *testing.T, connectionString string) core_goredis_pool.Config {
	t.Helper()

	u, err := url.Parse(connectionString)
	if err != nil {
		t.Fatalf("parse redis connection string: %v", err)
	}

	password, _ := u.User.Password()

	return core_goredis_pool.Config{
		Host:     u.Hostname(),
		Port:     u.Port(),
		Password: password,
		DB:       0,
		TTL:      time.Minute,
	}
}
