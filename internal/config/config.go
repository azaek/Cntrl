package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AppName is the application name used for paths and service registration
// This can be changed if the project is renamed
const AppName = "Cntrl"

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Display  DisplayConfig  `yaml:"display"`
	Features FeaturesConfig `yaml:"features"`
	Stats    StatsConfig    `yaml:"stats"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type DisplayConfig struct {
	Hostname string `yaml:"hostname"`
}

type FeaturesConfig struct {
	EnableShutdown  bool `yaml:"enable_shutdown"`
	EnableRestart   bool `yaml:"enable_restart"`
	EnableHibernate bool `yaml:"enable_hibernate"`
	EnableStats     bool `yaml:"enable_stats"`
}

type StatsConfig struct {
	GpuEnabled       bool `yaml:"gpu_enabled"`
	DiskCacheSeconds int  `yaml:"disk_cache_seconds"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 9990,
			Host: "0.0.0.0",
		},
		Display: DisplayConfig{
			Hostname: "",
		},
		Features: FeaturesConfig{
			EnableShutdown:  true,
			EnableRestart:   true,
			EnableHibernate: true,
			EnableStats:     true,
		},
		Stats: StatsConfig{
			GpuEnabled:       true,
			DiskCacheSeconds: 30,
		},
	}
}

// GetConfigPath returns the path to the config file in APPDATA
func GetConfigPath() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		// Fallback to user home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		appData = filepath.Join(home, "AppData", "Roaming")
	}

	configDir := filepath.Join(appData, AppName)
	return filepath.Join(configDir, "config.yaml"), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	return os.MkdirAll(configDir, 0755)
}

// Load loads the configuration from the config file
// If the file doesn't exist, it returns the default configuration
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save saves the configuration to the config file
func Save(cfg *Config) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// CreateDefaultConfig creates the default config file if it doesn't exist
func CreateDefaultConfig() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // Config exists, don't overwrite
	}

	return Save(DefaultConfig())
}
