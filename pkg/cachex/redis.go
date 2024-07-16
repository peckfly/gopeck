package cachex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisCacheWithClient Use redis client create cache
func NewRedisCacheWithClient(cli *redis.Client, opts ...Option) Cache {
	return newRedisCache(cli, opts...)
}

// NewRedisCacheWithClusterClient Use redis cluster client create cache
func NewRedisCacheWithClusterClient(cli *redis.ClusterClient, opts ...Option) Cache {
	return newRedisCache(cli, opts...)
}

func newRedisCache(cli redisClient, opts ...Option) Cache {
	defaultOpts := &options{
		Delimiter: defaultDelimiter,
	}
	for _, o := range opts {
		o(defaultOpts)
	}
	return &redisCache{
		opts: defaultOpts,
		cli:  cli,
	}
}

type redisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	Close() error
	Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error)
	RPop(ctx context.Context, key string) *redis.StringCmd
	LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	TxPipeline() redis.Pipeliner
	Pipeline() redis.Pipeliner
}

type redisCache struct {
	opts *options
	cli  redisClient
}

func (a *redisCache) getKey(ns, key string) string {
	return fmt.Sprintf("%s%s%s", ns, a.opts.Delimiter, key)
}

func (a *redisCache) Set(ctx context.Context, ns, key, value string, expiration ...time.Duration) error {
	var exp time.Duration
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	cmd := a.cli.Set(ctx, a.getKey(ns, key), value, exp)
	return cmd.Err()
}

func (a *redisCache) Get(ctx context.Context, ns, key string) (string, bool, error) {
	cmd := a.cli.Get(ctx, a.getKey(ns, key))
	if err := cmd.Err(); err != nil {
		if err == redis.Nil {
			return "", false, nil
		}
		return "", false, err
	}
	return cmd.Val(), true, nil
}

func (a *redisCache) Exists(ctx context.Context, ns, key string) (bool, error) {
	cmd := a.cli.Exists(ctx, a.getKey(ns, key))
	if err := cmd.Err(); err != nil {
		return false, err
	}
	return cmd.Val() > 0, nil
}

func (a *redisCache) Delete(ctx context.Context, ns, key string) error {
	b, err := a.Exists(ctx, ns, key)
	if err != nil {
		return err
	} else if !b {
		return nil
	}

	cmd := a.cli.Del(ctx, a.getKey(ns, key))
	if err := cmd.Err(); err != nil && err != redis.Nil {
		return err
	}
	return nil
}

func (a *redisCache) GetAndDelete(ctx context.Context, ns, key string) (string, bool, error) {
	value, ok, err := a.Get(ctx, ns, key)
	if err != nil {
		return "", false, err
	} else if !ok {
		return "", false, nil
	}

	cmd := a.cli.Del(ctx, a.getKey(ns, key))
	if err := cmd.Err(); err != nil && err != redis.Nil {
		return "", false, err
	}
	return value, true, nil
}

func (a *redisCache) Iterator(ctx context.Context, ns string, fn func(ctx context.Context, key, value string) bool) error {
	var cursor uint64 = 0

LbLoop:
	for {
		cmd := a.cli.Scan(ctx, cursor, a.getKey(ns, "*"), 100)
		if err := cmd.Err(); err != nil {
			return err
		}

		keys, c, err := cmd.Result()
		if err != nil {
			return err
		}

		for _, key := range keys {
			cmd := a.cli.Get(ctx, key)
			if err := cmd.Err(); err != nil {
				if err == redis.Nil {
					continue
				}
				return err
			}
			if next := fn(ctx, strings.TrimPrefix(key, a.getKey(ns, "")), cmd.Val()); !next {
				break LbLoop
			}
		}

		if c == 0 {
			break
		}
		cursor = c
	}

	return nil
}

func (a *redisCache) Close(ctx context.Context) error {
	return a.cli.Close()
}

func (a *redisCache) Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return a.cli.Pipelined(ctx, fn)
}

func (a *redisCache) RPop(ctx context.Context, ns, key string) (string, error) {
	cmd := a.cli.RPop(ctx, a.getKey(ns, key))
	if err := cmd.Err(); err != nil {
		return "", err
	}
	return cmd.Val(), nil
}

func (a *redisCache) LPush(ctx context.Context, ns, key, value string) error {
	cmd := a.cli.LPush(ctx, a.getKey(ns, key), value)
	return cmd.Err()
}

func (a *redisCache) TxPipeline() redis.Pipeliner {
	return a.cli.TxPipeline()
}

func (a *redisCache) Pipeline() redis.Pipeliner {
	return a.cli.Pipeline()
}
