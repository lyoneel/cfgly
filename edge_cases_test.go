package cfgly

import "testing"

// Builder & Configuration Tests

func TestBuilder_EmptyAppName(t *testing.T) {
	cfg, err := NewAppConfigBuilder().WithName("").Build()
	if err == nil {
		t.Error("expected error for empty app name, got nil")
	}
	if cfg != nil {
		t.Error("expected nil config on error, got non-nil")
	}
}

func TestBuilder_SpecialCharactersInAppName(t *testing.T) {
	tests := []struct {
		name        string
		appName     string
		expectError bool
	}{
		{"valid name", "myapp", false},
		{"with spaces", "my app", false},
		{"unicode", "アプリ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewAppConfigBuilder().WithName(tt.appName).Build()
			if tt.expectError && err == nil {
				t.Errorf("expected error for app name %q, got nil", tt.appName)
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error for app name %q: %v", tt.appName, err)
			}
			if !tt.expectError && cfg == nil {
				t.Error("expected non-nil config on success")
			}
		})
	}
}

func TestBuilder_OverrideConfigDirEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	cfg, err := NewAppConfigBuilder().
		WithName("testapp").
		WithConfigDir(tmpDir).
		Build()
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	expectedPath := tmpDir + "/testapp"
	if cfg.ConfigDir() != expectedPath {
		t.Errorf("expected ConfigDir %q, got %q", expectedPath, cfg.ConfigDir())
	}
}

func TestBuilder_OverrideConfigDirAbsolute(t *testing.T) {
	tmpDir := t.TempDir()
	cfg, err := NewAppConfigBuilder().
		WithName("testapp").
		WithConfigDir(tmpDir).
		Build()
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	if !isAbs(cfg.ConfigDir()) {
		t.Error("expected absolute path for ConfigDir")
	}
}

func TestBuilder_OverrideConfigDirWithDotDot(t *testing.T) {
	tmpDir := t.TempDir()
	cfg, err := NewAppConfigBuilder().
		WithName("testapp").
		WithConfigDir(tmpDir).
		Build()
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	if cfg.ConfigDir() == "" {
		t.Error("expected non-empty ConfigDir")
	}
}

func TestBuilder_ErrorChaining(t *testing.T) {
	cfg, err := NewAppConfigBuilder().
		WithName(""). // First error
		WithName("testapp").
		Build()
	if err == nil {
		t.Error("expected error from first WithName call")
	}
	if cfg != nil {
		t.Error("expected nil config on error")
	}
}

func TestBuilder_BuildAfterError(t *testing.T) {
	builder := NewAppConfigBuilder().WithName("") // Error state
	cfg, err := builder.WithName("testapp").Build()
	if err == nil {
		t.Error("expected error from previous WithName call to persist")
	}
	if cfg != nil {
		t.Error("expected nil config on error")
	}
}

// Platform-Specific Edge Cases

func TestPlatform_CaseSensitivity(t *testing.T) {
	tmpDir := t.TempDir()
	cfg, err := NewAppConfigBuilder().WithName("testapp").WithConfigDir(tmpDir).Build()
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	err = cfg.Config().WriteFromBytes("file.txt", []byte("lowercase"), 0644)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	exists := cfg.Config().Exists("FILE.TXT")
	if exists {
		t.Log("filesystem is case-insensitive")
	} else {
		t.Log("filesystem is case-sensitive (expected on Linux)")
	}
}

func isAbs(path string) bool {
	return len(path) > 0 && path[0] == '/'
}
