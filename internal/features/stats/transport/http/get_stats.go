package stats_transport_http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	core_logger "github.com/horizoonn/todoapp/internal/core/logger"
	core_http_request "github.com/horizoonn/todoapp/internal/core/transport/http/request"
	core_http_response "github.com/horizoonn/todoapp/internal/core/transport/http/response"
	stats_feature "github.com/horizoonn/todoapp/internal/features/stats"
)

type GetStatsResponse struct {
	TasksCreated               int      `json:"tasks_created"`
	TasksCompleted             int      `json:"tasks_completed"`
	TasksCompletedRate         *float64 `json:"tasks_completed_rate"`
	TasksAverageCompletionTime *string  `json:"tasks_average_completion_time"`
}

func (h *StatsHTTPHandler) GetStats(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userID, from, to, err := getUserIDFromToQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get userID/from/to query params")

		return
	}

	filter := stats_feature.NewGetStatsFilter(userID, from, to)

	statsDomain, err := h.statsService.GetStats(ctx, filter)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get stats")

		return
	}

	response := GetStatsResponse(toDTOFromDomain(statsDomain))

	responseHandler.JSONResponse(response, http.StatusOK)
}

func toDTOFromDomain(stats domain.Stats) GetStatsResponse {
	response := GetStatsResponse{
		TasksCreated:       stats.TasksCreated,
		TasksCompleted:     stats.TasksCompleted,
		TasksCompletedRate: stats.TasksCompletedRate,
	}

	if stats.TasksAverageCompletionTime != nil {
		durationStr := stats.TasksAverageCompletionTime.String()
		response.TasksAverageCompletionTime = &durationStr
	}

	return response
}

func getUserIDFromToQueryParams(r *http.Request) (*uuid.UUID, *time.Time, *time.Time, error) {
	const (
		userIDQueryParamKey = "user_id"
		fromQueryParamKey   = "from"
		toQueryParamKey     = "to"
	)

	userID, err := core_http_request.GetUUIDQueryParam(r, userIDQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'user_id' query param: %w", err)
	}

	from, err := core_http_request.GetDateQueryParam(r, fromQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'from' query param: %w", err)
	}

	to, err := core_http_request.GetDateQueryParam(r, toQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'to' query param: %w", err)
	}

	return userID, from, to, nil
}
