package cfgly

import (
	"fmt"
	"os"
	"path/filepath"
)

// envAbsPath returns the value of the environment variable key when it is a
// non-empty absolute path for the target OS family ("windows" or any other
// value for Unix-like systems). Empty or relative values are treated as unset,
// as the XDG Base Directory Specification requires: "If an implementation
// encounters a relative path in any of these variables it should consider the
// path invalid and ignore it."
func envAbsPath(key, goos string) (string, bool) {
	v := os.Getenv(key)
	if v == "" || !isAbsForOS(v, goos) {
		return "", false
	}
	return v, true
}

// isAbsForOS reports whether p is an absolute path under the path rules of
// goos. The check is explicit rather than delegating to filepath.IsAbs so the
// result does not depend on the OS the code runs on (the Windows resolvers
// are also exercised by tests on Linux).
//
// Windows: drive-letter paths with a separator ("C:\x", "C:/x") and UNC or
// device paths ("\\server\share", "\\?\C:\x"). Drive-relative ("C:x")
// and rooted-without-drive (`\temp`) paths are not absolute.
// Other systems: paths starting with "/".
func isAbsForOS(p, goos string) bool {
	if goos != "windows" {
		return len(p) > 0 && p[0] == '/'
	}
	isSep := func(c byte) bool { return c == '\\' || c == '/' }
	if len(p) >= 3 && isDriveLetter(p[0]) && p[1] == ':' && isSep(p[2]) {
		return true
	}
	return len(p) >= 3 && isSep(p[0]) && isSep(p[1]) && !isSep(p[2])
}

func isDriveLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// getConfigBaseForOS returns the base configuration directory for the given
// OS. strategy selects the macOS layout and is ignored on other systems.
func getConfigBaseForOS(goos string, strategy MacOSStrategy) (string, error) {
	switch goos {
	case "linux":
		return getLinuxConfigDir()
	case "darwin":
		return getMacOSConfigDir(strategy)
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
// strategy selects the macOS layout and is ignored on other systems.
func getDataBaseForOS(goos string, strategy MacOSStrategy) (string, error) {
	switch goos {
	case "linux":
		return getLinuxDataDir()
	case "darwin":
		return getMacOSDataDir(strategy)
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
	if xdgConfig, ok := envAbsPath("XDG_CONFIG_HOME", "linux"); ok {
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
	if xdgData, ok := envAbsPath("XDG_DATA_HOME", "linux"); ok {
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
	if appData, ok := envAbsPath("APPDATA", "windows"); ok {
		return appData, nil
	}

	if userProfile, ok := envAbsPath("USERPROFILE", "windows"); ok {
		return filepath.Join(userProfile, "AppData", "Roaming"), nil
	}

	return "", fmt.Errorf("cannot determine Windows APPDATA directory")
}

// getWindowsDataDir returns the Windows data directory.
func getWindowsDataDir() (string, error) {
	if localAppData, ok := envAbsPath("LOCALAPPDATA", "windows"); ok {
		return localAppData, nil
	}

	if userProfile, ok := envAbsPath("USERPROFILE", "windows"); ok {
		return filepath.Join(userProfile, "AppData", "Local"), nil
	}

	return "", fmt.Errorf("cannot determine Windows LOCALAPPDATA directory")
}

// getMacOSAppleConfigDir returns the traditional macOS configuration directory.
func getMacOSAppleConfigDir() (string, error) {
	if envDir, ok := envAbsPath("OSDIR_MACOS_APPLE_CONFIG_HOME", "darwin"); ok {
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
	if envDir, ok := envAbsPath("OSDIR_MACOS_APPLE_DATA_HOME", "darwin"); ok {
		return envDir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	return filepath.Join(home, "Library", "Application Support"), nil
}
