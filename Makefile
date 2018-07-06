.PHONY: pg proto

# App variables
APP_NAME=time-tracker

# Postgres variables
PG_CONTAINER_NAME=${APP_NAME}-pg

PG_DB=${APP_NAME}-dev

PG_USER=${APP_NAME}-dev
PG_PASSWORD=${APP_NAME}-dev-password

PG_HOST_DATA_DIR=${PWD}/run-data/pg/
PG_GUEST_DATA_DIR=/var/lib/postgresql/data

# Protocol buffers variables
PROTO_USERS_IN_DIR=users/*.proto

# Starts a local PostgreSQL database
pg:
	mkdir -p "${PG_HOST_DATA_DIR}"
	docker run \
		--name "${PG_CONTAINER_NAME}" \
		--rm -it \
		--net host \
		-e POSTGRES_USER="${PG_USER}" \
		-e POSTGRES_PASSWORD="${PG_PASSWORD}" \
		-v "${PG_HOST_DATA_DIR}:${PG_GUEST_DATA_DIR}" \
		postgres

# Compiles protocol buffer files
proto:
	protoc ${PROTO_USERS_IN_DIR} \
		 --go_out=plugins=grpc:.
