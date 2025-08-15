package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application, loaded with Viper.
type Config struct {
	HTTPServer HTTPServerConfig `mapstructure:"http_server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Kafka      KafkaConfig      `mapstructure:"kafka"`
	Health     HealthConfig     `mapstructure:"health"`
	Consumer   ConsumerConfig   `mapstructure:"consumer"`
}

// HTTPServerConfig holds HTTP server-specific settings (port)
type HTTPServerConfig struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout_s"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout_s"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout_s"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// KafkaConfig holds Kafka-specific settings
type KafkaConfig struct {
	BootstrapServers    string `mapstructure:"bootstrap_servers"`
	ConsumerGroupID     string `mapstructure:"consumer_group_id"`
	AutoOffsetReset     string `mapstructure:"auto_offset_reset"`
	EnableAutoCommit    bool   `mapstructure:"enable_auto_commit"`
	IsolationLevel      string `mapstructure:"isolation_level"`
	MaxPollIntervalMs   int    `mapstructure:"poll_interval_ms"`
	MinFetchSizeBytes   int    `mapstructure:"fetch_min_bytes"`
	MaxFetchSizeBytes   int    `mapstructure:"fetch_max_bytes"`
	SessionTimeoutMs    int    `mapstructure:"session_timeout_ms"`
	HeartbeatIntervalMs int    `mapstructure:"heartbeat_interval_ms"`
}

// DatabaseConfig holds database-specific settings.
type DatabaseConfig struct {
	User                 string `mapstructure:"user"`
	Password             string `mapstructure:"password"`
	Host                 string `mapstructure:"host"`
	Port                 string `mapstructure:"port"`
	DBName               string `mapstructure:"dbname"`
	SSLMode              string `mapstructure:"sslmode"`
	MaxConnections       int    `mapstructure:"max_connections"`
	MaxIdlingConnections int    `mapstructure:"max_idle_conns"`
}

// DSN returns a concatenated dsn
// string consisting of the configs db user, pass, host, etc.
func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
}

// HealthConfig holds health check settings.
type HealthConfig struct {
	DBCheckInterval time.Duration `mapstructure:"db_hp_interval"`
	DBCheckTimeout  time.Duration `mapstructure:"db_hp_timeout"`
}

// ConsumerConfig holds consumer-specific settings.
type ConsumerConfig struct {
	Topic         string        `mapstructure:"topic"`
	WorkerCount   int           `mapstructure:"worker_count"`
	JobBufferSize int           `mapstructure:"job_buffer_size"`
	MaxRetries    int           `mapstructure:"max_retries"`
	RetryBackoff  time.Duration `mapstructure:"retry_backoff"`
}

// LoadConfig reads configuration from file and environment variables:
//   - first it loads defaults
//   - reads a .yaml file if there's one, overwrites the above
//   - looks for variables in the enviroment, overwrites all of the above
func LoadConfig() (*Config, error) {
	// THE DEFAULTS ARE SET HERE (they're somewhat reasonable(i'd like to think so))
	// http server
	viper.SetDefault("http_server.port", "8080")
	viper.SetDefault("http_server.read_timeout", "5s")
	viper.SetDefault("http_server.write_timeout", "10s")
	viper.SetDefault("http_server.idle_timeout", "120s")
	viper.SetDefault("http_server.shutdown_timeout", "30s")

	// db
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "orders_db")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_connections", 10)
	viper.SetDefault("database.max_idle_conns", 5)

	// Kafka
	viper.SetDefault("kafka.bootstrap_servers", "localhost:9092")
	viper.SetDefault("kafka.consumer_group_id", "orders-consumer")
	viper.SetDefault("kafka.auto_offset_reset", "earliest")
	viper.SetDefault("kafka.enable_auto_commit", false)
	viper.SetDefault("kafka.isolation_level", "read_committed")
	viper.SetDefault("kafka.poll_interval_ms", 300000)
	viper.SetDefault("kafka.fetch_min_bytes", 1)
	viper.SetDefault("kafka.fetch_max_bytes", 1048576)
	viper.SetDefault("kafka.session_timeout_ms", 10000)
	viper.SetDefault("kafka.heartbeat_interval_ms", 3000)

	// consumer
	viper.SetDefault("consumer.topic", "orders")
	viper.SetDefault("consumer.worker_count", 4)
	viper.SetDefault("consumer.job_buffer_size", 8)
	viper.SetDefault("consumer.max_retries", 3)
	viper.SetDefault("consumer.retry_backoff", "250ms")

	// health
	viper.SetDefault("health.db_hp_interval", "5s")
	viper.SetDefault("health.db_hp_timeout", "180s")

	// Configure Viper
	viper.SetConfigName("config")    // name of config file (without extension)
	viper.SetConfigType("yaml")      // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")         // look for config in the working directory
	viper.AddConfigPath("./configs") // optionally look in a configs directory
	viper.AddConfigPath("/etc/app/") // and other optional paths like thsi

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Attempt to read the config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore err if we have env vars
		} else {
			// Config file was found but another err
			return nil, fmt.Errorf("Failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
