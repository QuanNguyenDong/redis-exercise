# Redis Exercises in Go

Five progressive exercises covering core Redis patterns. Each lives in its own package with a test file — fill in the implementation to make the tests pass.

## Setup

```bash
go mod init redis-exercise
go get github.com/redis/go-redis/v9

# Start Redis locally
docker run -d -p 6379:6379 redis:7
```

---

## Exercise 1: Session Store

**Concepts:** `SET`, `GET`, `DEL`, `EXPIRE`, `TTL`, JSON marshaling

Build a session store backed by Redis.

```
session/
  session.go      ← implement this
  session_test.go
```

### Interface to implement

```go
type Store interface {
    Set(ctx context.Context, userID string, data map[string]string, ttl time.Duration) error
    Get(ctx context.Context, userID string) (map[string]string, error)
    Delete(ctx context.Context, userID string) error
    Extend(ctx context.Context, userID string, ttl time.Duration) error
}
```

### Requirements

- Keys should be namespaced: `session:<userID>`
- `Get` on a missing or expired key returns `nil, nil`
- `Extend` returns an error if the session does not exist
- Store the map as a JSON string

---

## Exercise 2: Rate Limiter

**Concepts:** `INCR`, `EXPIRE`, `SETNX`, atomic operations, fixed window

Implement a fixed-window rate limiter. Each client gets a quota of N requests per time window.

```
ratelimiter/
  ratelimiter.go      ← implement this
  ratelimiter_test.go
```

### Interface to implement

```go
type Limiter interface {
    // Allow returns true if the request is within the rate limit.
    Allow(ctx context.Context, clientID string, limit int, window time.Duration) (bool, error)

    // Remaining returns how many requests the client can still make in the current window.
    Remaining(ctx context.Context, clientID string, limit int) (int, error)
}
```

### Requirements

- Key format: `ratelimit:<clientID>`
- The window starts on the first request and expires after `window` duration
- `INCR` + set `EXPIRE` only on first increment (use `SETNX` or check TTL)
- `Remaining` returns 0 if the client has exceeded the limit, never negative

---

## Exercise 3: Leaderboard

**Concepts:** `ZADD`, `ZRANK`, `ZREVRANK`, `ZRANGE WITHSCORES`, `ZRANGEBYSCORE`

Build a game leaderboard using a Redis sorted set.

```
leaderboard/
  leaderboard.go      ← implement this
  leaderboard_test.go
```

### Types and interface to implement

```go
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
```

### Requirements

- Use a single sorted set key, e.g. `leaderboard:global`
- `GetRank` returns an error if the player does not exist
- `GetTopN` and `GetScoreRange` return an empty slice (not an error) when no results match

---

## Exercise 4: Pub/Sub Chat Room

**Concepts:** `PUBLISH`, `SUBSCRIBE`, goroutines, Go channels

Build a simple pub/sub message bus where multiple subscribers can listen to a named room.

```
pubsub/
  pubsub.go      ← implement this
  pubsub_test.go
```

### Interface to implement

```go
type Bus interface {
    // Publish sends a message to a room.
    Publish(ctx context.Context, room, message string) error

    // Subscribe listens to a room. Each received message is passed to handler.
    // Returns a cancel function that stops the subscription.
    Subscribe(ctx context.Context, room string, handler func(msg string)) (cancel context.CancelFunc, err error)
}
```

### Requirements

- Channel name format: `chat:<room>`
- `Subscribe` must be non-blocking — run the handler in a background goroutine
- Calling `cancel()` must cleanly stop the goroutine and unsubscribe from Redis
- Write a `main_test.go` scenario: 2 subscribers on the same room, 1 publisher sends 3 messages, assert both subscribers receive all 3

---

## Exercise 5: Distributed Lock

**Concepts:** `SET NX EX`, Lua scripting with `EVAL`, compare-and-delete

Implement a distributed mutex using Redis so only one process holds the lock at a time.

```
distlock/
  distlock.go      ← implement this
  distlock_test.go
```

### Interface to implement

```go
type Locker interface {
    // Acquire tries to obtain the lock. Returns a unique token if successful.
    // ok=false means the lock is already held.
    Acquire(ctx context.Context, key string, ttl time.Duration) (token string, ok bool, err error)

    // Release releases the lock only if the token matches the one used to acquire it.
    // Returns an error if the token does not match or the lock has expired.
    Release(ctx context.Context, key string, token string) error
}
```

### Requirements

- Use `SET key token NX EX ttl` for atomic acquire
- Use a Lua script for `Release` to atomically check the token and delete: if the stored value matches the token, delete; otherwise return an error
- Generate `token` as a UUID or random hex string
- **Concurrency test:** spawn 20 goroutines each trying to acquire the same lock and increment a shared counter. Assert the final counter equals the number of goroutines that successfully acquired the lock, and no double-increments occurred

---

## Progression

| # | Exercise       | Key Redis features                        | Difficulty |
|---|----------------|-------------------------------------------|------------|
| 1 | Session Store  | String, TTL, JSON                         | Beginner   |
| 2 | Rate Limiter   | INCR, atomic SETNX, fixed window          | Beginner   |
| 3 | Leaderboard    | Sorted sets (ZADD, ZRANK, ZRANGE)         | Intermediate |
| 4 | Pub/Sub        | PUBLISH/SUBSCRIBE, goroutines             | Intermediate |
| 5 | Distributed Lock | SET NX EX, Lua scripting, concurrency   | Advanced   |
