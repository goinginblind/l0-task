package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Set environment variables for the test
	t.Setenv("POSTGRES_DSN", "test_dsn")
	t.Setenv("HTTP_SERVER_PORT", "8888")
	t.Setenv("KAFKA_BROKERS", "kafka1:9092,kafka2:9092")
	t.Setenv("KAFKA_TOPIC", "test_topic")
	t.Setenv("KAFKA_GROUP_ID", "test_group")

	cfg, err := Load()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "test_dsn", cfg.PostgresDSN)
	assert.Equal(t, ":8888", cfg.HTTPServerPort)
	assert.Equal(t, []string{"kafka1:9092,kafka2:9092"}, cfg.KafkaBrokers)
	assert.Equal(t, "test_topic", cfg.KafkaTopic)
	assert.Equal(t, "test_group", cfg.KafkaGroupID)
}
