package queue

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/redis/go-redis/v9"
)

var queue RedisQueue

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not create pool: %s", err)
	}

	if err = pool.Client.Ping(); err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name:       "asyncq-redis-tests",
		Repository: "redis",
		Tag:        "latest",
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start Redis container: %s", err)
	}

	resource.Expire(60)

	var client *redis.Client
	if err := pool.Retry(func() error {
		hostPort := resource.GetHostPort("6379/tcp")
		opt, err := redis.ParseURL(fmt.Sprintf("redis://%s", hostPort))
		if err != nil {
			log.Fatalf("Could not parse redis url: %v", err)
		}
		client = redis.NewClient(opt)

		_, err = client.Ping(context.Background()).Result()
		if err != nil {
			log.Fatalf("Could not ping redis: %v", err)
		}

		log.Println("Redis connection successful.")
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to test redis after retry: %s", err)
	}

	queue = *NewRedisQueue(slog.New(slog.NewTextHandler(os.Stdout, nil)), client)

	// Run tests
	code := m.Run()

	// Cleanup
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestRedisEnqueue(t *testing.T) {
	t.Skip()
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		jobId := uuid.NewString()
		runAt := time.Now().Add(1 * time.Hour)

		err := queue.Enqueue(t.Context(), jobId, runAt)
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}

		score, err := queue.Redis.ZScore(t.Context(), defaultQueue, jobId).Result()
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}

		expectedScore := float64(runAt.Unix())
		if score != expectedScore {
			t.Fatalf("expected a score of %f, got %f", expectedScore, score)
		}

	})
}

func TestRedisDequeue(t *testing.T) {
	t.Skip()
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		now := time.Now()

		// job to run now
		jobId1 := uuid.NewString()
		runAt1 := now

		// job to run in 1 hour
		jobId2 := uuid.NewString()
		runAt2 := now.Add(1 * time.Hour)

		// job to run in 2 hours
		jobId3 := uuid.NewString()
		runAt3 := now.Add(2 * time.Hour)

		err := queue.Enqueue(t.Context(), jobId1, runAt1)
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}

		err = queue.Enqueue(t.Context(), jobId2, runAt2)
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}

		err = queue.Enqueue(t.Context(), jobId3, runAt3)
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}

		// dequeue all jobs that are scheduled to run until 1 hour ahead
		jobIds, err := queue.Dequeue(t.Context(), now.Add(1*time.Hour))
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}

		if len(jobIds) != 2 {
			t.Fatalf("expected 2 dequeued jobs, got %d - jobIds=%v", len(jobIds), jobIds)
		}

		expectedJobIds := []string{jobId1, jobId2}
		if slices.Compare(jobIds, expectedJobIds) != 0 {
			t.Fatalf("expected dequeued jobIds=%v, got %v", expectedJobIds, jobIds)
		}
	})
}

func TestRedisRemove(t *testing.T) {
	t.Skip()
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		jobId := uuid.NewString()
		runAt := time.Now().Add(1 * time.Hour)

		err := queue.Enqueue(t.Context(), jobId, runAt)
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}

		err = queue.Remove(t.Context(), jobId)
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}

		_, err = queue.Redis.ZScore(t.Context(), defaultQueue, jobId).Result()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}
