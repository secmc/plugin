package plugin

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// ConfigFile is the default configuration file used for plugin definitions.
const ConfigFile = "plugins/plugins.yaml"

// Config holds the global plugin configuration file structure.
type Config struct {
	Plugins []PluginConfig `yaml:"plugins"`
}

// PluginConfig holds process configuration for a single plugin instance.
type PluginConfig struct {
	ID      string            `yaml:"id"`
	Name    string            `yaml:"name"`
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args"`
	WorkDir string            `yaml:"work_dir"`
	Env     map[string]string `yaml:"env"`
	Address string            `yaml:"address"`
}

// LoadConfig reads and decodes the plugin configuration file. If the file does not
// exist, os.ErrNotExist is returned.
func LoadConfig(path string) (Config, error) {
	if path == "" {
		path = ConfigFile
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}
	if err != nil {
		return Config{}, fmt.Errorf("read plugin config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode plugin config: %w", err)
	}
	for i := range cfg.Plugins {
		if cfg.Plugins[i].ID == "" {
			cfg.Plugins[i].ID = fmt.Sprintf("plugin-%d", i+1)
		}
		if cfg.Plugins[i].Address == "" {
			cfg.Plugins[i].Address = "127.0.0.1:0"
		}
		if cfg.Plugins[i].Command != "" && cfg.Plugins[i].WorkDir != "" {
			if !filepath.IsAbs(cfg.Plugins[i].WorkDir) {
				cfg.Plugins[i].WorkDir = filepath.Clean(cfg.Plugins[i].WorkDir)
			}
		}
	}
	return cfg, nil
}
