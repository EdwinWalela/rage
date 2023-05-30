package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type target struct {
	Url    string
	Method string
}

type load struct {
	Users    int
	Attempts int
}

type Config struct {
	Target  target
	Load    load
	Headers map[string]string
	Body    map[string]interface{}
}

func Parse(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{}
	unmarshalErr := yaml.Unmarshal(data, &cfg)
	if unmarshalErr != nil {
		return Config{}, unmarshalErr
	}

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func generateError(val string) error {
	return fmt.Errorf("%v value is required", val)
}

func validate(cfg Config) error {
	if cfg.Target.Method == "" {
		return generateError("target.method")
	}
	if cfg.Target.Url == "" {
		return generateError("URL")
	}
	return nil
}
