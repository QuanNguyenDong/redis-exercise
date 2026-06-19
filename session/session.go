package session

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store interface {
	Set(ctx context.Context, userID string, data map[string]string, ttl time.Duration) error
	Get(ctx context.Context, userID string) (map[string]string, error)
	Delete(ctx context.Context, userID string) error
	Extend(ctx context.Context, userID string, ttl time.Duration) error
}

type Session struct {
	Client *redis.Client
}

func NewSession() *Session {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	return &Session{Client: client}
}

func (session *Session) Set(ctx context.Context, userID string, data map[string]string, ttl time.Duration) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return session.Client.Set(ctx, "session:" + userID, bytes, ttl).Err()
}

func (session *Session) Get(ctx context.Context, userID string) (map[string]string, error) {
	key := "session:" + userID
	bytes, err := session.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var data map[string]string
	err = json.Unmarshal([]byte(bytes), &data)
	return data, err
}

func (session *Session) Delete(ctx context.Context, userID string) error {
	return session.Client.Del(ctx, "session:"+userID).Err()
}

func (session *Session) Extend(ctx context.Context, userID string, ttl time.Duration) error {
	key := "session:" + userID
	result, err := session.Client.Expire(ctx, key, ttl).Result()
	if err != nil {
		return err
	}
	if !result {
		return redis.Nil
	}
	return nil
}
