package pagination_test

import (
	"errors"
	"testing"

	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
	"github.com/horizoonn/todoapp/internal/core/pagination"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name          string
		limit         *int
		offset        *int
		defaultLimit  int
		maxLimit      int
		defaultOffset int
		want          pagination.Page
		wantErr       bool
	}{
		{
			name:          "uses defaults when params are nil",
			defaultLimit:  100,
			maxLimit:      100,
			defaultOffset: 0,
			want: pagination.Page{
				Limit:  100,
				Offset: 0,
			},
		},
		{
			name:          "uses provided params",
			limit:         intPtr(25),
			offset:        intPtr(50),
			defaultLimit:  100,
			maxLimit:      100,
			defaultOffset: 0,
			want: pagination.Page{
				Limit:  25,
				Offset: 50,
			},
		},
		{
			name:          "caps limit by max limit",
			limit:         intPtr(200),
			offset:        intPtr(0),
			defaultLimit:  100,
			maxLimit:      100,
			defaultOffset: 0,
			want: pagination.Page{
				Limit:  100,
				Offset: 0,
			},
		},
		{
			name:          "zero limit is allowed",
			limit:         intPtr(0),
			offset:        intPtr(0),
			defaultLimit:  100,
			maxLimit:      100,
			defaultOffset: 0,
			want: pagination.Page{
				Limit:  0,
				Offset: 0,
			},
		},
		{
			name:          "rejects negative limit",
			limit:         intPtr(-1),
			defaultLimit:  100,
			maxLimit:      100,
			defaultOffset: 0,
			wantErr:       true,
		},
		{
			name:          "rejects negative offset",
			offset:        intPtr(-1),
			defaultLimit:  100,
			maxLimit:      100,
			defaultOffset: 0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pagination.Normalize(tt.limit, tt.offset, tt.defaultLimit, tt.maxLimit, tt.defaultOffset)
			if tt.wantErr {
				if !errors.Is(err, core_errors.ErrInvalidArgument) {
					t.Fatalf("expected ErrInvalidArgument, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected page %+v, got %+v", tt.want, got)
			}
		})
	}
}

func intPtr(value int) *int {
	return &value
}
