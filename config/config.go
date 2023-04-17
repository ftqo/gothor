package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger `yaml:"logger"`
	DB     `yaml:"db"`
	Server `yaml:"server"`
}

type Logger struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type DB struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int16  `yaml:"port"`
	Database string `yaml:"database"`
}

type Server struct {
	Port    uint16 `yaml:"port"`
	BaseURL string `yaml:"baseURL"`
}

func Get() (Config, error) {
	c := Config{}
	b, err := os.ReadFile("config.yaml")
	if err != nil {
		return c, fmt.Errorf("failed to read config file: %v", err)
	}
	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return c, fmt.Errorf("failed to unmarshal config file: %v", err)
	}
	return c, nil
}
