package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Set environment variables for the test
	t.Setenv("HTTP_SERVER_PORT", "8888")
	t.Setenv("KAFKA_BROKERS", "kafka1:9092,kafka2:9092")
	t.Setenv("KAFKA_TOPIC", "test_topic")
	t.Setenv("KAFKA_GROUP_ID", "test_group")

	t.Setenv("POSTGRES_USER", "test_user")
	t.Setenv("POSTGRES_PASSWORD", "test_pass")
	t.Setenv("POSTGRES_HOST", "test_host")
	t.Setenv("POSTGRES_PORT", "test_port")
	t.Setenv("POSTGRES_DB", "test_db")

	cfg, err := Load()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "postgres://test_user:test_pass@test_host:test_port/test_db?sslmode=disable", cfg.PostgresDSN)
	assert.Equal(t, ":8888", cfg.HTTPServerPort)
	assert.Equal(t, []string{"kafka1:9092,kafka2:9092"}, cfg.KafkaBrokers)
	assert.Equal(t, "test_topic", cfg.KafkaTopic)
	assert.Equal(t, "test_group", cfg.KafkaGroupID)
}
