ENV_FILE := .env

# build bins
build:
	go build -o bin/service ./cmd/service
	go build -o bin/producer ./producer

# run bins 
run: build
	./bin/producer &
	./bin/service

# run all tests no cache
test:
	go test -v ./... -count=1

# run test containers 
up_t:
	docker compose up -d postgres-test zookeeper-test broker-test

# migrate psql
migrate:
	@env $(shell grep -v '^#' $(ENV_FILE) | xargs) bash -c '\
		while ! PGPASSWORD=$$POSTGRES_PASSWORD psql -h $$POSTGRES_HOST -p $$POSTGRES_PORT -U $$POSTGRES_USER -d $$POSTGRES_DB -c "\q" >/dev/null 2>&1; do \
			echo "Waiting for Postgres to be ready..."; \
			sleep 2; \
		done; \
		echo "Postgres is ready - running migrations"; \
		goose -dir sql postgres "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$$POSTGRES_HOST:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" up; \
	'


dev: build up_t migrate run

# tearing down
clean:
	docker compose down -v --remove-orphans
	rm -rf bin/
