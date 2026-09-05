package cfgly

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDirEmbedding(t *testing.T) {
	tmpDir := t.TempDir()

	cfgDir := NewConfigDir(tmpDir)

	if cfgDir.BasePath() != tmpDir {
		t.Errorf("BasePath() = %q, want %q", cfgDir.BasePath(), tmpDir)
	}

	if cfgDir.Directory == nil {
		t.Fatal("Directory field should be embedded")
	}
}

func TestConfigDirInheritedMethods(t *testing.T) {
	tmpDir := t.TempDir()

	createTestFile(t, tmpDir, "test.txt", "hello world")

	cfgDir := NewConfigDir(tmpDir)

	// Test ReadText (inherited from Directory)
	content, err := cfgDir.ReadText("test.txt")
	if err != nil {
		t.Fatalf("ReadText failed: %v", err)
	}
	if content != "hello world" {
		t.Errorf("ReadText() = %q, want %q", content, "hello world")
	}

	// Test Exists (inherited from Directory)
	if !cfgDir.Exists("test.txt") {
		t.Error("Exists() should return true for existing file")
	}

	if cfgDir.Exists("nonexistent.txt") {
		t.Error("Exists() should return false for non-existing file")
	}

	// Test WriteFile (inherited from Directory)
	err = cfgDir.WriteFromBytes("output.txt", []byte("written"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if !cfgDir.Exists("output.txt") {
		t.Error("WriteFile should create file that Exists() can find")
	}
}

func createTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", name, err)
	}
}
