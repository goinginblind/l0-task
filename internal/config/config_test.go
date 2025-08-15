package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Set environment variables for the test
	t.Setenv("DATABASE_HOST", "test_host")
	t.Setenv("DATABASE_PORT", "1234")
	t.Setenv("DATABASE_USER", "no_usr")
	t.Setenv("DATABASE_PASSWORD", "qwerty")
	t.Setenv("DATABASE_DBNAME", "postgre")
	t.Setenv("DATABASE_SSLMODE", "disable")
	t.Setenv("DATABASE_MAX_CONNECTIONS", "10")
	t.Setenv("DATABASE_MAX_IDLE_CONNS", "5")

	cfg, err := LoadConfig()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "test_host", cfg.Database.Host)
	assert.Equal(t, "1234", cfg.Database.Port)
	assert.Equal(t, "no_usr", cfg.Database.User)
	assert.Equal(t, "qwerty", cfg.Database.Password)
	assert.Equal(t, "postgre", cfg.Database.DBName)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, 10, cfg.Database.MaxConnections)
}
