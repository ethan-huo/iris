package cmd

import (
	"fmt"
	"os"

	"github.com/anthropics/iris/internal/config"
)

type ConfigCmd struct{}

func (c *ConfigCmd) Run(cfg *config.AppConfig) error {
	path := cfg.ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Config file: %s (not found)\n", path)
		return nil
	}

	fmt.Printf("Config file: %s\n\n", path)
	fmt.Print(string(data))
	return nil
}
