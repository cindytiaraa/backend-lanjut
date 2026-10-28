package repository

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/topology"
)

func TestIsUnavailable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "MongoDB selection failure", err: topology.ServerSelectionError{}, want: true},
		{name: "MongoDB network failure", err: mongo.CommandError{Labels: []string{"NetworkError"}}, want: true},
		{name: "request timeout", err: context.DeadlineExceeded, want: true},
		{name: "domain error", err: ErrNotFound, want: false},
		{name: "unrelated error", err: errors.New("invalid document"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnavailable(tt.err); got != tt.want {
				t.Errorf("IsUnavailable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
