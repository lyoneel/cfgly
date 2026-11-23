//go:build unix

package cfgly

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// withUmask runs f with the process umask set to mask (umask is process-wide,
// so these tests must not run in parallel).
func withUmask(t *testing.T, mask int, f func()) {
	t.Helper()
	old := syscall.Umask(mask)
	defer syscall.Umask(old)
	f()
}

func mode(t *testing.T, path string) os.FileMode {
	t.Helper()
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return st.Mode().Perm()
}

func TestPerm_DefaultsAreOwnerOnly(t *testing.T) {
	withUmask(t, 0, func() {
		base := t.TempDir()
		cfg, err := NewAppConfigBuilder().WithName("app").
			WithConfigDir(filepath.Join(base, "cfg")).
			WithDataDir(filepath.Join(base, "data")).
			WithAutoCreate(true).Build()
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range []string{cfg.ConfigDir(), cfg.DataDir(), filepath.Join(base, "cfg"), filepath.Join(base, "data")} {
			if got := mode(t, d); got != 0o700 {
				t.Errorf("dir %s mode = %#o, want 0700", d, got)
			}
		}

		if err := cfg.Config().WriteFromString("token.yaml", "secret", 0); err != nil {
			t.Fatal(err)
		}
		if got := mode(t, filepath.Join(cfg.ConfigDir(), "token.yaml")); got != 0o600 {
			t.Errorf("file mode = %#o, want 0600", got)
		}
		if err := cfg.Data().BatchWriteFromString(map[string]string{"cache/state.json": "{}"}); err != nil {
			t.Fatal(err)
		}
		if got := mode(t, filepath.Join(cfg.DataDir(), "cache")); got != 0o700 {
			t.Errorf("nested dir created by write: mode = %#o, want 0700", got)
		}
		if got := mode(t, filepath.Join(cfg.DataDir(), "cache", "state.json")); got != 0o600 {
			t.Errorf("batch-written file mode = %#o, want 0600", got)
		}

		// An explicit perm argument still wins.
		if err := cfg.Config().WriteFromString("public.yaml", "x", 0o644); err != nil {
			t.Fatal(err)
		}
		if got := mode(t, filepath.Join(cfg.ConfigDir(), "public.yaml")); got != 0o644 {
			t.Errorf("explicit perm: mode = %#o, want 0644", got)
		}
	})
}

func TestPerm_NewConfigDirDefaults(t *testing.T) {
	withUmask(t, 0, func() {
		base := filepath.Join(t.TempDir(), "standalone")
		d := NewConfigDir(base)
		if err := d.WriteFromString("sub/a.txt", "x", 0); err != nil {
			t.Fatal(err)
		}
		if got := mode(t, base); got != 0o700 {
			t.Errorf("base created by write: mode = %#o, want 0700", got)
		}
		if got := mode(t, filepath.Join(base, "sub", "a.txt")); got != 0o600 {
			t.Errorf("file mode = %#o, want 0600", got)
		}
	})
}

func TestPerm_CustomModes(t *testing.T) {
	withUmask(t, 0, func() {
		base := t.TempDir()
		cfg, err := NewAppConfigBuilder().WithName("app").WithConfigDir(base).WithDataDir(base).
			WithDirPerm(0o750).WithFilePerm(0o640).WithAutoCreate(true).Build()
		if err != nil {
			t.Fatal(err)
		}
		if got := mode(t, cfg.ConfigDir()); got != 0o750 {
			t.Errorf("dir mode = %#o, want 0750", got)
		}
		if err := cfg.Config().WriteFromBytes("x/y.conf", []byte("x"), 0); err != nil {
			t.Fatal(err)
		}
		if got := mode(t, filepath.Join(cfg.ConfigDir(), "x")); got != 0o750 {
			t.Errorf("nested dir mode = %#o, want 0750", got)
		}
		if got := mode(t, filepath.Join(cfg.ConfigDir(), "x", "y.conf")); got != 0o640 {
			t.Errorf("file mode = %#o, want 0640", got)
		}
	})
}

func TestPerm_ExistingDirsAndFilesAreNotChmodded(t *testing.T) {
	withUmask(t, 0, func() {
		base := t.TempDir()
		existing := filepath.Join(base, "app")
		if err := os.MkdirAll(existing, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(existing, 0o755); err != nil {
			t.Fatal(err)
		}
		f := filepath.Join(existing, "old.yaml")
		if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := NewAppConfigBuilder().WithName("app").WithConfigDir(base).WithDataDir(base).WithAutoCreate(true).Build()
		if err != nil {
			t.Fatal(err)
		}
		if err := cfg.Config().WriteFromString("old.yaml", "y", 0); err != nil {
			t.Fatal(err)
		}
		if got := mode(t, existing); got != 0o755 {
			t.Errorf("existing dir mode changed to %#o, want 0755", got)
		}
		if got := mode(t, f); got != 0o644 {
			t.Errorf("existing file mode changed to %#o, want 0644", got)
		}
	})
}

func TestPerm_UmaskStillApplies(t *testing.T) {
	withUmask(t, 0o077, func() {
		base := t.TempDir()
		cfg, err := NewAppConfigBuilder().WithName("app").WithConfigDir(base).WithDirPerm(0o755).WithAutoCreate(true).Build()
		if err != nil {
			t.Fatal(err)
		}
		if got := mode(t, cfg.ConfigDir()); got != 0o700 {
			t.Errorf("dir mode with umask 077 = %#o, want 0700", got)
		}
	})
}
