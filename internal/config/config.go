package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the application.
type Config struct {
	PostgresDSN    string
	HTTPServerPort string
	KafkaBrokers   []string
	KafkaTopic     string
	KafkaGroupID   string
}

// Load populates the config from environment variables.
// It does not have default values. Yeah I know, pretty sad.
func Load() (*Config, error) {
	required := []string{"POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_DB"}
	for _, key := range required {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			return nil, fmt.Errorf("missing required env var %s", key)
		}
	}

	portStr := os.Getenv("HTTP_SERVER_PORT")
	if portStr == "" {
		portStr = "8080"
	}
	_, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		strings.TrimSpace(os.Getenv("POSTGRES_USER")),
		strings.TrimSpace(os.Getenv("POSTGRES_PASSWORD")),
		strings.TrimSpace(os.Getenv("POSTGRES_HOST")),
		strings.TrimSpace(os.Getenv("POSTGRES_PORT")),
		strings.TrimSpace(os.Getenv("POSTGRES_DB")),
	)

	return &Config{
		PostgresDSN:    dsn,
		HTTPServerPort: ":" + portStr,
		KafkaBrokers:   []string{os.Getenv("KAFKA_BROKERS")},
		KafkaTopic:     os.Getenv("KAFKA_TOPIC"),
		KafkaGroupID:   os.Getenv("KAFKA_GROUP_ID"),
	}, nil
}
