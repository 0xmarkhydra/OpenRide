package main

import "testing"

func TestNetworkTarget(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "postgres default", raw: "postgres://user:pass@marketplace-db/marketplace?sslmode=disable", want: "marketplace-db:5432"},
		{name: "postgres explicit", raw: "postgres://user:pass@127.0.0.1:55433/marketplace", want: "127.0.0.1:55433"},
		{name: "nats default", raw: "nats://nats", want: "nats:4222"},
		{name: "nats explicit", raw: "nats://127.0.0.1:4223", want: "127.0.0.1:4223"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := networkTarget(tt.raw)
			if err != nil {
				t.Fatalf("networkTarget(%q): %v", tt.raw, err)
			}
			if got != tt.want {
				t.Fatalf("networkTarget(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestNetworkTargetRejectsUnsupportedURLWithoutPort(t *testing.T) {
	if _, err := networkTarget("redis://cache"); err == nil {
		t.Fatal("expected unsupported scheme without port to fail")
	}
}
