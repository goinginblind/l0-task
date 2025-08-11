package config

import (
	"os"
	"strconv"
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
	portStr := os.Getenv("HTTP_SERVER_PORT")
	if portStr == "" {
		portStr = "8080"
	}
	_, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	return &Config{
		PostgresDSN:    os.Getenv("POSTGRES_DSN"),
		HTTPServerPort: ":" + portStr,
		KafkaBrokers:   []string{os.Getenv("KAFKA_BROKERS")},
		KafkaTopic:     os.Getenv("KAFKA_TOPIC"),
		KafkaGroupID:   os.Getenv("KAFKA_GROUP_ID"),
	}, nil
}
