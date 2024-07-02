package config

import (
	"os"
	"sync"
)

type Config struct {
	OpenAiModel string
}

var (
	cfg  *Config
	once sync.Once
)

func initConfig() {
	cfg = &Config{
		OpenAiModel: getEnv("OPEN_API_KEY", "gpt-3.5-turbo"),
	}

	validateConfig(cfg)
}

func GetConfig() *Config {
	once.Do(initConfig)
	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func validateConfig(cfg *Config) {
	// Currently no validations are needed
}
