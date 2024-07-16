package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/peckfly/gopeck/internal/mods/common/repo"
	"github.com/peckfly/gopeck/pkg/cachex"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

const (
	QueNs                = "que_ns"
	queAggregateKey      = "gopeck:stress:que:aggregate:%d"
	rateSecondKey        = "gopeck:stress:que:rate:%d"
	taskCostNodeCountKey = "gopeck:stress:task:node:count:%d"
)

type (
	queRepository struct {
		Client cachex.Cache
	}
)

func NewQueRepository(cache cachex.Cache) repo.QueRepository {
	return &queRepository{
		Client: cache,
	}
}

func (s *queRepository) AggregatePush(ctx context.Context, taskId uint64, result *repo.Aggregate) error {
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}
	err = s.Client.LPush(ctx, QueNs, fmt.Sprintf(queAggregateKey, taskId), string(b))
	if err != nil {
		return err
	}
	return nil
}

func (s *queRepository) AggregatePop(ctx context.Context, taskId uint64) (*repo.Aggregate, error) {
	b, err := s.Client.RPop(ctx, QueNs, fmt.Sprintf(queAggregateKey, taskId))
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var result repo.Aggregate
	err = json.Unmarshal([]byte(b), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *queRepository) AggregateClear(ctx context.Context, taskId uint64) error {
	err := s.Client.Delete(ctx, QueNs, fmt.Sprintf(queAggregateKey, taskId))
	return err
}

func (s *queRepository) RatePush(ctx context.Context, taskId uint64, result *repo.Aggregate) error {
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}
	err = s.Client.LPush(ctx, QueNs, fmt.Sprintf(rateSecondKey, taskId), string(b))
	if err != nil {
		return err
	}
	return nil
}

func (s *queRepository) RatePop(ctx context.Context, taskId uint64) (*repo.Aggregate, error) {
	b, err := s.Client.RPop(ctx, QueNs, fmt.Sprintf(rateSecondKey, taskId))
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var result repo.Aggregate
	err = json.Unmarshal([]byte(b), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *queRepository) RateClear(ctx context.Context, taskId uint64) error {
	err := s.Client.Delete(ctx, QueNs, fmt.Sprintf(rateSecondKey, taskId))
	return err
}

func (s *queRepository) BatchSetTaskNodeCount(ctx context.Context, counts map[uint64]int) error {
	_, err := s.Client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for taskId, count := range counts {
			pipe.Set(ctx, fmt.Sprintf(taskCostNodeCountKey, taskId), count, time.Hour*24*3)
		}
		return nil
	})
	return err
}

func (s *queRepository) GetTaskNodeCount(ctx context.Context, taskId uint64) (int, error) {
	value, b, err := s.Client.Get(ctx, QueNs, fmt.Sprintf(taskCostNodeCountKey, taskId))
	if err != nil {
		return 0, err
	}
	if !b {
		return 0, nil
	}
	iValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return iValue, nil
}
