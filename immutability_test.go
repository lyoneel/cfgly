package cfgly

import (
	"path/filepath"
	"sync"
	"testing"
)

func TestBuild_ReturnsIndependentConfigs(t *testing.T) {
	base := t.TempDir()
	b := NewAppConfigBuilder().WithName("a").WithConfigDir(base).WithDataDir(base)

	c1, err := b.Build()
	if err != nil {
		t.Fatalf("first Build: %v", err)
	}
	c2, err := b.Build()
	if err != nil {
		t.Fatalf("second Build: %v", err)
	}
	if c1 == c2 {
		t.Fatal("Build returned the same *AppConfig twice; want independent values")
	}
	if c1.Config() == c2.Config() || c1.Data() == c2.Data() {
		t.Error("configs share ConfigDir instances")
	}
}

func TestBuild_BuilderReuseDoesNotAffectEarlierResults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("APPDATA", "")
	t.Setenv("LOCALAPPDATA", "")
	t.Setenv("OSDIR_MACOS_APPLE_CONFIG_HOME", "")
	t.Setenv("OSDIR_MACOS_APPLE_DATA_HOME", "")

	b := NewAppConfigBuilder().WithName("a")
	c1, err := b.Build()
	if err != nil {
		t.Fatalf("Build a: %v", err)
	}
	name1, cfgDir1, dataDir1 := c1.Name(), c1.ConfigDir(), c1.DataDir()

	c2, err := b.WithName("b").Build()
	if err != nil {
		t.Fatalf("Build b: %v", err)
	}

	if c1.Name() != name1 || c1.ConfigDir() != cfgDir1 || c1.DataDir() != dataDir1 {
		t.Errorf("earlier config changed after builder reuse: name=%q config=%q data=%q",
			c1.Name(), c1.ConfigDir(), c1.DataDir())
	}
	if c1.Name() != "a" || filepath.Base(c1.ConfigDir()) != "a" || filepath.Base(c1.DataDir()) != "a" {
		t.Errorf("c1 = {%q, %q, %q}, want name and dirs for %q", c1.Name(), c1.ConfigDir(), c1.DataDir(), "a")
	}
	// The second build resolves its own paths from the new name instead of
	// reusing paths cached by the first build.
	if c2.Name() != "b" || filepath.Base(c2.ConfigDir()) != "b" || filepath.Base(c2.DataDir()) != "b" {
		t.Errorf("c2 = {%q, %q, %q}, want name and dirs for %q", c2.Name(), c2.ConfigDir(), c2.DataDir(), "b")
	}
}

func TestBuild_DoesNotMutateBuilderOverrides(t *testing.T) {
	base := t.TempDir()
	b := NewAppConfigBuilder().WithName("a").WithConfigDir(base)
	if _, err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if b.dataDirCustomPath != "" {
		t.Errorf("Build cached a resolved data path in the builder: %q", b.dataDirCustomPath)
	}
	if want := filepath.Join(base, "a"); b.configDirCustomPath != want {
		t.Errorf("configDirCustomPath = %q, want %q", b.configDirCustomPath, want)
	}
}

// TestAppConfig_ConcurrentAccess must be run with -race: AppConfig accessors
// are read from many goroutines while the (single-goroutine) builder that
// produced it keeps being reconfigured and rebuilt.
func TestAppConfig_ConcurrentAccess(t *testing.T) {
	base := t.TempDir()
	b := NewAppConfigBuilder().WithName("shared").WithConfigDir(base).WithDataDir(base)
	cfg, err := b.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	wantName, wantCfg, wantData := cfg.Name(), cfg.ConfigDir(), cfg.DataDir()

	var wg sync.WaitGroup
	errs := make(chan string, 64)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				if cfg.Name() != wantName || cfg.ConfigDir() != wantCfg || cfg.DataDir() != wantData ||
					cfg.Config().BasePath() != wantCfg || cfg.Data().BasePath() != wantData {
					errs <- "AppConfig changed while being read concurrently"
					return
				}
			}
		}()
	}

	for i := 0; i < 50; i++ {
		if _, err := b.WithName("other").WithConfigDir(base).Build(); err != nil {
			t.Fatalf("rebuild: %v", err)
		}
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}
