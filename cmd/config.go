package cmd

import (
	"fmt"

	"github.com/ethan-huo/iris/internal/config"
	"gopkg.in/yaml.v3"
)

type ConfigCmd struct{}

func (c *ConfigCmd) Run(cfg *config.AppConfig) error {
	path := cfg.ConfigPath()
	if cfg.Config.IsZero() {
		fmt.Printf("Config file: %s (not configured)\n", path)
		return nil
	}

	fmt.Printf("Config file: %s\n\n", path)
	data, err := yaml.Marshal(cfg.Config.Redacted())
	if err != nil {
		return fmt.Errorf("marshal redacted config: %w", err)
	}
	fmt.Print(string(data))
	return nil
}
