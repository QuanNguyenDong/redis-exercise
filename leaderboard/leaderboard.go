package leaderboard

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type PlayerScore struct {
	PlayerID string
	Score    float64
}

type Leaderboard interface {
	// AddScore upserts a player's score (higher score wins).
	AddScore(ctx context.Context, playerID string, score float64) error

	// GetRank returns the player's rank (1 = highest score).
	GetRank(ctx context.Context, playerID string) (int64, error)

	// GetTopN returns the top N players, rank 1 first.
	GetTopN(ctx context.Context, n int) ([]PlayerScore, error)

	// GetScoreRange returns all players with scores between min and max (inclusive).
	GetScoreRange(ctx context.Context, min, max float64) ([]PlayerScore, error)
}

type RedisLeaderboard struct {
	Client *redis.Client
}

func NewLeaderboard() *RedisLeaderboard {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	return &RedisLeaderboard{Client: client}
}

// TODO: implement AddScore
// Hint: use Client.ZAdd with redis.Z{Score: score, Member: playerID}
// Key: "leaderboard:global"
func (lb *RedisLeaderboard) AddScore(ctx context.Context, playerID string, score float64) error {
	err := lb.Client.ZAdd(ctx, "leaderboard:global", redis.Z{Score: score, Member: playerID}).Err()
	if err != nil {
		panic(err)
	}

	return nil
}

// TODO: implement GetRank
// Hint: sorted sets rank from lowest score by default — use ZRevRank for highest-first ranking
// Add 1 to convert from 0-based index to 1-based rank
// Return an error if the player does not exist (ZRevRank returns redis.Nil)
func (lb *RedisLeaderboard) GetRank(ctx context.Context, playerID string) (int64, error) {
	rank, err := lb.Client.ZRevRank(ctx, "leaderboard:global", playerID).Result()
	return rank + 1, err
}

// TODO: implement GetTopN
// Hint: use ZRevRangeWithScores to get top N members in descending score order
// Return an empty slice (not nil, not an error) when no results match
func (lb *RedisLeaderboard) GetTopN(ctx context.Context, n int) ([]PlayerScore, error) {
	members, err := lb.Client.ZRevRangeWithScores(ctx, "leaderboard:global", 0, int64(n-1)).Result()
	if err != nil {
		return []PlayerScore{}, err
	}

	top := make([]PlayerScore, len(members))
	for i, z := range members {
		top[i] = PlayerScore{PlayerID: z.Member.(string), Score: z.Score}
	}
	return top, nil
}

// TODO: implement GetScoreRange
// Hint: use ZRangeByScoreWithScores with redis.ZRangeBy{Min: ..., Max: ...}
// strconv.FormatFloat or fmt.Sprintf can convert float64 to the string Redis expects
// Return an empty slice (not nil, not an error) when no results match
func (lb *RedisLeaderboard) GetScoreRange(ctx context.Context, min, max float64) ([]PlayerScore, error) {
	members, err := lb.Client.ZRangeByScoreWithScores(ctx, "leaderboard:global", &redis.ZRangeBy{
		Min: fmt.Sprintf("%g", min),
		Max: fmt.Sprintf("%g", max),
	}).Result()
	if err != nil {
		return []PlayerScore{}, err
	}

	result := make([]PlayerScore, len(members))
	for i, z := range members {
		result[i] = PlayerScore{PlayerID: z.Member.(string), Score: z.Score}
	}
	return result, nil
}
