// Package cfgly provides cross-platform file management with OS-specific directory resolution.
//
// It supports:
//   - Linux: XDG Base Directory Specification (with fallbacks)
//   - macOS: Configurable between XDG-style and traditional Apple directories
//   - Windows: APPDATA and LOCALAPPDATA environment variables
//
// This is an instance-based API with no global state. Create instances using
// NewAppConfigBuilder with optional configuration via builder methods.
package cfgly

// AppConfig represents an application configuration with OS-specific configuration and data directories.
type AppConfig struct {
	name      string
	dataDir   *ConfigDir
	configDir *ConfigDir
}

// Name returns the application name.
func (cfg *AppConfig) Name() string {
	return cfg.name
}

// Config returns the configuration directory instance.
func (cfg *AppConfig) Config() *ConfigDir {
	return cfg.configDir
}

// Data returns the data directory instance.
func (cfg *AppConfig) Data() *ConfigDir {
	return cfg.dataDir
}

// ConfigDir returns the configuration directory for this application.
//
// The path is resolved at build time and never changes:
// On Linux: $XDG_CONFIG_HOME/appName (default: ~/.config/appName)
// On macOS: ~/Library/Preferences/appName (Apple strategy) or ~/.config/appName (XDG strategy)
// On Windows: %APPDATA%\appName
func (cfg *AppConfig) ConfigDir() string {
	return cfg.configDir.BasePath()
}

// DataDir returns the data directory for this application.
//
// The path is resolved at build time and never changes:
// - If set via WithDataDir, uses that path + app name
// - Otherwise, resolves to OS-specific default (e.g., ~/.local/share/appName on Linux)
//
// On Linux: $XDG_DATA_HOME/appName (default: ~/.local/share/appName)
// On macOS: ~/Library/Application Support/appName (Apple strategy) or ~/.local/share/appName (XDG strategy)
// On Windows: %LOCALAPPDATA%\appName
func (cfg *AppConfig) DataDir() string {
	return cfg.dataDir.BasePath()
}
