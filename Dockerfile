# syntax=docker/dockerfile:1
# Common Builder Stage
FROM golang:1.24.5-bookworm AS builder
WORKDIR /app

# Install CGO dependencies for confluent-kafka-go using apt-get
RUN apt-get update && apt-get install -y build-essential librdkafka-dev

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build BOTH binaries
RUN go build -o /app/service ./cmd/service
RUN go build -o /app/producer ./cmd/producer




# Final Stage for the 'service' binary
FROM debian:bookworm-slim AS service
WORKDIR /app

# Install the runtime dependency for confluent-kafka-go
RUN apt-get update && apt-get install -y --no-install-recommends librdkafka1

COPY --from=builder /app/service .
COPY config.yaml.example ./config.yaml

EXPOSE 8080
CMD ["./service"]



# Final Stage for the 'producer' binary
FROM debian:bookworm-slim AS producer
WORKDIR /app

# Install the runtime dependency for confluent-kafka-go
RUN apt-get update && apt-get install -y --no-install-recommends librdkafka1

COPY --from=builder /app/producer .
COPY mock.json .

CMD ["./producer"]
