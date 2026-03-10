package config

import (
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
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "iris")
}

func Load() (*AppConfig, error) {
	cfgDir := configDir()
	os.MkdirAll(cfgDir, 0755)

	ac := &AppConfig{}

	cfgPath := filepath.Join(cfgDir, "config.yaml")
	if data, err := os.ReadFile(cfgPath); err == nil {
		if err := yaml.Unmarshal(data, &ac.Config); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
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
	return filepath.Join(configDir(), "config.yaml")
}

func Save(cfg Config) error {
	dir := configDir()
	os.MkdirAll(dir, 0755)

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "config.yaml"), data, 0600)
}

func expandEnv(s string) string {
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		name := s[2 : len(s)-1]
		return os.Getenv(name)
	}
	return s
}
