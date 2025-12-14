package config

import (
	"errors"
	"fmt"
	"os"
)

var (
	dbUserEmptyError = errors.New("DB User is Empty")
	dbNameEmptyError = errors.New("DB Name is Empty")
)

type AppConfig struct {
	GrpcPort string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	Password string
	User     string
	URL      string
}

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

type LoggerConfig struct {
	Level string
}

type AnalysisConfig struct {
	URL string
}

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Minio    MinioConfig
	Logger   LoggerConfig
	Analysis AnalysisConfig
}

func LoadConfig() (*Config, error) {
	c := &Config{
		App: AppConfig{
			GrpcPort: getEnv("GRPC_PORT", "50051"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			User:     getEnv("DB_USER", "postgres"),
		},
		Minio: MinioConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "user"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "password"),
			Bucket:    getEnv("MINIO_BUCKET", "tasks"),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "prod"),
		},
		Analysis: AnalysisConfig{
			URL: getEnv("ANALYSIS_URL", "localhost:9000"),
		},
	}

	err := makeDbUrl(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func makeDbUrl(cfg *Config) error {
	if cfg.Database.URL == "" {
		if cfg.Database.User == "" {
			return dbUserEmptyError
		}
		if cfg.Database.Name == "" {
			return dbNameEmptyError
		}
		cfg.Database.URL = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Name,
		)
	}
	return nil
}
