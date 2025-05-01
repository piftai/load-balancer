package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Backend struct {
	URL    string `yaml:"url"`
	Weight int    `yaml:"weight"`
}

type Config struct {
	Port       string    `yaml:"port" default:"8080"`
	Backends   []Backend `yaml:"backends"`
	HealthTick int       `yaml:"health_tick"`
}

func Load() Config {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("Failed to read config.yaml. Error:", err)
	}

	var config Config
	if err = yaml.Unmarshal(data, &config); err != nil {
		log.Fatal("Failed to parse config. Error:", err)
	}

	return config
}
