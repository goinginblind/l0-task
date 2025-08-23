ENV_FILE := .env

# ---- DOCKER COMPOSES && USEFUL THINGS ----

# Generate 90% valid JSON orders
gen:		
# -n [TOTAL_JSON_ENTRIES] --invalid-rate [0<=%_OF_INVALID<=1] -o [NAME OF THE OUTPUT, mock.json IS DEFAULT]
# If on Windows PC probably should use 'python' instead of 'python3'. It doesn't need 'pip isntall's ㄟ( ▔, ▔ )ㄏ
	python3 gen_orders.py -n 3600 --invalid-rate 0.1 -o mock.json

# Run main containers 
up-kafka:
	docker compose up -d zookeeper broker 

up-psql:
	docker compose up -d postgres

up-prod:
	docker compose up -d producer

up-app:
	docker compose up -d app

up-metric:
	docker compose up -d prometheus grafana

# migrate psql using goose; it waits for the psql to be up
migrate:
	@env $(shell grep -v '^#' $(ENV_FILE) | xargs) bash -c '\
		while ! PGPASSWORD=$$POSTGRES_PASSWORD psql -h $$POSTGRES_HOST -p $$POSTGRES_PORT -U $$POSTGRES_USER -d $$POSTGRES_DB -c "\q" >/dev/null 2>&1; do \
			echo "Waiting for Postgres to be ready..."; \
			sleep 2; \
		done; \
		echo "Postgres is ready - running migrations"; \
		goose -dir sql postgres "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$$POSTGRES_HOST:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" up; \
	'

# Runs everything as containers + migrates the
# database (Goose is needed) + and generates order json payload via python script 
run-all: gen up-kafka up-psql migrate up-prod up-app up-metric


# tearing down and cleanup
clean:
	@read -p "Delete 'bin/', 'mock.json' and ALL containers and their volumes. Continue? (y/N) " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down -v --remove-orphans; \
		rm -f mock.json; \
		rm -rf bin/; \
	else \
		echo "Aborted."; \
	fi

# ----LOCAL BUILDS----
# they are prefixed with 'loc-'

# Build bins
loc-build:
	go build -o bin/service ./cmd/service
	go build -o bin/producer ./producer

# Run service in foreground
loc-runs:
	./bin/service 

# Run producer in background
loc-runp:
	./bin/producer --file orders.json --rps 1

# run all tests no cache
test:
	go test -v ./... -count=1 -cover