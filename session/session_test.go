package session

import (
	"context"
	"testing"
	"time"
)

var session = NewSession()

func TestSessionSet(t *testing.T) {
	ctx := context.Background()

	err := session.Set(ctx, "user-1", map[string]string{"location": "Sydney"}, 5*time.Second)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSessionGet(t *testing.T) {
	ctx := context.Background()

	// Get existing key
	data, err := session.Get(ctx, "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if data["location"] != "Sydney" {
		t.Fatalf("expected Sydney, got %v", data["location"])
	}

	// Get missing key returns nil, nil
	data, err = session.Get(ctx, "nonexistent-user")
	if err != nil {
		t.Fatalf("expected no error for missing key, got %v", err)
	}
	if data != nil {
		t.Fatalf("expected nil for missing key, got %v", data)
	}
}

func TestSessionDelete(t *testing.T) {
	ctx := context.Background()

	session.Set(ctx, "user-2", map[string]string{"name": "Alice"}, 5*time.Second)

	err := session.Delete(ctx, "user-2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, err := session.Get(ctx, "user-2")
	if err != nil {
		t.Fatalf("expected no error after delete, got %v", err)
	}
	if data != nil {
		t.Fatalf("expected nil after delete, got %v", data)
	}
}

func TestSessionExtend(t *testing.T) {
	ctx := context.Background()

	session.Set(ctx, "user-3", map[string]string{"role": "admin"}, 2*time.Second)

	// Extend an existing session
	err := session.Extend(ctx, "user-3", 10*time.Second)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Extend a non-existent session should return an error
	err = session.Extend(ctx, "nonexistent-user", 10*time.Second)
	if err == nil {
		t.Fatal("expected error for non-existent session, got nil")
	}
}
