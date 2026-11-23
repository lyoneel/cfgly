package cfgly

import (
	"errors"
	"path/filepath"
	"testing"
)

func darwinEnv(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, k := range []string{"XDG_CONFIG_HOME", "XDG_DATA_HOME", "OSDIR_MACOS_APPLE_CONFIG_HOME", "OSDIR_MACOS_APPLE_DATA_HOME"} {
		t.Setenv(k, "")
	}
	return home
}

func buildFor(t *testing.T, goos string, b *ConfigBuilder) *AppConfig {
	t.Helper()
	b.goos = goos
	cfg, err := b.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return cfg
}

func TestMacOSStrategy_DefaultIsApple(t *testing.T) {
	home := darwinEnv(t)
	t.Setenv("XDG_CONFIG_HOME", "/xdg/config") // must be ignored by the Apple strategy
	cfg := buildFor(t, "darwin", NewAppConfigBuilder().WithName("app"))
	if want := filepath.Join(home, "Library", "Preferences", "app"); cfg.ConfigDir() != want {
		t.Errorf("ConfigDir = %q, want %q", cfg.ConfigDir(), want)
	}
	if want := filepath.Join(home, "Library", "Application Support", "app"); cfg.DataDir() != want {
		t.Errorf("DataDir = %q, want %q", cfg.DataDir(), want)
	}
	if DefaultMacOSStrategy != MacOSStrategyApple {
		t.Errorf("DefaultMacOSStrategy = %q, want apple", DefaultMacOSStrategy)
	}
}

func TestMacOSStrategy_AppleHonorsOverrides(t *testing.T) {
	darwinEnv(t)
	t.Setenv("OSDIR_MACOS_APPLE_CONFIG_HOME", "/apple/cfg")
	t.Setenv("OSDIR_MACOS_APPLE_DATA_HOME", "/apple/data")
	cfg := buildFor(t, "darwin", NewAppConfigBuilder().WithName("app").WithMacOSStrategy(MacOSStrategyApple))
	if cfg.ConfigDir() != filepath.Join("/apple/cfg", "app") || cfg.DataDir() != filepath.Join("/apple/data", "app") {
		t.Errorf("got %q, %q", cfg.ConfigDir(), cfg.DataDir())
	}
}

func TestMacOSStrategy_XDGDefaults(t *testing.T) {
	home := darwinEnv(t)
	cfg := buildFor(t, "darwin", NewAppConfigBuilder().WithName("app").WithMacOSStrategy(MacOSStrategyXDG))
	if want := filepath.Join(home, ".config", "app"); cfg.ConfigDir() != want {
		t.Errorf("ConfigDir = %q, want %q", cfg.ConfigDir(), want)
	}
	if want := filepath.Join(home, ".local", "share", "app"); cfg.DataDir() != want {
		t.Errorf("DataDir = %q, want %q", cfg.DataDir(), want)
	}
}

func TestMacOSStrategy_XDGHonorsXDGEnv(t *testing.T) {
	home := darwinEnv(t)
	t.Setenv("XDG_CONFIG_HOME", "/xdg/config")
	t.Setenv("XDG_DATA_HOME", "relative/data") // relative: ignored, falls back
	t.Setenv("OSDIR_MACOS_APPLE_CONFIG_HOME", "/apple/cfg")
	cfg := buildFor(t, "darwin", NewAppConfigBuilder().WithMacOSStrategy(MacOSStrategyXDG).WithName("app"))
	if want := filepath.Join("/xdg/config", "app"); cfg.ConfigDir() != want {
		t.Errorf("ConfigDir = %q, want %q", cfg.ConfigDir(), want)
	}
	if want := filepath.Join(home, ".local", "share", "app"); cfg.DataDir() != want {
		t.Errorf("DataDir = %q, want %q", cfg.DataDir(), want)
	}
}

func TestMacOSStrategy_IgnoredOnLinux(t *testing.T) {
	home := darwinEnv(t)
	cfg := buildFor(t, "linux", NewAppConfigBuilder().WithName("app").WithMacOSStrategy(MacOSStrategyApple))
	if want := filepath.Join(home, ".config", "app"); cfg.ConfigDir() != want {
		t.Errorf("ConfigDir = %q, want %q", cfg.ConfigDir(), want)
	}
}

func TestMacOSStrategy_Invalid(t *testing.T) {
	for _, s := range []MacOSStrategy{"", "Apple", "linux", "xdg "} {
		_, err := NewAppConfigBuilder().WithName("app").WithMacOSStrategy(s).Build()
		if !errors.Is(err, ErrInvalidMacOSStrategy) {
			t.Errorf("WithMacOSStrategy(%q): err = %v, want ErrInvalidMacOSStrategy", s, err)
		}
	}
}
