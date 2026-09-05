package cfgly

import (
	"fmt"
	"os"
	"path/filepath"
)

// getConfigBaseForOS returns the base configuration directory for the given OS.
func getConfigBaseForOS(goos string) (string, error) {
	switch goos {
	case "linux":
		return getLinuxConfigDir()
	case "darwin":
		// TODO: make MacOSStrategy configurable at dispatcher level when macOS strategy becomes user-configurable
		return getMacOSConfigDir(MacOSStrategyApple)
	case "windows":
		return getWindowsConfigDir()
	default:
		// Fallback to home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		return filepath.Join(home, ".config"), nil
	}
}

// getDataBaseForOS returns the base data directory for the given OS.
func getDataBaseForOS(goos string) (string, error) {
	switch goos {
	case "linux":
		return getLinuxDataDir()
	case "darwin":
		// macOS: configurable strategy
		return getMacOSDataDir(MacOSStrategyApple)
	case "windows":
		return getWindowsDataDir()
	default:
		// Fallback to home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		return filepath.Join(home, ".local", "share"), nil
	}
}

// getLinuxConfigDir returns the Linux configuration directory following XDG Base Directory Specification.
func getLinuxConfigDir() (string, error) {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return xdgConfig, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	return filepath.Join(home, ".config"), nil
}

// getLinuxDataDir returns the Linux data directory following XDG Base Directory Specification.
func getLinuxDataDir() (string, error) {
	if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
		return xdgData, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	return filepath.Join(home, ".local", "share"), nil
}

// getMacOSConfigDir returns the macOS configuration directory based on strategy.
func getMacOSConfigDir(strategy MacOSStrategy) (string, error) {
	switch strategy {
	case MacOSStrategyApple:
		return getMacOSAppleConfigDir()
	case MacOSStrategyXDG:
		fallthrough
	default:
		return getLinuxConfigDir() // Use XDG style
	}
}

// getMacOSDataDir returns the macOS data directory based on strategy.
func getMacOSDataDir(strategy MacOSStrategy) (string, error) {
	switch strategy {
	case MacOSStrategyApple:
		return getMacOSAppleDataDir()
	case MacOSStrategyXDG:
		fallthrough
	default:
		return getLinuxDataDir() // Use XDG style
	}
}

// getWindowsConfigDir returns the Windows configuration directory.
func getWindowsConfigDir() (string, error) {
	if appData := os.Getenv("APPDATA"); appData != "" {
		return appData, nil
	}

	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		return filepath.Join(userProfile, "AppData", "Roaming"), nil
	}

	return "", fmt.Errorf("cannot determine Windows APPDATA directory")
}

// getWindowsDataDir returns the Windows data directory.
func getWindowsDataDir() (string, error) {
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		return localAppData, nil
	}

	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		return filepath.Join(userProfile, "AppData", "Local"), nil
	}

	return "", fmt.Errorf("cannot determine Windows LOCALAPPDATA directory")
}

// getMacOSAppleConfigDir returns the traditional macOS configuration directory.
func getMacOSAppleConfigDir() (string, error) {
	if envDir := os.Getenv("OSDIR_MACOS_APPLE_CONFIG_HOME"); envDir != "" {
		return envDir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	return filepath.Join(home, "Library", "Preferences"), nil
}

// getMacOSAppleDataDir returns the traditional macOS data directory.
func getMacOSAppleDataDir() (string, error) {
	if envDir := os.Getenv("OSDIR_MACOS_APPLE_DATA_HOME"); envDir != "" {
		return envDir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	return filepath.Join(home, "Library", "Application Support"), nil
}
