package leaderboard_test

import (
	"context"
	"testing"

	"redis-exercise/leaderboard"
)

func setup(t *testing.T) *leaderboard.RedisLeaderboard {
	t.Helper()
	lb := leaderboard.NewLeaderboard()

	// Clean slate before each test
	lb.Client.Del(context.Background(), "leaderboard:global")

	t.Cleanup(func() {
		lb.Client.Del(context.Background(), "leaderboard:global")
		lb.Client.Close()
	})
	return lb
}

func TestAddAndGetRank(t *testing.T) {
	lb := setup(t)
	ctx := context.Background()

	if err := lb.AddScore(ctx, "alice", 100); err != nil {
		t.Fatal(err)
	}
	if err := lb.AddScore(ctx, "bob", 200); err != nil {
		t.Fatal(err)
	}
	if err := lb.AddScore(ctx, "carol", 150); err != nil {
		t.Fatal(err)
	}

	// bob has the highest score → rank 1
	rank, err := lb.GetRank(ctx, "bob")
	if err != nil {
		t.Fatal(err)
	}
	if rank != 1 {
		t.Errorf("expected bob rank 1, got %d", rank)
	}

	// carol is second
	rank, err = lb.GetRank(ctx, "carol")
	if err != nil {
		t.Fatal(err)
	}
	if rank != 2 {
		t.Errorf("expected carol rank 2, got %d", rank)
	}
}

func TestGetRankMissingPlayer(t *testing.T) {
	lb := setup(t)
	ctx := context.Background()

	_, err := lb.GetRank(ctx, "ghost")
	if err == nil {
		t.Error("expected error for missing player, got nil")
	}
}

func TestAddScoreUpserts(t *testing.T) {
	lb := setup(t)
	ctx := context.Background()

	lb.AddScore(ctx, "alice", 50)
	lb.AddScore(ctx, "alice", 300) // update, not duplicate

	top, err := lb.GetTopN(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(top) != 1 {
		t.Errorf("expected 1 player, got %d", len(top))
	}
	if top[0].Score != 300 {
		t.Errorf("expected score 300, got %f", top[0].Score)
	}
}

func TestGetTopN(t *testing.T) {
	lb := setup(t)
	ctx := context.Background()

	lb.AddScore(ctx, "alice", 100)
	lb.AddScore(ctx, "bob", 200)
	lb.AddScore(ctx, "carol", 150)

	top2, err := lb.GetTopN(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(top2) != 2 {
		t.Fatalf("expected 2 results, got %d", len(top2))
	}
	if top2[0].PlayerID != "bob" {
		t.Errorf("expected bob first, got %s", top2[0].PlayerID)
	}
	if top2[1].PlayerID != "carol" {
		t.Errorf("expected carol second, got %s", top2[1].PlayerID)
	}
}

func TestGetTopNEmpty(t *testing.T) {
	lb := setup(t)
	ctx := context.Background()

	results, err := lb.GetTopN(ctx, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty slice, got %d items", len(results))
	}
}

func TestGetScoreRange(t *testing.T) {
	lb := setup(t)
	ctx := context.Background()

	lb.AddScore(ctx, "alice", 100)
	lb.AddScore(ctx, "bob", 200)
	lb.AddScore(ctx, "carol", 150)

	// scores 100–150 inclusive: alice and carol
	results, err := lb.GetScoreRange(ctx, 100, 150)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// no player has score in 300–400
	empty, err := lb.GetScoreRange(ctx, 300, 400)
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 {
		t.Errorf("expected empty slice, got %d items", len(empty))
	}
}
