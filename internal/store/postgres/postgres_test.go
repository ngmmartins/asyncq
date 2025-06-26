package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ngmmartins/asyncq/internal/store"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/lib/pq"
)

var s store.JobStore

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not create pool: %s", err)
	}

	if err = pool.Client.Ping(); err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "17",
		Env: []string{
			"POSTGRES_PASSWORD=changeme",
			"POSTGRES_USER=asyncq_test",
			"POSTGRES_DB=asyncq_test",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start Postgres container: %s", err)
	}

	resource.Expire(60)

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://asyncq_test:changeme@%s/asyncq_test?sslmode=disable", hostAndPort)
	log.Println("Connecting to database on url: ", databaseUrl)

	pool.MaxWait = 45 * time.Second

	var db *sql.DB
	// Retry connection until it's ready
	if err := pool.Retry(func() error {
		log.Println("Connecting to test database...")
		db, err = sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}

		err = db.Ping()
		if err != nil {
			return err
		}

		log.Println("Successfully connected to test database")
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to test database after retry: %s", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create migration driver: %v", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		"file://../../../migrations",
		"postgres", driver,
	)
	if err != nil {
		log.Fatalf("Could not create migrator: %v", err)
	}

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	s = newPostgresJobStore(&PostgresStore{db: db})

	// Run tests
	code := m.Run()

	// Cleanup
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}
