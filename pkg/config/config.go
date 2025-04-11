package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig
	Logger     LoggerConfig
	Postgres   PostgresConfig
	JWT        JWTConfig
	GRPC       GRPCConfig
	Prometheus PrometheusConfig
}

type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	GinMode      string
	AppEnv       string
}

type LoggerConfig struct {
	Level    string
	Format   string // "json" или "console"
	Output   string // "stdout" или "file"
	FilePath string // путь к файлу логов
}

type PostgresConfig struct {
	Host               string
	Port               string
	DB                 string
	User               string
	Password           string
	SSLMode            string
	MaxConnections     int
	IdleConnections    int
	ConnectionLifetime time.Duration
	QueryTimeout       time.Duration
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type GRPCConfig struct {
	Port string
}

type PrometheusConfig struct {
	Port string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found. Using environment variables.\n")
	}

	viper.AutomaticEnv()

	setDefaults()

	config := &Config{
		Server: ServerConfig{
			Host:         viper.GetString("APP_HOST"),
			Port:         viper.GetString("APP_PORT"),
			ReadTimeout:  viper.GetDuration("HTTP_READ_TIMEOUT"),
			WriteTimeout: viper.GetDuration("HTTP_WRITE_TIMEOUT"),
			IdleTimeout:  viper.GetDuration("HTTP_IDLE_TIMEOUT"),
			GinMode:      viper.GetString("GIN_MODE"),
			AppEnv:       viper.GetString("APP_ENV"),
		},
		Logger: LoggerConfig{
			Level:    viper.GetString("LOG_LEVEL"),
			Format:   viper.GetString("LOG_FORMAT"),
			Output:   viper.GetString("LOG_OUTPUT"),
			FilePath: viper.GetString("LOG_FILE_PATH"),
		},
		Postgres: PostgresConfig{
			Host:               viper.GetString("POSTGRES_HOST"),
			Port:               viper.GetString("POSTGRES_PORT"),
			DB:                 viper.GetString("POSTGRES_DB"),
			User:               viper.GetString("POSTGRES_USER"),
			Password:           viper.GetString("POSTGRES_PASSWORD"),
			SSLMode:            viper.GetString("POSTGRES_SSL_MODE"),
			MaxConnections:     viper.GetInt("POSTGRES_MAX_CONNECTIONS"),
			IdleConnections:    viper.GetInt("POSTGRES_IDLE_CONNECTIONS"),
			ConnectionLifetime: viper.GetDuration("POSTGRES_CONNECTION_LIFETIME"),
			QueryTimeout:       viper.GetDuration("DB_QUERY_TIMEOUT"),
		},
		JWT: JWTConfig{
			Secret:     viper.GetString("JWT_SECRET"),
			Expiration: viper.GetDuration("JWT_EXPIRATION"),
		},
		GRPC: GRPCConfig{
			Port: viper.GetString("APP_GRPC_PORT"),
		},
		Prometheus: PrometheusConfig{
			Port: viper.GetString("APP_PROMETHEUS_PORT"),
		},
	}

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

func setDefaults() {
	viper.SetDefault("APP_HOST", "0.0.0.0")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("HTTP_READ_TIMEOUT", 5*time.Second)
	viper.SetDefault("HTTP_WRITE_TIMEOUT", 10*time.Second)
	viper.SetDefault("HTTP_IDLE_TIMEOUT", 120*time.Second)
	viper.SetDefault("GIN_MODE", "release")
	viper.SetDefault("APP_ENV", "development")

	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "json")
	viper.SetDefault("LOG_OUTPUT", "stdout")
	viper.SetDefault("LOG_FILE_PATH", "./logs/app.log")

	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_PORT", "5432")
	viper.SetDefault("POSTGRES_DB", "pvz_db")
	viper.SetDefault("POSTGRES_USER", "postgres")
	viper.SetDefault("POSTGRES_PASSWORD", "postgres")
	viper.SetDefault("POSTGRES_SSL_MODE", "disable")
	viper.SetDefault("POSTGRES_MAX_CONNECTIONS", 20)
	viper.SetDefault("POSTGRES_IDLE_CONNECTIONS", 5)
	viper.SetDefault("POSTGRES_CONNECTION_LIFETIME", 300*time.Second)
	viper.SetDefault("DB_QUERY_TIMEOUT", 5*time.Second)

	viper.SetDefault("JWT_SECRET", "default_secret_key_change_this_in_production")
	viper.SetDefault("JWT_EXPIRATION", 24*time.Hour)

	viper.SetDefault("APP_GRPC_PORT", "3000")

	viper.SetDefault("APP_PROMETHEUS_PORT", "9000")
}

func validateConfig(cfg *Config) error {
	if cfg.JWT.Secret == "default_secret_key_change_this_in_production" && cfg.Server.AppEnv == "production" {
		return fmt.Errorf("JWT_SECRET should be changed in production")
	}

	if cfg.Postgres.DB == "" || cfg.Postgres.User == "" {
		return fmt.Errorf("POSTGRES_DB and POSTGRES_USER are required")
	}

	return nil
}
