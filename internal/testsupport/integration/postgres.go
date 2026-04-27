//go:build integration

package integration

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	core_pgx_pool "github.com/horizoonn/todoapp/internal/core/repository/postgres/pool/pgx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	postgresImage    = "postgres:18.1-bookworm"
	postgresUser     = "todoapp-test-user"
	postgresPassword = "todoapp-test-password"
	postgresDatabase = "todoapp"
)

func NewPostgresPool(t *testing.T) *core_pgx_pool.Pool {
	t.Helper()

	ctx := context.Background()
	root := projectRoot(t)
	migrationPath := filepath.Join(root, "migrations", "000001_init.up.sql")

	container, err := postgres.Run(
		ctx,
		postgresImage,
		postgres.WithUsername(postgresUser),
		postgres.WithPassword(postgresPassword),
		postgres.WithDatabase(postgresDatabase),
		postgres.WithInitScripts(migrationPath),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Fatalf("terminate postgres container: %v", err)
		}
	})

	connectionString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get postgres connection string: %v", err)
	}

	cfg := postgresConfigFromURL(t, connectionString)
	pool, err := core_pgx_pool.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("create pgx pool: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool
}

func TruncateAll(t *testing.T, pool *core_pgx_pool.Pool) {
	t.Helper()

	ctx := context.Background()
	_, err := pool.Exec(ctx, "TRUNCATE TABLE todoapp.tasks, todoapp.users RESTART IDENTITY CASCADE;")
	if err != nil {
		t.Fatalf("truncate postgres tables: %v", err)
	}
}

func postgresConfigFromURL(t *testing.T, connectionString string) core_pgx_pool.Config {
	t.Helper()

	u, err := url.Parse(connectionString)
	if err != nil {
		t.Fatalf("parse postgres connection string: %v", err)
	}

	password, _ := u.User.Password()

	return core_pgx_pool.Config{
		Host:     u.Hostname(),
		Port:     u.Port(),
		User:     u.User.Username(),
		Password: password,
		Database: strings.TrimPrefix(u.Path, "/"),
		Timeout:  5 * time.Second,
	}
}

func projectRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("get runtime caller")
	}

	dir := filepath.Dir(filename)
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := filepath.Abs(goModPath); err == nil {
			if fileExists(goModPath) {
				return dir
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("project root not found")
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
