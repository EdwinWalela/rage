package config

import (
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
	return cfg, nil
}
