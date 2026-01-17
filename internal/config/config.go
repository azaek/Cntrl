package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AppName is the application name used for paths and service registration
// This can be changed if the project is renamed
const AppName = "Cntrl"

// AppURL is the project's GitHub repository
const AppURL = "https://github.com/azaek/cntrl"

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
	EnableSleep     bool `yaml:"enable_sleep"`
	EnableSystem    bool `yaml:"enable_system"` // Static system info
	EnableUsage     bool `yaml:"enable_usage"`  // Dynamic usage data
	EnableStats     bool `yaml:"enable_stats"`  // Legacy combined endpoint
	EnableMedia     bool `yaml:"enable_media"`
	EnableProcesses bool `yaml:"enable_processes"`
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
			EnableShutdown:  false, // Disabled by default - Critical Action, allows remote shutdown!
			EnableRestart:   false, // Disabled by default - Critical Action, allows remote restart!
			EnableHibernate: true,
			EnableSleep:     true,
			EnableSystem:    true,
			EnableUsage:     true,
			EnableStats:     true, // Legacy endpoint
			EnableMedia:     true,
			EnableProcesses: true,
		},
		Stats: StatsConfig{
			GpuEnabled:       true,
			DiskCacheSeconds: 30,
		},
	}
}

// GetConfigPath returns the path to the config file in the user's config directory
func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appConfigDir := filepath.Join(configDir, AppName)
	return filepath.Join(appConfigDir, "config.yaml"), nil
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
