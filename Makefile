ENV_FILE := .env

# build bins
build:
	go build -o bin/service ./cmd/service
	go build -o bin/producer ./producer

# Generate 100% valid JSON orders
gen:
	python3 gen_orders.py -n 600 --invalid-rate 0.0 -o orders.json
# Run main containers 
up: 
	docker compose up -d postgres zookeeper broker

# migrate psql using goose; it wait for the psql to be up
migrate:
	@env $(shell grep -v '^#' $(ENV_FILE) | xargs) bash -c '\
		while ! PGPASSWORD=$$POSTGRES_PASSWORD psql -h $$POSTGRES_HOST -p $$POSTGRES_PORT -U $$POSTGRES_USER -d $$POSTGRES_DB -c "\q" >/dev/null 2>&1; do \
			echo "Waiting for Postgres to be ready..."; \
			sleep 2; \
		done; \
		echo "Postgres is ready - running migrations"; \
		goose -dir sql postgres "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$$POSTGRES_HOST:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" up; \
	'
	
# sets up everything before you run the producer and the service
dev: build gen-o up migrate

# Run producer in background
runp:
	./bin/producer --file orders.json --rps 1

# Run service in foreground
runs:
	./bin/service 

# run all tests no cache
test:
	go test -v ./... -count=1


# tearing down and cleanup
clean:
	docker compose down -v --remove-orphans
	rm -f orders.json
	rm -rf bin/