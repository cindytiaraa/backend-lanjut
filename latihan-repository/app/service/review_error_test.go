package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gofiber/fiber/v2"

	"latihan-repository/helper"
)

func TestReviewTimeoutIsMappedAccordingToDatabase(t *testing.T) {
	tests := []struct {
		name string
		got  error
		want int
	}{
		{
			name: "MongoDB timeout",
			got:  translateMongoReviewError(context.DeadlineExceeded),
			want: fiber.StatusServiceUnavailable,
		},
		{
			name: "PostgreSQL timeout",
			got:  translateReviewError(context.DeadlineExceeded),
			want: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var appErr *helper.AppError
			if !errors.As(tt.got, &appErr) {
				t.Fatalf("error %T does not wrap helper.AppError", tt.got)
			}
			if appErr.Status != tt.want {
				t.Errorf("status = %d, want %d", appErr.Status, tt.want)
			}
		})
	}
}
