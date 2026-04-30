package grpc //nolint:testpackage // it's intended to be a test package of a private functions

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/genvmoroz/lale/service/internal/core"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestResolveCoreError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		wantCode codes.Code
	}{
		{
			name:     "context canceled",
			err:      context.Canceled,
			wantCode: codes.Canceled,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			wantCode: codes.DeadlineExceeded,
		},
		{
			name: "validation",
			err: fmt.Errorf("%w: %w",
				core.NewValidationError(),
				errors.New("missing field"),
			),
			wantCode: codes.InvalidArgument,
		},
		{
			name: "not found",
			err: fmt.Errorf("%w: card gone",
				core.NewNotFoundError(),
			),
			wantCode: codes.NotFound,
		},
		{
			name: "already exists",
			err: fmt.Errorf("%w: dup",
				core.NewAlreadyExistsError(),
			),
			wantCode: codes.AlreadyExists,
		},
		{
			name: "failed precondition",
			err: fmt.Errorf("%w: card already learnt",
				core.NewFailedPreconditionError(),
			),
			wantCode: codes.FailedPrecondition,
		},
		{
			name:     "internal fallback",
			err:      errors.New("mongodb exploded"),
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := resolveCoreError(tt.err)
			st, ok := status.FromError(got)
			if !ok {
				t.Fatalf("resolveCoreError returned non-status error: %v", got)
			}
			if st.Code() != tt.wantCode {
				t.Fatalf("code = %v, want %v", st.Code(), tt.wantCode)
			}
		})
	}
}
