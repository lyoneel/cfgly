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
	name               string
	configBaseOverride string // absolute base set by WithConfigDir; the name is joined in Build
	dataBaseOverride   string // absolute base set by WithDataDir; the name is joined in Build
	configExtensions   []string
	autoCreate         bool
	err                error
}

// NewAppConfigBuilder creates a new builder for configuring an application instance.
func NewAppConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{}
}

// WithName sets the application name (required).
//
// The name is appended to the base directories, so it must be a single plain
// path segment: empty, ".", "..", names containing '/', '\', ':', NUL or
// other control characters, and Windows reserved device names (CON, NUL,
// COM1, LPT1, ...) are rejected. Build then returns an error wrapping
// ErrInvalidAppName. Names like "myapp", "my-app" and "com.example.app" are
// valid.
func (b *ConfigBuilder) WithName(name string) *ConfigBuilder {
	if b.err != nil {
		return b
	}
	if err := validateAppName(name); err != nil {
		b.err = err
		return b
	}
	b.name = name
	return b
}

// WithConfigDir overrides the base directory for configuration. Build
// appends the app name, so the config directory is dir/<name> regardless of
// whether WithName is called before or after WithConfigDir. dir must be a
// non-empty absolute path; otherwise Build returns an error wrapping
// ErrInvalidBaseDir.
func (b *ConfigBuilder) WithConfigDir(dir string) *ConfigBuilder {
	if b.err != nil {
		return b
	}
	if err := validateBaseDir("config", dir); err != nil {
		b.err = err
		return b
	}
	b.configBaseOverride = dir
	return b
}

// WithDataDir overrides the base directory for data. Build appends the app
// name, so the data directory is dir/<name> regardless of call order. dir
// must be a non-empty absolute path; otherwise Build returns an error
// wrapping ErrInvalidBaseDir.
func (b *ConfigBuilder) WithDataDir(dir string) *ConfigBuilder {
	if b.err != nil {
		return b
	}
	if err := validateBaseDir("data", dir); err != nil {
		b.err = err
		return b
	}
	b.dataBaseOverride = dir
	return b
}

func validateBaseDir(kind, dir string) error {
	if dir == "" {
		return fmt.Errorf("%w: %s base directory cannot be empty", ErrInvalidBaseDir, kind)
	}
	if !filepath.IsAbs(dir) {
		return fmt.Errorf("%w: %s base directory %q must be an absolute path", ErrInvalidBaseDir, kind, dir)
	}
	return nil
}

// WithConfigExtensions restricts the configuration directory to files with
// the given extensions (case-insensitive, with or without a leading dot, e.g.
// "yaml", ".json"). Reads and writes of other files through Config() fail
// with ErrExtensionNotAllowed, Exists reports false for them, and listing and
// glob search skip them. Each call replaces the previous set; calling it with
// no arguments removes the filter. The data directory is never filtered.
func (b *ConfigBuilder) WithConfigExtensions(extensions ...string) *ConfigBuilder {
	b.configExtensions = normalizeExtensions(extensions)
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
	if err := validateAppName(b.name); err != nil {
		return nil, err // WithName was never called
	}

	// Resolve the data base if not explicitly set, then append the name.
	dataBase := b.dataBaseOverride
	if dataBase == "" {
		base, err := getDataBaseForOS(runtime.GOOS)
		if err != nil {
			return nil, fmt.Errorf("cannot determine data directory: %w", err)
		}
		dataBase = base
	}
	dataPath := filepath.Join(dataBase, b.name)

	configBase := b.configBaseOverride
	if configBase == "" {
		base, err := getConfigBaseForOS(runtime.GOOS)
		if err != nil {
			return nil, fmt.Errorf("cannot determine config directory: %w", err)
		}
		configBase = base
	}
	configPath := filepath.Join(configBase, b.name)

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
		configDir: NewConfigDir(configPath, b.configExtensions...),
	}, nil
}
