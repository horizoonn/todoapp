package stats_transport_http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
	core_http_response "github.com/horizoonn/todoapp/internal/core/transport/http/response"
	stats_feature "github.com/horizoonn/todoapp/internal/features/stats"
	stats_transport_http "github.com/horizoonn/todoapp/internal/features/stats/transport/http"
	stats_transport_http_mocks "github.com/horizoonn/todoapp/internal/features/stats/transport/http/mocks"
	"go.uber.org/mock/gomock"
)

func TestGetStatsParsesQueryAndReturnsStats(t *testing.T) {
	userID := uuid.New()
	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	completionRate := 50.0
	averageCompletionTime := 90 * time.Minute
	stats := domain.NewStats(10, 5, &completionRate, &averageCompletionTime)
	filter := stats_feature.NewGetStatsFilter(&userID, &from, &to)
	ctrl := gomock.NewController(t)

	service := stats_transport_http_mocks.NewMockStatsService(ctrl)
	service.EXPECT().
		GetStats(gomock.Any(), filter).
		Return(stats, nil)

	handler := stats_transport_http.NewStatsHTTPHandler(service)
	request := httptest.NewRequest(
		http.MethodGet,
		"/stats?user_id="+userID.String()+"&from=2026-04-01&to=2026-04-25",
		nil,
	)
	recorder := httptest.NewRecorder()

	handler.GetStats(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	var response stats_transport_http.GetStatsResponse
	decodeJSONResponse(t, recorder, &response)
	if response.TasksCreated != stats.TasksCreated || response.TasksCompleted != stats.TasksCompleted {
		t.Fatalf("expected stats %+v, got %+v", stats, response)
	}
	if response.TasksCompletedRate == nil || *response.TasksCompletedRate != completionRate {
		t.Fatalf("expected completion rate %v, got %v", completionRate, response.TasksCompletedRate)
	}
	if response.TasksAverageCompletionTime == nil || *response.TasksAverageCompletionTime != averageCompletionTime.String() {
		t.Fatalf("expected average completion time %q, got %v", averageCompletionTime.String(), response.TasksAverageCompletionTime)
	}
}

func TestGetStatsReturnsBadRequestForInvalidDateQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := stats_transport_http_mocks.NewMockStatsService(ctrl)
	handler := stats_transport_http.NewStatsHTTPHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/stats?from=25-04-2026", nil)
	recorder := httptest.NewRecorder()

	handler.GetStats(recorder, request)

	assertErrorResponse(t, recorder, http.StatusBadRequest, "invalid_argument")
}

func TestGetStatsReturnsBadRequestForServiceValidationError(t *testing.T) {
	from := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	filter := stats_feature.NewGetStatsFilter(nil, &from, &to)
	ctrl := gomock.NewController(t)

	service := stats_transport_http_mocks.NewMockStatsService(ctrl)
	service.EXPECT().
		GetStats(gomock.Any(), filter).
		Return(domain.Stats{}, core_errors.ErrInvalidArgument)

	handler := stats_transport_http.NewStatsHTTPHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/stats?from=2026-04-25&to=2026-04-01", nil)
	recorder := httptest.NewRecorder()

	handler.GetStats(recorder, request)

	assertErrorResponse(t, recorder, http.StatusBadRequest, "invalid_argument")
}

func assertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, statusCode int, code string) {
	t.Helper()

	if recorder.Code != statusCode {
		t.Fatalf("expected status %d, got %d with body %s", statusCode, recorder.Code, recorder.Body.String())
	}

	var response core_http_response.ErrorResponse
	decodeJSONResponse(t, recorder, &response)
	if response.Code != code {
		t.Fatalf("expected error code %q, got %q", code, response.Code)
	}
}

func decodeJSONResponse(t *testing.T, recorder *httptest.ResponseRecorder, dest any) {
	t.Helper()

	if err := json.NewDecoder(recorder.Body).Decode(dest); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
}
