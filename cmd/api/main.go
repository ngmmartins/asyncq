package main

import (
	"flag"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ngmmartins/asyncq/internal/bootstrap"
	"github.com/ngmmartins/asyncq/internal/queue"
	"github.com/ngmmartins/asyncq/internal/service"
	"github.com/ngmmartins/asyncq/internal/store"
	"github.com/ngmmartins/asyncq/internal/store/postgres"
	"github.com/ngmmartins/asyncq/internal/util"
)

type config struct {
	port     int
	env      string
	logLevel slog.Leveler
	redis    struct {
		url string
	}
	db   postgres.PostgresConfig
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config         config
	logger         *slog.Logger
	queue          queue.Queue
	store          store.Store
	jobService     *service.JobService
	tokenService   *service.TokenService
	accountService *service.AccountService
	apiKeyService  *service.APIKeyService
	wg             sync.WaitGroup
}

func main() {
	var cfg config
	parseFlags(&cfg)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.logLevel}))

	redis := bootstrap.NewRedisClient(logger, cfg.redis.url)
	store := postgres.New(&cfg.db, logger)
	queue := queue.NewRedisQueue(logger, redis)
	jobService := service.NewJobService(logger, queue, store)
	tokenService := service.NewTokenService(logger, store)
	accountService := service.NewAccountService(logger, store)
	apiKeyService := service.NewAPIKeyService(logger, store)

	app := &application{
		config:         cfg,
		logger:         logger,
		queue:          queue,
		store:          store,
		jobService:     jobService,
		tokenService:   tokenService,
		accountService: accountService,
		apiKeyService:  apiKeyService,
	}

	err := app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func parseFlags(cfg *config) {
	flag.IntVar(&cfg.port, "port", 4040, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	var logLevel string
	flag.StringVar(&logLevel, "log-level", "Info", "Log level (Debug|Info|Warn|Error)")

	flag.StringVar(&cfg.redis.url, "redis-url", "", "Redis URL")

	flag.StringVar(&cfg.db.DSN, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.MaxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.MaxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.MaxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	cfg.logLevel = util.ParseLogLevel(logLevel)
}
