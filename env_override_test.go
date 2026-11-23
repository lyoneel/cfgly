package cfgly

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestIsAbsForOS(t *testing.T) {
	tests := []struct {
		path string
		goos string
		want bool
	}{
		{"/home/u/.config", "linux", true},
		{"/Users/u/Library", "darwin", true},
		{"", "linux", false},
		{"relative/dir", "linux", false},
		{"./dir", "linux", false},
		{"../dir", "linux", false},
		{"~/dir", "linux", false},
		{`C:\Users\u\AppData\Roaming`, "linux", false},
		{`C:\Users\u\AppData\Roaming`, "windows", true},
		{`c:/Users/u`, "windows", true},
		{`\\server\share\dir`, "windows", true},
		{`\\?\C:\Users\u`, "windows", true},
		{`C:Users\u`, "windows", false},
		{`\Users\u`, "windows", false},
		{`Users\u`, "windows", false},
		{`\\\x`, "windows", false},
		{"", "windows", false},
		{"/unix/style", "windows", false},
	}
	for _, tt := range tests {
		if got := isAbsForOS(tt.path, tt.goos); got != tt.want {
			t.Errorf("isAbsForOS(%q, %q) = %v, want %v", tt.path, tt.goos, got, tt.want)
		}
	}
}

// TestEnvOverridesIgnoreRelativeAndEmpty checks that every supported override
// variable is ignored (falls back as if unset) when empty or relative.
func TestEnvOverridesIgnoreRelativeAndEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", "") // keep Windows fallbacks deterministic unless set below

	type resolver func() (string, error)
	tests := []struct {
		name     string
		key      string
		fn       resolver
		absValue string
		fallback string // expected result when the variable is ignored ("" = error expected)
	}{
		{"XDG_CONFIG_HOME", "XDG_CONFIG_HOME", getLinuxConfigDir, "/abs/xdg/config", filepath.Join(home, ".config")},
		{"XDG_DATA_HOME", "XDG_DATA_HOME", getLinuxDataDir, "/abs/xdg/data", filepath.Join(home, ".local", "share")},
		{"OSDIR_MACOS_APPLE_CONFIG_HOME", "OSDIR_MACOS_APPLE_CONFIG_HOME", getMacOSAppleConfigDir, "/abs/apple/config", filepath.Join(home, "Library", "Preferences")},
		{"OSDIR_MACOS_APPLE_DATA_HOME", "OSDIR_MACOS_APPLE_DATA_HOME", getMacOSAppleDataDir, "/abs/apple/data", filepath.Join(home, "Library", "Application Support")},
		{"APPDATA", "APPDATA", getWindowsConfigDir, `C:\Users\u\AppData\Roaming`, ""},
		{"LOCALAPPDATA", "LOCALAPPDATA", getWindowsDataDir, `C:\Users\u\AppData\Local`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/absolute is used", func(t *testing.T) {
			t.Setenv(tt.key, tt.absValue)
			got, err := tt.fn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.absValue {
				t.Errorf("got %q, want %q", got, tt.absValue)
			}
		})
		for _, bad := range []string{"", "relative/dir", "./dir", "../dir", "dir"} {
			t.Run(tt.name+"/ignored "+bad, func(t *testing.T) {
				t.Setenv(tt.key, bad)
				got, err := tt.fn()
				if tt.fallback == "" {
					if err == nil {
						t.Fatalf("expected error when %s=%q and no fallback, got %q", tt.key, bad, got)
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tt.fallback {
					t.Errorf("%s=%q: got %q, want fallback %q", tt.key, bad, got, tt.fallback)
				}
			})
		}
	}
}

func TestWindowsUserProfileFallbackValidation(t *testing.T) {
	t.Setenv("APPDATA", "relative")
	t.Setenv("LOCALAPPDATA", "")

	t.Setenv("USERPROFILE", `C:\Users\u`)
	if got, err := getWindowsConfigDir(); err != nil || got != filepath.Join(`C:\Users\u`, "AppData", "Roaming") {
		t.Errorf("config: got %q, %v", got, err)
	}
	if got, err := getWindowsDataDir(); err != nil || got != filepath.Join(`C:\Users\u`, "AppData", "Local") {
		t.Errorf("data: got %q, %v", got, err)
	}

	for _, bad := range []string{"", `Users\u`, "C:Users"} {
		t.Setenv("USERPROFILE", bad)
		if got, err := getWindowsConfigDir(); err == nil {
			t.Errorf("USERPROFILE=%q: expected error, got %q", bad, got)
		}
		if got, err := getWindowsDataDir(); err == nil {
			t.Errorf("USERPROFILE=%q: expected error, got %q", bad, got)
		}
	}
}

func TestBuildIgnoresRelativeXDG(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skipf("XDG resolution only applies on linux, running on %s", runtime.GOOS)
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "relative/dir")
	t.Setenv("XDG_DATA_HOME", "")

	cfg, err := NewAppConfigBuilder().WithName("myapp").Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if want := filepath.Join(home, ".config", "myapp"); cfg.ConfigDir() != want {
		t.Errorf("ConfigDir = %q, want %q", cfg.ConfigDir(), want)
	}
	if want := filepath.Join(home, ".local", "share", "myapp"); cfg.DataDir() != want {
		t.Errorf("DataDir = %q, want %q", cfg.DataDir(), want)
	}
}
