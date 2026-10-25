package main

import "testing"

func TestVerifyTally(t *testing.T) {
	tests := []struct {
		name     string
		actual   map[string]int
		expected map[string]int
		wantErr  bool
	}{
		{
			name:     "matching results",
			actual:   map[string]int{"201 OK": 5, "409 COURSE_FULL": 55},
			expected: map[string]int{"201 OK": 5, "409 COURSE_FULL": 55},
		},
		{
			name:     "unexpected result",
			actual:   map[string]int{"201 OK": 5, "500 INTERNAL_ERROR": 55},
			expected: map[string]int{"201 OK": 5, "409 COURSE_FULL": 55},
			wantErr:  true,
		},
		{
			name:     "missing expected result",
			actual:   map[string]int{"201 OK": 5},
			expected: map[string]int{"201 OK": 5, "409 COURSE_FULL": 55},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyTally(tt.name, tt.actual, tt.expected)
			if (err != nil) != tt.wantErr {
				t.Fatalf("verifyTally() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerifyConsistent(t *testing.T) {
	if err := verifyConsistent("consistent state", true); err != nil {
		t.Fatalf("verifyConsistent() returned error for consistent state: %v", err)
	}
	if err := verifyConsistent("inconsistent state", false); err == nil {
		t.Fatal("verifyConsistent() did not return an error for inconsistent state")
	}
}
