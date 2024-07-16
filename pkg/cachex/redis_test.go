package cachex

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRedisCache(t *testing.T) {
	assert := assert.New(t)

	cache := NewRedisCacheWithClient(redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "happy123",
		DB:       0,
	}))

	ctx := context.Background()
	err := cache.Set(ctx, "tt", "foo", "bar")
	assert.Nil(err)

	val, exists, err := cache.Get(ctx, "tt", "foo")
	assert.Nil(err)
	assert.True(exists)
	assert.Equal("bar", val)

	err = cache.Delete(ctx, "tt", "foo")
	assert.Nil(err)

	val, exists, err = cache.Get(ctx, "tt", "foo")
	assert.Nil(err)
	assert.False(exists)
	assert.Equal("", val)

	tmap := make(map[string]bool)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("foo%d", i)
		err = cache.Set(ctx, "tt", key, "bar")
		assert.Nil(err)
		tmap[key] = true

		err = cache.Set(ctx, "ff", key, "bar")
		assert.Nil(err)
	}

	err = cache.Iterator(ctx, "tt", func(ctx context.Context, key, value string) bool {
		assert.True(tmap[key])
		assert.Equal("bar", value)
		return true
	})
	assert.Nil(err)

	err = cache.Close(ctx)
	assert.Nil(err)
}
