//go:build unix

package cfgly

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// symlinkFixture builds a config dir containing a link to a directory outside
// it, a link to a file outside it, and an in-root link.
func symlinkFixture(t *testing.T, policy SymlinkPolicy) (cfg *AppConfig, outside string) {
	t.Helper()
	root := t.TempDir()
	outside = filepath.Join(root, "outside")
	if err := os.MkdirAll(outside, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "secret.yaml"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	b := NewAppConfigBuilder().WithName("app").WithConfigDir(filepath.Join(root, "cfg")).
		WithDataDir(filepath.Join(root, "data")).WithAutoCreate(true)
	if policy != DefaultSymlinkPolicy {
		b = b.WithSymlinkPolicy(policy)
	}
	cfg, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	dir := cfg.ConfigDir()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.Symlink(outside, filepath.Join(dir, "outlink")))
	must(os.Symlink(filepath.Join(outside, "secret.yaml"), filepath.Join(dir, "settings.yaml")))
	must(os.MkdirAll(filepath.Join(dir, "sub"), 0o700))
	must(os.WriteFile(filepath.Join(dir, "sub", "real.yaml"), []byte("inside"), 0o600))
	must(os.Symlink(filepath.Join("sub", "real.yaml"), filepath.Join(dir, "inlink.yaml")))
	return cfg, outside
}

func TestSymlink_DefaultRefusesEscapes(t *testing.T) {
	cfg, outside := symlinkFixture(t, DefaultSymlinkPolicy)
	c := cfg.Config()

	if _, err := c.ReadText("outlink/secret.yaml"); !errors.Is(err, ErrPathEscapesRoot) {
		t.Errorf("read through escaping dir link: err = %v, want ErrPathEscapesRoot", err)
	}
	if err := c.WriteFromString("settings.yaml", "pwned", 0); !errors.Is(err, ErrPathEscapesRoot) {
		t.Errorf("write through escaping file link: err = %v, want ErrPathEscapesRoot", err)
	}
	if err := c.WriteFromString("outlink/new/deep.yaml", "x", 0); !errors.Is(err, ErrPathEscapesRoot) {
		t.Errorf("write below escaping dir link: err = %v, want ErrPathEscapesRoot", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "new")); !os.IsNotExist(err) {
		t.Error("parent directories were created outside the config dir")
	}
	if b, _ := os.ReadFile(filepath.Join(outside, "secret.yaml")); string(b) != "secret" {
		t.Errorf("outside file modified: %q", b)
	}
	if c.Exists("outlink/secret.yaml") {
		t.Error("Exists followed an escaping link")
	}

	// In-root links still work.
	if s, err := c.ReadText("inlink.yaml"); err != nil || s != "inside" {
		t.Errorf("in-root link: %q, %v", s, err)
	}
	if err := c.WriteFromString("sub/new/x.yaml", "ok", 0); err != nil {
		t.Errorf("nested write inside root: %v", err)
	}
}

func TestSymlink_FollowAnyRestoresOldBehavior(t *testing.T) {
	cfg, outside := symlinkFixture(t, FollowAny)
	c := cfg.Config()
	if s, err := c.ReadText("outlink/secret.yaml"); err != nil || s != "secret" {
		t.Errorf("FollowAny read: %q, %v", s, err)
	}
	if err := c.WriteFromString("outlink/new/deep.yaml", "x", 0); err != nil {
		t.Errorf("FollowAny nested write: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "new", "deep.yaml")); err != nil {
		t.Errorf("FollowAny write did not land at the link target: %v", err)
	}
}

func TestSymlink_NoFollowRefusesAllLinks(t *testing.T) {
	cfg, _ := symlinkFixture(t, NoFollow)
	c := cfg.Config()
	if _, err := c.ReadText("inlink.yaml"); !errors.Is(err, ErrSymlinkNotAllowed) {
		t.Errorf("NoFollow in-root link: err = %v, want ErrSymlinkNotAllowed", err)
	}
	if _, err := c.ReadText("sub/real.yaml"); err != nil {
		t.Errorf("NoFollow plain file: %v", err)
	}
}

func TestSymlink_SymlinkedConfigDirWorks(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "sync", "app")
	if err := os.MkdirAll(real, 0o700); err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(root, "cfg")
	if err := os.MkdirAll(base, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, filepath.Join(base, "app")); err != nil {
		t.Fatal(err)
	}
	for _, p := range []SymlinkPolicy{FollowWithinRoot, FollowAny, NoFollow} {
		cfg, err := NewAppConfigBuilder().WithName("app").WithConfigDir(base).WithSymlinkPolicy(p).Build()
		if err != nil {
			t.Fatal(err)
		}
		if err := cfg.Config().WriteFromString("nested/c.yaml", p.String(), 0); err != nil {
			t.Errorf("%v: write in symlinked config dir: %v", p, err)
			continue
		}
		if b, err := os.ReadFile(filepath.Join(real, "nested", "c.yaml")); err != nil || string(b) != p.String() {
			t.Errorf("%v: file not at real location: %q, %v", p, b, err)
		}
	}
}
