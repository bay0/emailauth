package emailauth

import (
	"context"
	"encoding/json"
	"time"
)

type RedisCodeStore struct {
	client RedisClient
}

type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
}

func NewRedisCodeStore(client RedisClient) *RedisCodeStore {
	return &RedisCodeStore{client: client}
}

func (r *RedisCodeStore) Set(ctx context.Context, email string, code *AuthCode) error {
	data, err := json.Marshal(code)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, email, string(data), codeExpiration)
}

func (r *RedisCodeStore) Get(ctx context.Context, email string) (*AuthCode, error) {
	data, err := r.client.Get(ctx, email)
	if err != nil {
		return nil, err
	}

	var code AuthCode
	err = json.Unmarshal([]byte(data), &code)
	if err != nil {
		return nil, err
	}

	return &code, nil
}

func (r *RedisCodeStore) Delete(ctx context.Context, email string) error {
	return r.client.Del(ctx, email)
}
