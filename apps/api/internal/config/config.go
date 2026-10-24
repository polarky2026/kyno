package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnv string
	Port   int
}

func Load() (Config, error) {
	return parse(os.Getenv)
}

func parse(getenv func(string) string) (Config, error) {
	cfg := Config{AppEnv: "development", Port: 3000}
	var errs []error

	if env := getenv("APP_ENV"); env != "" {
		if env != "development" && env != "production" {
			errs = append(errs, fmt.Errorf("APP_ENV must be development or production, got %q", env))
		}
		cfg.AppEnv = env
	}

	if raw := getenv("PORT"); raw != "" {
		port, err := strconv.Atoi(raw)
		if err != nil || port <= 0 || port > 65535 {
			errs = append(errs, fmt.Errorf("PORT must be a valid port number, got %q", raw))
		}
		cfg.Port = port
	}

	if len(errs) > 0 {
		return Config{}, fmt.Errorf("invalid environment variables: %w", errors.Join(errs...))
	}
	return cfg, nil
}
