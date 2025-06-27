package queue

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultQueue = "default"

var atomicDequeueScript = redis.NewScript(`
local jobs = redis.call("ZRANGEBYSCORE", KEYS[1], ARGV[1], ARGV[2])
if #jobs > 0 then
  redis.call("ZREM", KEYS[1], unpack(jobs))
end
return jobs
`)

type RedisQueue struct {
	Redis  *redis.Client
	logger *slog.Logger
}

func NewRedisQueue(logger *slog.Logger, redis *redis.Client) *RedisQueue {
	return &RedisQueue{
		Redis:  redis,
		logger: logger,
	}
}

func (d *RedisQueue) Enqueue(ctx context.Context, jobId string, runAt time.Time) error {
	return d.Redis.ZAdd(ctx, defaultQueue, redis.Z{
		Score:  float64(runAt.Unix()),
		Member: jobId,
	}).Err()
}

func (d *RedisQueue) Dequeue(ctx context.Context, timeThreshold time.Time) ([]string, error) {
	score := float64(timeThreshold.Unix())

	ids, err := d.atomicDequeue(ctx, score)
	if err != nil {
		return nil, err
	}

	// no jobs schedule to run at this point
	if len(ids) == 0 {
		return nil, nil
	}

	return ids, nil
}

func (d *RedisQueue) Remove(ctx context.Context, jobId string) error {
	return d.Redis.ZRem(ctx, defaultQueue, []string{jobId}).Err()
}

func (d *RedisQueue) atomicDequeue(ctx context.Context, maxScore float64) ([]string, error) {
	result, err := atomicDequeueScript.Run(ctx, d.Redis, []string{defaultQueue}, 0, maxScore).Result()
	if err != nil {
		return nil, err
	}

	rawIds, ok := result.([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected script result type: %T", result)
	}

	ids := make([]string, len(rawIds))
	for i, val := range rawIds {
		ids[i] = val.(string)
	}
	return ids, nil
}
