package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type PaddleConfig struct {
	APIKey string `yaml:"api_key"`
}

type Config struct {
	Paddle PaddleConfig `yaml:"paddle"`
}

type AppConfig struct {
	Config Config
	path   string
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}
	return filepath.Join(dir, "iris", "config.yaml"), nil
}

func Load() (*AppConfig, error) {
	cfgPath, err := configPath()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0700); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}

	ac := &AppConfig{path: cfgPath}

	if data, err := os.ReadFile(cfgPath); err == nil {
		if err := yaml.Unmarshal(data, &ac.Config); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read config: %w", err)
	}

	ac.Config.Paddle.APIKey = expandEnv(ac.Config.Paddle.APIKey)

	return ac, nil
}

func (ac *AppConfig) RequireAPIKey() (string, error) {
	key := ac.Config.Paddle.APIKey
	if key == "" {
		return "", fmt.Errorf("Paddle API key not configured\nRun: iris auth login")
	}
	return key, nil
}

func (ac *AppConfig) ConfigPath() string {
	return ac.path
}

func (ac *AppConfig) Save() error {
	if err := os.MkdirAll(filepath.Dir(ac.path), 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(&ac.Config)
	if err != nil {
		return err
	}

	return os.WriteFile(ac.path, data, 0600)
}

func expandEnv(s string) string {
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		name := s[2 : len(s)-1]
		return os.Getenv(name)
	}
	return s
}

func (c Config) IsZero() bool {
	return c.Paddle.APIKey == ""
}

func (c Config) Redacted() Config {
	clone := c
	clone.Paddle.APIKey = MaskSecret(clone.Paddle.APIKey)
	return clone
}

func MaskSecret(secret string) string {
	if secret == "" {
		return ""
	}
	if len(secret) <= 8 {
		return strings.Repeat("*", len(secret))
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
}
