package cachex

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache is the interface that wraps the basic Get, Set, and Delete methods.
type Cache interface {
	Set(ctx context.Context, ns, key, value string, expiration ...time.Duration) error
	Get(ctx context.Context, ns, key string) (string, bool, error)
	GetAndDelete(ctx context.Context, ns, key string) (string, bool, error)
	Exists(ctx context.Context, ns, key string) (bool, error)
	Delete(ctx context.Context, ns, key string) error
	Iterator(ctx context.Context, ns string, fn func(ctx context.Context, key, value string) bool) error
	Close(ctx context.Context) error
	Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error)
	RPop(ctx context.Context, ns, key string) (string, error)
	LPush(ctx context.Context, ns, key, value string) error
	TxPipeline() redis.Pipeliner
	Pipeline() redis.Pipeliner
}

var defaultDelimiter = ":"

type options struct {
	Delimiter string
}

type Option func(*options)

func WithDelimiter(delimiter string) Option {
	return func(o *options) {
		o.Delimiter = delimiter
	}
}
