include .envrc

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api -redis-url=${REDIS_URL} -db-dsn=${ASYNCQ_DB_DSN} -log-level=Debug

## run/worker: run the cmd/worker application
.PHONY: run/worker
run/worker:
	go run ./cmd/worker -redis-url=${REDIS_URL} -db-dsn=${ASYNCQ_DB_DSN} -tick-interval=10s -log-level=Debug -smtp-host=${MAILTRAP_HOST} -smtp-port=25 -smtp-username=${MAILTRAP_USERNAME} -smtp-password=${MAILTRAP_PASSWORD}

## db/psql: connect to redis using redis-cli
.PHONY: redis/cli
redis/cli:
	docker exec -it ${REDIS_CONTAINER_ID} redis-cli -p ${REDIS_CONTAINER_PORT}

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${ASYNCQ_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${ASYNCQ_DB_DSN} up

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: tidy module dependencies and format all .go files
.PHONY: tidy
tidy:
	@echo 'Tidying module dependencies...'
	go mod tidy
	@echo 'Verifying and vendoring module dependencies...'
	go mod verify
	go mod vendor
	@echo 'Formatting .go files...'
	go fmt ./...

## audit: run quality control checks
.PHONY: audit
audit:
	@echo 'Checking module dependencies...'
	go mod tidy -diff
	go mod verify
	@echo 'Vetting code...'
	go vet ./...
	go tool staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## test: run all tests
.PHONY: test
test:
	go test -v ./...