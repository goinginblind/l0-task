build:
	go build -o bin/service ./cmd/service
	go build -o bin/producer ./producer

run: build
	./bin/producer
	./bin/service

test:
	go test -v ./... -count=1