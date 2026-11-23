package cfgly

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func newFilteredFixture(t *testing.T, exts ...string) (string, *ConfigDir) {
	t.Helper()
	dir := t.TempDir()
	for name, content := range map[string]string{
		"app.yaml":        "a: 1\n",
		"other.JSON":      "{}\n",
		"secret.txt":      "token\n",
		"noext":           "x\n",
		"sub/nested.yaml": "n: 1\n",
		"sub/nested.env":  "K=V\n",
	} {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir, NewConfigDir(dir, exts...)
}

func TestNormalizeExtensions(t *testing.T) {
	got := normalizeExtensions([]string{".YAML", "yaml", " json ", "", ".", "Toml"})
	want := []string{"yaml", "json", "toml"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("normalizeExtensions = %v, want %v", got, want)
	}
	if got := NewConfigDir(t.TempDir()).Extensions(); got != nil {
		t.Errorf("unfiltered Extensions() = %v, want nil", got)
	}
}

func TestConfigDir_ExtensionsAccessorReturnsCopy(t *testing.T) {
	d := NewConfigDir(t.TempDir(), "yaml")
	exts := d.Extensions()
	exts[0] = "txt"
	if d.Extensions()[0] != "yaml" {
		t.Error("mutating Extensions() result changed the ConfigDir")
	}
}

func TestConfigDir_FilterAppliedConsistently(t *testing.T) {
	// Leading dot and mixed case must work (dirly itself expects "yaml").
	_, d := newFilteredFixture(t, ".YAML", "json")

	// Allowed files.
	for _, name := range []string{"app.yaml", "other.JSON", "sub/nested.yaml"} {
		if !d.Exists(name) {
			t.Errorf("Exists(%q) = false, want true", name)
		}
		if _, err := d.ReadText(name); err != nil {
			t.Errorf("ReadText(%q) error: %v", name, err)
		}
	}

	// Disallowed files: every read path refuses them, Exists says false.
	for _, name := range []string{"secret.txt", "noext", "sub/nested.env"} {
		if d.Exists(name) {
			t.Errorf("Exists(%q) = true, want false", name)
		}
		checks := map[string]error{}
		_, checks["ReadText"] = d.ReadText(name)
		_, checks["ReadToBytes"] = d.ReadToBytes(name)
		_, checks["ReadToString"] = d.ReadToString(name)
		_, checks["ReadToLines"] = d.ReadToLines(name)
		f, _, err := d.ReadToBuff(name)
		checks["ReadToBuff"] = err
		if f != nil {
			f.Close()
		}
		f, _, err = d.ReadToScanner(name)
		checks["ReadToScanner"] = err
		if f != nil {
			f.Close()
		}
		_, checks["BatchReadToBytes"] = d.BatchReadToBytes([]string{"app.yaml", name})
		_, checks["BatchReadToString"] = d.BatchReadToString([]string{name})
		_, checks["BatchReadToLines"] = d.BatchReadToLines([]string{name})
		_, checks["BatchReadToBuff"] = d.BatchReadToBuff([]string{name})
		_, checks["BatchReadToScanner"] = d.BatchReadToScanner([]string{name})
		for op, err := range checks {
			if !errors.Is(err, ErrExtensionNotAllowed) {
				t.Errorf("%s(%q) error = %v, want ErrExtensionNotAllowed", op, name, err)
			}
		}
	}

	// Listing and glob only surface allowed files.
	rel, err := d.GetAllRel()
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(rel)
	want := []string{"app.yaml", "other.JSON", filepath.FromSlash("sub/nested.yaml")}
	if !reflect.DeepEqual(rel, want) {
		t.Errorf("GetAllRel = %v, want %v", rel, want)
	}
	matches, err := d.GetByGlob("*")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range matches {
		if ext := strings.ToLower(filepath.Ext(m)); ext != ".yaml" && ext != ".json" {
			t.Errorf("GetByGlob returned disallowed file %q", m)
		}
	}
	if len(matches) != 3 {
		t.Errorf("GetByGlob(*) returned %d files (%v), want 3", len(matches), matches)
	}
}

func TestConfigDir_WritesRespectFilter(t *testing.T) {
	dir, d := newFilteredFixture(t, "yaml")

	if err := d.WriteFromString("new.yaml", "ok: true\n", 0o600); err != nil {
		t.Fatalf("WriteFromString(new.yaml): %v", err)
	}

	writes := map[string]func(string) error{
		"WriteFromBytes":   func(n string) error { return d.WriteFromBytes(n, []byte("x"), 0o600) },
		"WriteFromString":  func(n string) error { return d.WriteFromString(n, "x", 0o600) },
		"WriteFromLines":   func(n string) error { return d.WriteFromLines(n, []string{"x"}, 0o600) },
		"WriteFromBuff":    func(n string) error { return d.WriteFromBuff(n, bufio.NewReader(strings.NewReader("x")), 0o600) },
		"WriteFromScanner": func(n string) error { return d.WriteFromScanner(n, bufio.NewScanner(strings.NewReader("x")), 0o600) },
		"BatchWriteFromBytes": func(n string) error {
			return d.BatchWriteFromBytes(map[string][]byte{"ok.yaml": []byte("x"), n: []byte("x")})
		},
		"BatchWriteFromString": func(n string) error { return d.BatchWriteFromString(map[string]string{n: "x"}) },
		"BatchWriteFromLines":  func(n string) error { return d.BatchWriteFromLines(map[string][]string{n: {"x"}}) },
		"BatchWriteFromBuff": func(n string) error {
			return d.BatchWriteFromBuff(map[string]*bufio.Reader{n: bufio.NewReader(strings.NewReader("x"))})
		},
		"BatchWriteFromScanner": func(n string) error {
			return d.BatchWriteFromScanner(map[string]*bufio.Scanner{n: bufio.NewScanner(strings.NewReader("x"))})
		},
	}
	for op, write := range writes {
		name := "blocked-" + strings.ToLower(op) + ".sh"
		if err := write(name); !errors.Is(err, ErrExtensionNotAllowed) {
			t.Errorf("%s(%q) error = %v, want ErrExtensionNotAllowed", op, name, err)
		}
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Errorf("%s created disallowed file %q", op, name)
		}
	}
	// Batch writes validate every name before writing anything.
	if _, err := os.Stat(filepath.Join(dir, "ok.yaml")); !os.IsNotExist(err) {
		t.Error("BatchWriteFromBytes wrote ok.yaml although another name was rejected")
	}
}

func TestConfigDir_NoFilterAllowsEverything(t *testing.T) {
	_, d := newFilteredFixture(t)
	for _, name := range []string{"app.yaml", "secret.txt", "noext"} {
		if !d.Exists(name) {
			t.Errorf("Exists(%q) = false on unfiltered dir", name)
		}
		if _, err := d.ReadText(name); err != nil {
			t.Errorf("ReadText(%q) on unfiltered dir: %v", name, err)
		}
	}
	// Empty names keep dirly's own validation error.
	if _, err := d.ReadText(""); err == nil || errors.Is(err, ErrExtensionNotAllowed) {
		t.Errorf("ReadText(\"\") error = %v, want dirly validation error", err)
	}
}

func TestBuilder_WithConfigExtensions(t *testing.T) {
	base := t.TempDir()
	cfg, err := NewAppConfigBuilder().
		WithName("app").
		WithConfigDir(base).
		WithDataDir(base).
		WithConfigExtensions("yaml", ".toml").
		WithAutoCreate(true).
		Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if got, want := cfg.Config().Extensions(), []string{"yaml", "toml"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Config().Extensions() = %v, want %v", got, want)
	}
	if got := cfg.Data().Extensions(); got != nil {
		t.Errorf("Data().Extensions() = %v, want nil (data dir is never filtered)", got)
	}
	if err := cfg.Config().WriteFromString("token.txt", "x", 0o600); !errors.Is(err, ErrExtensionNotAllowed) {
		t.Errorf("config write of .txt: err = %v, want ErrExtensionNotAllowed", err)
	}
	if err := cfg.Data().WriteFromString("cache.bin", "x", 0o600); err != nil {
		t.Errorf("data write: %v", err)
	}

	// Last call wins; no arguments removes the filter.
	b := NewAppConfigBuilder().WithName("app").WithConfigDir(base).WithConfigExtensions("yaml")
	cfg2, err := b.WithConfigExtensions().Build()
	if err != nil {
		t.Fatal(err)
	}
	if cfg2.Config().Extensions() != nil {
		t.Errorf("WithConfigExtensions() should clear the filter, got %v", cfg2.Config().Extensions())
	}
	// Earlier result unaffected by later builder changes.
	if !reflect.DeepEqual(cfg.Config().Extensions(), []string{"yaml", "toml"}) {
		t.Error("earlier config's extensions changed")
	}
}
