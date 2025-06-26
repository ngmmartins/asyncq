package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/ngmmartins/asyncq/internal/bootstrap"
	"github.com/ngmmartins/asyncq/internal/queue"
	"github.com/ngmmartins/asyncq/internal/store/postgres"
	"github.com/ngmmartins/asyncq/internal/util"
	"github.com/ngmmartins/asyncq/internal/worker"
)

type config struct {
	env          string
	logLevel     slog.Leveler
	tickInterval time.Duration
	redis        struct {
		url string
	}
	db postgres.PostgresConfig
}

func main() {
	var cfg config
	parseFlags(&cfg)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	redis := bootstrap.NewRedisClient(logger, cfg.redis.url)
	store := postgres.New(&cfg.db, logger)
	dispatcher := queue.NewDispatcher(logger, redis)

	w := worker.New(store, dispatcher, logger)

	logger.Info("worker started", "env", cfg.env)
	w.Run(context.Background(), cfg.tickInterval)
}

func parseFlags(cfg *config) {
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.DurationVar(&cfg.tickInterval, "tick-interval", 2*time.Second, "How frequentlly the worker will poll jobs from queue")

	var logLevel string
	flag.StringVar(&logLevel, "log-level", "Info", "Log level (Debug|Info|Warn|Error)")

	flag.StringVar(&cfg.redis.url, "redis-url", "", "Redis URL")

	flag.StringVar(&cfg.db.DSN, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.MaxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.MaxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.MaxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	flag.Parse()

	cfg.logLevel = util.ParseLogLevel(logLevel)
}
