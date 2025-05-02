package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Backend struct {
	URL    string `yaml:"url"`
	Weight int    `yaml:"weight"`
}

type RateLimitingConfig struct {
	DefaultCapacity int `yaml:"default_capacity"`
	DefaultRate     int `yaml:"default_rate"` // запросов в секунду
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

func (p *PostgresConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode)
}

type DatabaseConfig struct {
	Postgres PostgresConfig `yaml:"postgres"`
}

type Config struct {
	Port         string             `yaml:"port" default:"8080"`
	Backends     []Backend          `yaml:"backends"`
	HealthTick   int                `yaml:"health_tick"`
	RateLimiting RateLimitingConfig `yaml:"rate_limiting"`
	Database     DatabaseConfig     `yaml:"database"`
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
