package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_config "github.com/horizoonn/todoapp/internal/core/config"
	core_logger "github.com/horizoonn/todoapp/internal/core/logger"
	core_pgx_pool "github.com/horizoonn/todoapp/internal/core/repository/postgres/pool/pgx"
	http_middleware "github.com/horizoonn/todoapp/internal/core/transport/http/middleware"
	http_server "github.com/horizoonn/todoapp/internal/core/transport/http/server"
	stats_postgres_repository "github.com/horizoonn/todoapp/internal/features/stats/repository/postgres"
	stats_service "github.com/horizoonn/todoapp/internal/features/stats/service"
	stats_transport_http "github.com/horizoonn/todoapp/internal/features/stats/transport/http"
	tasks_postgres_repository "github.com/horizoonn/todoapp/internal/features/tasks/repository/postgres"
	tasks_service "github.com/horizoonn/todoapp/internal/features/tasks/service"
	tasks_transport_http "github.com/horizoonn/todoapp/internal/features/tasks/transport/http"
	users_postgres_repository "github.com/horizoonn/todoapp/internal/features/users/repository/postgres"
	users_service "github.com/horizoonn/todoapp/internal/features/users/service"
	users_transport_http "github.com/horizoonn/todoapp/internal/features/users/transport/http"
	"go.uber.org/zap"
)

func main() {
	cfg := core_config.NewConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Printf("failed to init application logger: %v", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("application time zone", zap.Any("zone", time.Local))

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

	logger.Debug("initializing feature", zap.String("feature", "stats"))
	statsRepository := stats_postgres_repository.NewStatsRepository(pool)
	statsService := stats_service.NewStatsService(statsRepository)
	statsTransportHTTP := stats_transport_http.NewStatsHTTPHandler(statsService)

	logger.Debug("initializing HTTP server")
	httpConfig := http_server.NewConfigMust()
	httpServer := http_server.NewHTTPServer(
		httpConfig,
		*logger,
		http_middleware.RequestID(),
		http_middleware.Logger(logger),
		http_middleware.Trace(),
		http_middleware.Panic(),
		http_middleware.CORS(httpConfig.AllowedOrigins),
	)

	apiVersionRouterV1 := http_server.NewAPIVersionRouter(http_server.ApiVersion1)
	apiVersionRouterV1.AddRoutes(usersTransportHTTP.Routes()...)
	apiVersionRouterV1.AddRoutes(tasksTransportHTTP.Routes()...)
	apiVersionRouterV1.AddRoutes(statsTransportHTTP.Routes()...)

	httpServer.RegisterAPIRouters(apiVersionRouterV1)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}

}
