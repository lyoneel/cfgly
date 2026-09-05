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
type ConfigBuilder struct {
	cfg                 *AppConfig
	configDirCustomPath string
	dataDirCustomPath   string
	autoCreate          bool
	err                 error
}

// NewAppConfigBuilder creates a new builder for configuring an application instance.
func NewAppConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{cfg: &AppConfig{}}
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
	b.cfg.name = name
	return b
}

// WithConfigDir sets an explicit configuration directory override.
func (b *ConfigBuilder) WithConfigDir(dir string) *ConfigBuilder {
	b.configDirCustomPath = filepath.Join(dir, b.cfg.name)
	return b
}

// WithDataDir sets an explicit data directory override.
func (b *ConfigBuilder) WithDataDir(dir string) *ConfigBuilder {
	b.dataDirCustomPath = filepath.Join(dir, b.cfg.name)
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
func (b *ConfigBuilder) Build() (*AppConfig, error) {
	if b.err != nil {
		return nil, b.err
	}

	// Resolve data path if not explicitly set
	if b.dataDirCustomPath == "" {
		base, err := getDataBaseForOS(runtime.GOOS)
		if err != nil {
			return nil, fmt.Errorf("cannot determine data directory: %w", err)
		}
		b.dataDirCustomPath = filepath.Join(base, b.cfg.name)
	}
	b.cfg.dataDir = NewConfigDir(b.dataDirCustomPath)

	if b.configDirCustomPath == "" {
		base, err := getConfigBaseForOS(runtime.GOOS)
		if err != nil {
			return nil, fmt.Errorf("cannot determine config directory: %w", err)
		}
		b.configDirCustomPath = filepath.Join(base, b.cfg.name)
	}
	b.cfg.configDir = NewConfigDir(b.configDirCustomPath)

	// Create directories if auto-create is enabled
	if b.autoCreate {
		if err := os.MkdirAll(b.dataDirCustomPath, 0o755); err != nil {
			return nil, fmt.Errorf("create data directory: %w", err)
		}
		if err := os.MkdirAll(b.configDirCustomPath, 0o755); err != nil {
			return nil, fmt.Errorf("create config directory: %w", err)
		}
	}

	return b.cfg, nil
}
