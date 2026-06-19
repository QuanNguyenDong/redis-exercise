package ratelimiter

import (
	"context"
	"testing"
	"time"
)

func setup() (*RateLimiter, func()) {
	limiter := NewRateLimiter()
	cleanup := func() {
		limiter.client.FlushDB(context.Background())
	}
	return limiter, cleanup
}

func TestAllow_WithinLimit(t *testing.T) {
	limiter, cleanup := setup()
	defer cleanup()

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		ok, err := limiter.Allow(ctx, "client1", 5, time.Minute)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestAllow_ExceedsLimit(t *testing.T) {
	limiter, cleanup := setup()
	defer cleanup()

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		limiter.Allow(ctx, "client2", 3, time.Minute)
	}

	ok, err := limiter.Allow(ctx, "client2", 3, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("4th request should be denied")
	}
}

func TestAllow_WindowExpiry(t *testing.T) {
	limiter, cleanup := setup()
	defer cleanup()

	ctx := context.Background()
	window := time.Second

	for i := 0; i < 3; i++ {
		limiter.Allow(ctx, "client3", 3, window)
	}

	ok, _ := limiter.Allow(ctx, "client3", 3, window)
	if ok {
		t.Fatal("4th request should be denied before window expires")
	}

	time.Sleep(1100 * time.Millisecond)

	ok, err := limiter.Allow(ctx, "client3", 3, window)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("request should be allowed after window resets")
	}
}

func TestAllow_IsolatedClients(t *testing.T) {
	limiter, cleanup := setup()
	defer cleanup()

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		limiter.Allow(ctx, "clientA", 3, time.Minute)
	}

	ok, err := limiter.Allow(ctx, "clientB", 3, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("clientB should not be affected by clientA's quota")
	}
}

func TestRemaining_NoRequests(t *testing.T) {
	limiter, cleanup := setup()
	defer cleanup()

	remaining, err := limiter.Remaining(context.Background(), "fresh", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remaining != 10 {
		t.Fatalf("expected 10, got %d", remaining)
	}
}

func TestRemaining_AfterRequests(t *testing.T) {
	limiter, cleanup := setup()
	defer cleanup()

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		limiter.Allow(ctx, "client5", 5, time.Minute)
	}

	remaining, err := limiter.Remaining(ctx, "client5", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remaining != 2 {
		t.Fatalf("expected 2 remaining, got %d", remaining)
	}
}

func TestRemaining_NeverNegative(t *testing.T) {
	limiter, cleanup := setup()
	defer cleanup()

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		limiter.Allow(ctx, "client6", 3, time.Minute)
	}

	remaining, err := limiter.Remaining(ctx, "client6", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected 0, got %d", remaining)
	}
}
