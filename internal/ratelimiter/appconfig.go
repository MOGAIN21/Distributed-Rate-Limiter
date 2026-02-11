package ratelimiter

import (
	"fmt"
	"os"
	"strconv"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Server struct {
		Port int `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	RateLimiter struct {
		DefaultCapacity int32 `yaml:"default_capacity"`
		DefaultRefillRate float64 `yaml:"default_refill_rate"`
	} `yaml:"ratelimiter"`
	Redis struct {
		Enabled bool `yaml:"enabled"`
		Host string `yaml:"host"`
		Port int `yaml:"port"`
	} `yaml:"redis"`
}

func LoadConfig(filename string) (*AppConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AppConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	//Override with environment variables if set
	if host := os.Getenv("REDIS_HOST"); host != "" {
		config.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		if portInt, err := strconv.Atoi(port); err == nil {
			config.Redis.Port = portInt
		}
	}
	if enabled := os.Getenv("REDIS_ENABLED"); enabled != "" {
		config.Redis.Enabled = enabled == "true"
	}
	return &config, nil
}