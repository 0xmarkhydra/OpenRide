package store

import (
	"testing"
	"time"
)

func TestOutboxBackoffIsBoundedExponential(t *testing.T) {
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{attempt: 0, want: time.Second},
		{attempt: 1, want: time.Second},
		{attempt: 2, want: 2 * time.Second},
		{attempt: 3, want: 4 * time.Second},
		{attempt: 9, want: 256 * time.Second},
		{attempt: 10, want: 256 * time.Second},
		{attempt: 50, want: 256 * time.Second},
	}
	for _, tc := range cases {
		if got := outboxBackoff(tc.attempt); got != tc.want {
			t.Fatalf("attempt %d: got %s want %s", tc.attempt, got, tc.want)
		}
		if got := outboxBackoff(tc.attempt); got > outboxMaxBackoff {
			t.Fatalf("attempt %d exceeded max backoff: %s", tc.attempt, got)
		}
	}
}

func TestOutboxLeaseTokenIsNonEmptyAndChanges(t *testing.T) {
	first, err := newOutboxLeaseToken()
	if err != nil {
		t.Fatal(err)
	}
	second, err := newOutboxLeaseToken()
	if err != nil {
		t.Fatal(err)
	}
	if first == "" || second == "" || first == second {
		t.Fatalf("unexpected lease tokens: %q %q", first, second)
	}
}
