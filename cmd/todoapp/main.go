package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_logger "github.com/horizoonn/todoapp/internal/core/logger"
	core_pgx_pool "github.com/horizoonn/todoapp/internal/core/repository/postgres/pool/pgx"
	http_middleware "github.com/horizoonn/todoapp/internal/core/transport/http/middleware"
	http_server "github.com/horizoonn/todoapp/internal/core/transport/http/server"
	tasks_postgres_repository "github.com/horizoonn/todoapp/internal/features/tasks/repository/postgres"
	tasks_service "github.com/horizoonn/todoapp/internal/features/tasks/service"
	tasks_transport_http "github.com/horizoonn/todoapp/internal/features/tasks/transport/http"
	users_postgres_repository "github.com/horizoonn/todoapp/internal/features/users/repository/postgres"
	users_service "github.com/horizoonn/todoapp/internal/features/users/service"
	users_transport_http "github.com/horizoonn/todoapp/internal/features/users/transport/http"
	"go.uber.org/zap"
)

var (
	timeZone = time.UTC
)

func main() {
	time.Local = timeZone

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Printf("failed to init application logger: %v", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("application time zone", zap.Any("zone", timeZone))

	logger.Debug("initializing postgres connection pool")
	pool, err := core_pgx_pool.NewPool(ctx, core_pgx_pool.NewConfigMust())
	if err != nil {
		logger.Fatal("failed to init postgres connection pool", zap.Error(err))
	}
	defer pool.Close()

	logger.Debug("initializing feature", zap.String("feature", "users"))
	usersRepository := users_postgres_repository.NewUsersRepository(pool)
	usersService := users_service.NewUsersService(usersRepository)
	usersTransportHTTP := users_transport_http.NewUsersHTTPHandler(usersService)

	logger.Debug("initializing feature", zap.String("feature", "tasks"))
	tasksRepository := tasks_postgres_repository.NewTasksRepository(pool)
	tasksService := tasks_service.NewTasksService(tasksRepository)
	tasksTransportHTTP := tasks_transport_http.NewTasksHTTPHandler(tasksService)

	logger.Debug("initializing HTTP server")
	httpServer := http_server.NewHTTPServer(
		http_server.NewConfigMust(),
		*logger,
		http_middleware.RequestID(),
		http_middleware.Logger(logger),
		http_middleware.Trace(),
		http_middleware.Panic(),
	)

	apiVersionRouter := http_server.NewAPIVersionRouter(http_server.ApiVersion1)
	apiVersionRouter.RegisterRoutes(usersTransportHTTP.Routes()...)
	apiVersionRouter.RegisterRoutes(tasksTransportHTTP.Routes()...)

	httpServer.RegisterAPIRouters(apiVersionRouter)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}

}
