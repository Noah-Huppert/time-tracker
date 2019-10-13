.PHONY: run test db clean-db migrate-db

CONTAINER_CLI ?= podman
HOST_CONTAINER_DATA_DIR ?= ${PWD}/container-data

DB_HOST_DIR ?= ${HOST_CONTAINER_DATA_DIR}/db/
DB_CONTAINER ?= postgres:12-alpine
DB_SSL_MODE ?= disable
DB_HOST ?= localhost:5432
DB_NAME ?= dev-time-tracker
DB_USER ?= dev-time-tracker
DB_PASSWORD ?= dev-time-tracker

DB_MIGRATE_CONTAINER ?= migrate/migrate
DB_MIGRATIONS_HOST_DIR ?= ${PWD}/db-migrations

# Build and run server from source
run:
	go run .

# Run tests
test:
	go test -v ./...

# Start a local development Postgres server
db:
	mkdir -p ${DB_HOST_DIR}
	${CONTAINER_CLI} run \
		-it \
		--rm \
		--net host \
		-v ${DB_HOST_DIR}:/var/lib/postgresql/data \
		-e POSTGRES_DB=${DB_NAME} \
		-e POSTGRES_USER=${DB_USER} \
		-e POSTGRES_PASSWORD=${DB_PASSWORD} \
		${DB_CONTAINER}

# Clears local Postgres server's data
clean-db:
	sudo rm -rf ${HOST_CONTAINER_DATA_DIR}

# Run migrations on database
migrate-db:
	${CONTAINER_CLI} run \
		-v ${DB_MIGRATIONS_HOST_DIR}:/mnt \
		--network host \
		${DB_MIGRATE_CONTAINER} \
		-path /mnt \
		-database postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}/${DB_NAME}?sslmode=${DB_SSL_MODE} up
