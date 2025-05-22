package rdb

import (
	"app/env"
	"app/pkg/encryption"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nhnghia272/gopkg"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Rdb struct {
	client *redis.Client
	cache  gopkg.CacheShard[[]byte]
}

func New() *Rdb {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	opts, err := redis.ParseURL(env.RedisUri)
	if err != nil {
		logrus.Fatalln("Redis", err)
	}

	client := redis.NewClient(opts)

	if ping := client.Ping(ctx); ping.Err() != nil {
		logrus.Fatalln("Redis", ping.Err())
	}

	fmt.Printf("Redis connected %v\n", env.RedisUri)

	s := &Rdb{client, gopkg.NewCacheShard[[]byte](64)}

	gopkg.Async().Go(func() {
		for msg := range s.Subscribe(context.Background(), "clc").Channel() {
			s.cache.Delete(msg.Payload)
		}
	})

	return s
}

func (s Rdb) Uri() string {
	return env.RedisUri
}

func (s Rdb) Instance() *redis.Client {
	return s.client
}

func (s Rdb) TxPipeline() redis.Pipeliner {
	return s.client.TxPipeline()
}

func (s Rdb) Publish(ctx context.Context, key string, val []byte) error {
	return s.client.Publish(ctx, key, val).Err()
}

func (s Rdb) Subscribe(ctx context.Context, keys ...string) *redis.PubSub {
	return s.client.Subscribe(ctx, keys...)
}

func (s Rdb) Keys(ctx context.Context, pattern string) []string {
	return s.client.Keys(ctx, pattern).Val()
}

func (s Rdb) Del(ctx context.Context, keys ...string) error {
	gopkg.LoopFunc(keys, func(key string) { s.Publish(ctx, "clc", []byte(key)) })
	return s.client.Del(ctx, keys...).Err()
}

func (s Rdb) FlushAll(ctx context.Context) error {
	keys := gopkg.FilterFunc(s.Keys(ctx, "*"), func(key string) bool { return !strings.HasPrefix(key, "jti:") })
	return s.Del(ctx, keys...)
}

func (s Rdb) SetBytes(ctx context.Context, key string, val []byte, exp time.Duration) error {
	bytes := []byte(encryption.Encrypt(string(val), key))

	s.cache.Set(key, bytes, exp)
	return s.client.Set(ctx, key, bytes, exp).Err()
}

func (s Rdb) GetBytes(ctx context.Context, key string) ([]byte, error) {
	if val, err := s.cache.Get(key); err == nil {
		return []byte(encryption.Decrypt(string(val), key)), nil
	}
	val, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	s.cache.Set(key, val, s.client.TTL(ctx, key).Val())
	return []byte(encryption.Decrypt(string(val), key)), nil
}

func (s Rdb) Set(ctx context.Context, key string, val any, exp time.Duration) error {
	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return s.SetBytes(ctx, key, bytes, exp)
}

func (s Rdb) Get(ctx context.Context, key string, val any) error {
	bytes, err := s.GetBytes(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, val)
}
