package cfgly

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetConfigBaseForOS(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home dir: %v", err)
	}

	tests := []struct {
		name     string
		goos     string
		wantPath string
		envs     map[string]string
	}{
		{
			name:     "linux default",
			goos:     "linux",
			wantPath: filepath.Join(home, ".config"),
		},
		{
			name:     "linux with XDG_CONFIG_HOME",
			goos:     "linux",
			wantPath: "/custom/xdg/config",
			envs:     map[string]string{"XDG_CONFIG_HOME": "/custom/xdg/config"},
		},
		{
			name:     "darwin apple strategy",
			goos:     "darwin",
			wantPath: filepath.Join(home, "Library", "Preferences"),
		},
		{
			name:     "windows with APPDATA",
			goos:     "windows",
			wantPath: "C:\\Users\\test\\AppData\\Roaming",
			envs:     map[string]string{"APPDATA": `C:\Users\test\AppData\Roaming`},
		},
		{
			name:     "windows with USERPROFILE fallback",
			goos:     "windows",
			wantPath: filepath.Join(`C:\Users\test`, "AppData", "Roaming"),
			envs:     map[string]string{"USERPROFILE": `C:\Users\test`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars
			for k, v := range tt.envs {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envs {
					os.Unsetenv(k)
				}
			}()

			got, err := getConfigBaseForOS(tt.goos)
			if err != nil {
				t.Fatalf("getConfigBaseForOS(%q) error = %v", tt.goos, err)
			}
			// Normalize path separators for cross-platform comparison
			gotNorm := filepath.ToSlash(got)
			wantNorm := filepath.ToSlash(tt.wantPath)
			if gotNorm != wantNorm {
				t.Errorf("getConfigBaseForOS(%q) = %q, want %q", tt.goos, gotNorm, wantNorm)
			}
		})
	}
}

func TestGetDataBaseForOS(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home dir: %v", err)
	}

	tests := []struct {
		name     string
		goos     string
		wantPath string
		envs     map[string]string
	}{
		{
			name:     "linux default",
			goos:     "linux",
			wantPath: filepath.Join(home, ".local", "share"),
		},
		{
			name:     "linux with XDG_DATA_HOME",
			goos:     "linux",
			wantPath: "/custom/xdg/data",
			envs:     map[string]string{"XDG_DATA_HOME": "/custom/xdg/data"},
		},
		{
			name:     "darwin apple strategy",
			goos:     "darwin",
			wantPath: filepath.Join(home, "Library", "Application Support"),
		},
		{
			name:     "windows with LOCALAPPDATA",
			goos:     "windows",
			wantPath: `C:\Users\test\AppData\Local`,
			envs:     map[string]string{"LOCALAPPDATA": `C:\Users\test\AppData\Local`},
		},
		{
			name:     "windows with USERPROFILE fallback",
			goos:     "windows",
			wantPath: filepath.Join(`C:\Users\test`, "AppData", "Local"),
			envs:     map[string]string{"USERPROFILE": `C:\Users\test`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars
			for k, v := range tt.envs {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envs {
					os.Unsetenv(k)
				}
			}()

			got, err := getDataBaseForOS(tt.goos)
			if err != nil {
				t.Fatalf("getDataBaseForOS(%q) error = %v", tt.goos, err)
			}
			// Normalize path separators for cross-platform comparison
			gotNorm := filepath.ToSlash(got)
			wantNorm := filepath.ToSlash(tt.wantPath)
			if gotNorm != wantNorm {
				t.Errorf("getDataBaseForOS(%q) = %q, want %q", tt.goos, gotNorm, wantNorm)
			}
		})
	}
}

func TestGetWindowsConfigDir_NoEnv(t *testing.T) {
	// Save and clear env vars
	appData := os.Getenv("APPDATA")
	userProfile := os.Getenv("USERPROFILE")
	os.Unsetenv("APPDATA")
	os.Unsetenv("USERPROFILE")
	defer func() {
		if appData != "" {
			os.Setenv("APPDATA", appData)
		}
		if userProfile != "" {
			os.Setenv("USERPROFILE", userProfile)
		}
	}()

	_, err := getWindowsConfigDir()
	if err == nil {
		t.Error("expected error when no Windows env vars are set")
	}
}

func TestGetWindowsDataDir_NoEnv(t *testing.T) {
	// Save and clear env vars
	localAppData := os.Getenv("LOCALAPPDATA")
	userProfile := os.Getenv("USERPROFILE")
	os.Unsetenv("LOCALAPPDATA")
	os.Unsetenv("USERPROFILE")
	defer func() {
		if localAppData != "" {
			os.Setenv("LOCALAPPDATA", localAppData)
		}
		if userProfile != "" {
			os.Setenv("USERPROFILE", userProfile)
		}
	}()

	_, err := getWindowsDataDir()
	if err == nil {
		t.Error("expected error when no Windows env vars are set")
	}
}

func TestGetConfigBaseForOS_UnknownOS(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home dir: %v", err)
	}

	got, err := getConfigBaseForOS("unknown")
	if err != nil {
		t.Fatalf("getConfigBaseForOS(\"unknown\") error = %v", err)
	}
	wantPath := filepath.Join(home, ".config")
	if got != wantPath {
		t.Errorf("getConfigBaseForOS(\"unknown\") = %q, want %q", got, wantPath)
	}
}

func TestGetDataBaseForOS_UnknownOS(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home dir: %v", err)
	}

	got, err := getDataBaseForOS("unknown")
	if err != nil {
		t.Fatalf("getDataBaseForOS(\"unknown\") error = %v", err)
	}
	wantPath := filepath.Join(home, ".local", "share")
	if got != wantPath {
		t.Errorf("getDataBaseForOS(\"unknown\") = %q, want %q", got, wantPath)
	}
}

func TestBuilder_ConfigDirUsesConfigBase(t *testing.T) {
	// Set XDG_CONFIG_HOME and XDG_DATA_HOME to different values
	os.Setenv("XDG_CONFIG_HOME", "/custom/config")
	os.Setenv("XDG_DATA_HOME", "/custom/data")
	defer func() {
		os.Unsetenv("XDG_CONFIG_HOME")
		os.Unsetenv("XDG_DATA_HOME")
	}()

	// Only test on Linux where XDG is respected
	if runtime.GOOS != "linux" {
		t.Skip("XDG test only applies to Linux")
	}

	cfg, err := NewAppConfigBuilder().WithName("testapp").Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Config dir should use XDG_CONFIG_HOME, not XDG_DATA_HOME
	expectedConfigDir := filepath.Join("/custom/config", "testapp")
	if cfg.ConfigDir() != expectedConfigDir {
		t.Errorf("ConfigDir() = %q, want %q (should use config base, not data base)", cfg.ConfigDir(), expectedConfigDir)
	}

	// Data dir should use XDG_DATA_HOME
	expectedDataDir := filepath.Join("/custom/data", "testapp")
	if cfg.DataDir() != expectedDataDir {
		t.Errorf("DataDir() = %q, want %q", cfg.DataDir(), expectedDataDir)
	}
}
