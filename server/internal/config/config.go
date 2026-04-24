package config

import (
	"os"
	"time"
)

type Config struct {
	Addr           string
	LogLevel       string
	UIOrigin       string
	DataDir        string
	StorageDriver  string
	DatabaseURL    string
	MigrationsDir  string
	ObjectDir      string
	WorkerEnabled  bool
	WorkerInterval time.Duration
}

func LoadFromEnv() Config {
	return Config{
		Addr:           envOrDefault("GUILD_ADDR", ":8080"),
		LogLevel:       envOrDefault("GUILD_LOG_LEVEL", "debug"),
		UIOrigin:       envOrDefault("GUILD_UI_ORIGIN", "http://localhost:3000"),
		DataDir:        envOrDefault("GUILD_DATA_DIR", "./data"),
		StorageDriver:  envOrDefault("GUILD_STORAGE_DRIVER", "file"),
		DatabaseURL:    os.Getenv("GUILD_DATABASE_URL"),
		MigrationsDir:  envOrDefault("GUILD_MIGRATIONS_DIR", "server/migrations"),
		ObjectDir:      envOrDefault("GUILD_OBJECT_DIR", "./data/objects"),
		WorkerEnabled:  envOrDefault("GUILD_WORKER_ENABLED", "true") != "false",
		WorkerInterval: durationOrDefault("GUILD_WORKER_INTERVAL", time.Second),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func durationOrDefault(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
