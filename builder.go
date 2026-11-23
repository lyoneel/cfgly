package cfgly

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// MacOSStrategy defines how macOS directories should be resolved.
type MacOSStrategy string

const (
	// MacOSStrategyXDG uses XDG-style directories (~/.config, ~/.local/share) on macOS.
	// This provides consistency with Linux while being non-traditional for macOS.
	MacOSStrategyXDG MacOSStrategy = "xdg"

	// MacOSStrategyApple uses traditional Apple directories on macOS.
	// Config: ~/Library/Preferences
	// Data: ~/Library/Application Support
	MacOSStrategyApple MacOSStrategy = "apple"
)

// ConfigBuilder provides a fluent API for constructing AppConfig instances.
//
// Every call to Build returns a new, independent *AppConfig: changing or
// reusing the builder afterwards never affects configs that were already
// built. A ConfigBuilder itself is not safe for concurrent use; configure and
// build it from a single goroutine (the resulting *AppConfig values can be
// shared freely).
type ConfigBuilder struct {
	name                string
	configDirCustomPath string
	dataDirCustomPath   string
	autoCreate          bool
	err                 error
}

// NewAppConfigBuilder creates a new builder for configuring an application instance.
func NewAppConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{}
}

// WithName sets the application name (required).
func (b *ConfigBuilder) WithName(name string) *ConfigBuilder {
	if b.err != nil {
		return b
	}
	if name == "" {
		b.err = fmt.Errorf("appName cannot be empty")
		return b
	}
	b.name = name
	return b
}

// WithConfigDir sets an explicit configuration directory override.
func (b *ConfigBuilder) WithConfigDir(dir string) *ConfigBuilder {
	b.configDirCustomPath = filepath.Join(dir, b.name)
	return b
}

// WithDataDir sets an explicit data directory override.
func (b *ConfigBuilder) WithDataDir(dir string) *ConfigBuilder {
	b.dataDirCustomPath = filepath.Join(dir, b.name)
	return b
}

// WithAutoCreate sets whether config and data directories should be
// created automatically during Build. Defaults to false.
func (b *ConfigBuilder) WithAutoCreate(create bool) *ConfigBuilder {
	b.autoCreate = create
	return b
}

// Build resolves OS-specific directories and creates the final instance.
// Returns an error if directory resolution fails (e.g., home directory inaccessible).
// If WithAutoCreate was set, config and data directories are created before returning.
//
// Build does not modify the builder: OS-specific paths are resolved on every
// call and each call returns a new *AppConfig that shares no mutable state
// with the builder or with previously built configs.
func (b *ConfigBuilder) Build() (*AppConfig, error) {
	if b.err != nil {
		return nil, b.err
	}

	// Resolve data path if not explicitly set
	dataPath := b.dataDirCustomPath
	if dataPath == "" {
		base, err := getDataBaseForOS(runtime.GOOS)
		if err != nil {
			return nil, fmt.Errorf("cannot determine data directory: %w", err)
		}
		dataPath = filepath.Join(base, b.name)
	}

	configPath := b.configDirCustomPath
	if configPath == "" {
		base, err := getConfigBaseForOS(runtime.GOOS)
		if err != nil {
			return nil, fmt.Errorf("cannot determine config directory: %w", err)
		}
		configPath = filepath.Join(base, b.name)
	}

	// Create directories if auto-create is enabled
	if b.autoCreate {
		if err := os.MkdirAll(dataPath, 0o755); err != nil {
			return nil, fmt.Errorf("create data directory: %w", err)
		}
		if err := os.MkdirAll(configPath, 0o755); err != nil {
			return nil, fmt.Errorf("create config directory: %w", err)
		}
	}

	return &AppConfig{
		name:      b.name,
		dataDir:   NewConfigDir(dataPath),
		configDir: NewConfigDir(configPath),
	}, nil
}
