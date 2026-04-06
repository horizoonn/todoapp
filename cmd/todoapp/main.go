package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	core_logger "github.com/horizoonn/todoapp/internal/core/logger"
	postgres_pool "github.com/horizoonn/todoapp/internal/core/repository/postgres/pool"
	http_middleware "github.com/horizoonn/todoapp/internal/core/transport/http/middleware"
	http_server "github.com/horizoonn/todoapp/internal/core/transport/http/server"
	users_postgres_repository "github.com/horizoonn/todoapp/internal/features/users/repository/postgres"
	users_service "github.com/horizoonn/todoapp/internal/features/users/service"
	users_transport_http "github.com/horizoonn/todoapp/internal/features/users/transport/http"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init application logger: %w", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("initializing postgres connecion pool")
	pool, err := postgres_pool.NewConnectionPool(ctx, postgres_pool.NewConfigMust())
	if err != nil {
		logger.Fatal("failed to init postgres connection pool", zap.Error(err))
	}
	defer pool.Close()

	logger.Debug("initializing feature", zap.String("feature", "users"))
	usersRepository := users_postgres_repository.NewUsersRepository(pool)
	usersService := users_service.NewUsersService(usersRepository)

	usersTransportHTTP := users_transport_http.NewUsersHTTPHandler(usersService)

	logger.Debug("initializing HTTP server")
	httpServer := http_server.NewHTTPServer(
		http_server.NewConfigMust(),
		*logger,
		http_middleware.RequestID(),
		http_middleware.Logger(logger),
		http_middleware.Panic(),
		http_middleware.Trace(),
	)

	apiVersionRouter := http_server.NewAPIVersionRouter(http_server.ApiVersion1)
	apiVersionRouter.RegisterRoutes(usersTransportHTTP.Routes()...)
	httpServer.RegisterAPIRouters(apiVersionRouter)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}

}
