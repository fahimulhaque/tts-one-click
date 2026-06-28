package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Model      string `yaml:"model"`
	ServerPort int    `yaml:"server_port"`
	PythonPort int    `yaml:"python_port"`
	VenvPath   string `yaml:"venv_path"`
	DevMode    bool   `yaml:"dev_mode"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		ServerPort: 8080,
		PythonPort: 8001,
		VenvPath:   ".venv",
	}
	data, err := os.ReadFile(path)
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return cfg, err
		}
	}
	// Allow MODEL env var to override the config file value.
	if m := os.Getenv("MODEL"); m != "" {
		cfg.Model = m
	}
	return cfg, nil
}
