build:
	go build -o bin/service

run: build
	./bin/service

test:
	go test -v ./... -count=1