ENV_FILE := .env

# build bins
build:
	go build -o bin/service ./cmd/service
	go build -o bin/producer ./producer

# Generate 100% valid JSON orders
generate-orders:
	python3 gen_orders.py -n 600 --invalid-rate 0.0 -o orders.json

# Run producer in background
run-producer:
	./bin/producer --file orders.json --rps 1

# Run service in foreground
run-service:
	./bin/service 

# run all tests no cache
test:
	go test -v ./... -count=1

# run test containers
up-t:
	docker compose up -d postgres-test zookeeper-test broker-test

# Alias for up_t, as these are the main dev containers
up: up-t

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


dev: build generate-orders up migrate run-producer run-service

# tearing down and cleanup
clean:
	docker compose down -v --remove-orphans
	rm -f orders.json
	rm -rf bin/