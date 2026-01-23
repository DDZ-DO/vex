package config

import (
	"os"
	"path/filepath"
)

// Config holds the editor configuration.
type Config struct {
	// Editor settings
	TabWidth     int  `toml:"tab_width"`
	InsertSpaces bool `toml:"insert_spaces"`
	WordWrap     bool `toml:"word_wrap"`
	LineNumbers  bool `toml:"line_numbers"`

	// UI settings
	Theme        string `toml:"theme"`
	SidebarWidth int    `toml:"sidebar_width"`
	ShowSidebar  bool   `toml:"show_sidebar"`

	// File settings
	AutoSave               bool `toml:"auto_save"`
	TrimTrailingWhitespace bool `toml:"trim_trailing_whitespace"`
	InsertFinalNewline     bool `toml:"insert_final_newline"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		TabWidth:     4,
		InsertSpaces: true,
		WordWrap:     false,
		LineNumbers:  true,

		Theme:        "default",
		SidebarWidth: 25,
		ShowSidebar:  true,

		AutoSave:               false,
		TrimTrailingWhitespace: false,
		InsertFinalNewline:     true,
	}
}

// ConfigDir returns the configuration directory path.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Use XDG_CONFIG_HOME if set, otherwise ~/.config
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(home, ".config")
	}

	return filepath.Join(configHome, "vex"), nil
}

// ConfigPath returns the path to the config file.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// Load loads the configuration from the config file.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	configPath, err := ConfigPath()
	if err != nil {
		return cfg, nil // Return default on error
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return cfg, nil // Return default if file doesn't exist
	}

	// For now, just return default config
	// TODO: Implement TOML parsing when needed
	return cfg, nil
}

// Save saves the configuration to the config file.
func (c *Config) Save() error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// TODO: Implement TOML serialization when needed
	return nil
}

// EnsureConfigDir creates the config directory if it doesn't exist.
func EnsureConfigDir() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}
