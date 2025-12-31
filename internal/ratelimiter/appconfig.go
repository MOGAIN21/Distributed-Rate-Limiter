package ratelimiter

import (
	"fmt"
	"os"
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
	return &config, nil
}