package cfgly

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestBuilder_BaseDirCallOrderDoesNotMatter(t *testing.T) {
	cfgBase, dataBase := t.TempDir(), t.TempDir()

	before, err := NewAppConfigBuilder().WithConfigDir(cfgBase).WithDataDir(dataBase).WithName("myapp").Build()
	if err != nil {
		t.Fatalf("dirs before name: %v", err)
	}
	after, err := NewAppConfigBuilder().WithName("myapp").WithConfigDir(cfgBase).WithDataDir(dataBase).Build()
	if err != nil {
		t.Fatalf("dirs after name: %v", err)
	}
	for _, c := range []*AppConfig{before, after} {
		if want := filepath.Join(cfgBase, "myapp"); c.ConfigDir() != want {
			t.Errorf("ConfigDir = %q, want %q", c.ConfigDir(), want)
		}
		if want := filepath.Join(dataBase, "myapp"); c.DataDir() != want {
			t.Errorf("DataDir = %q, want %q", c.DataDir(), want)
		}
	}
}

func TestBuilder_RenameAfterBaseDirUsesNewName(t *testing.T) {
	base := t.TempDir()
	b := NewAppConfigBuilder().WithName("one").WithConfigDir(base).WithDataDir(base)
	c1, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	c2, err := b.WithName("two").Build()
	if err != nil {
		t.Fatal(err)
	}
	if c1.ConfigDir() != filepath.Join(base, "one") || c2.ConfigDir() != filepath.Join(base, "two") {
		t.Errorf("got %q and %q", c1.ConfigDir(), c2.ConfigDir())
	}
	if c2.DataDir() != filepath.Join(base, "two") {
		t.Errorf("DataDir = %q", c2.DataDir())
	}
}

func TestBuilder_RejectsEmptyOrRelativeBaseDirs(t *testing.T) {
	for _, dir := range []string{"", "relative", "./cfg", "../cfg", "cfg/sub"} {
		_, err := NewAppConfigBuilder().WithName("myapp").WithConfigDir(dir).Build()
		if !errors.Is(err, ErrInvalidBaseDir) {
			t.Errorf("WithConfigDir(%q): err = %v, want ErrInvalidBaseDir", dir, err)
		}
		_, err = NewAppConfigBuilder().WithDataDir(dir).WithName("myapp").Build()
		if !errors.Is(err, ErrInvalidBaseDir) {
			t.Errorf("WithDataDir(%q): err = %v, want ErrInvalidBaseDir", dir, err)
		}
	}
}
