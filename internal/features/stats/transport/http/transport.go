package stats_transport_http

import (
	"context"
	"net/http"

	"github.com/horizoonn/todoapp/internal/core/domain"
	core_http_server "github.com/horizoonn/todoapp/internal/core/transport/http/server"
	stats_feature "github.com/horizoonn/todoapp/internal/features/stats"
)

type StatsHTTPHandler struct {
	statsService StatsService
}

type StatsService interface {
	GetStats(ctx context.Context, filter stats_feature.GetStatsFilter) (domain.Stats, error)
}

func NewStatsHTTPHandler(statsService StatsService) *StatsHTTPHandler {
	return &StatsHTTPHandler{
		statsService: statsService,
	}
}

func (h *StatsHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/stats",
			Handler: h.GetStats,
		},
	}
}
