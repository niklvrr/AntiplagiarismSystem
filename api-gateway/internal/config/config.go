package config

import "os"

type AppConfig struct {
	HTTPPort string
}

type StoringServiceConfig struct {
	Endpoint string
}

type AnalysisServiceConfig struct {
	Endpoint string
}

type LoggerConfig struct {
	Level string
}

type Config struct {
	App             AppConfig
	StoringService  StoringServiceConfig
	AnalysisService AnalysisServiceConfig
	Logger          LoggerConfig
}

func LoadConfig() *Config {
	return &Config{
		App: AppConfig{
			HTTPPort: getEnv("HTTP_PORT", "8080"),
		},
		StoringService: StoringServiceConfig{
			Endpoint: getEnv("STORING_SERVICE_ENDPOINT", "localhost:50051"),
		},
		AnalysisService: AnalysisServiceConfig{
			Endpoint: getEnv("ANALYSIS_SERVICE_ENDPOINT", "localhost:50052"),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "prod"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

