package cfgly

import (
	"errors"
	"testing"
)

func TestValidateAppName(t *testing.T) {
	valid := []string{
		"myapp", "my-app", "my_app", "my app", "com.example.app", "MyApp2",
		"アプリ", "app.v2", "console", "nullable", "com10", "lpt", "auxiliary", ".hidden",
	}
	for _, n := range valid {
		if err := validateAppName(n); err != nil {
			t.Errorf("validateAppName(%q) = %v, want nil", n, err)
		}
	}

	invalid := []string{
		"", ".", "..",
		"a/b", "../etc", "../../../etc", "/etc/cron.d", "/", "app/",
		`a\b`, `..\evil`, `C:\Windows`, `\\server\share`,
		"C:", "c:app", "app:stream",
		"evil\x00x", "line\nbreak", "tab\tname", "del\x7f",
		"CON", "con", "Con.txt", "NUL", "nul.json", "PRN", "AUX",
		"COM1", "com9.log", "LPT1", "lpt9", "COM¹", "CONIN$", "CON .txt",
	}
	for _, n := range invalid {
		err := validateAppName(n)
		if !errors.Is(err, ErrInvalidAppName) {
			t.Errorf("validateAppName(%q) = %v, want ErrInvalidAppName", n, err)
		}
	}
}

func TestBuild_RejectsInvalidAppNames(t *testing.T) {
	base := t.TempDir()
	for _, n := range []string{"", ".", "..", "../../../etc", "a/b", `a\b`, "C:", "x\x00y", "NUL"} {
		cfg, err := NewAppConfigBuilder().WithName(n).WithConfigDir(base).WithDataDir(base).Build()
		if !errors.Is(err, ErrInvalidAppName) {
			t.Errorf("Build with name %q: err = %v, want ErrInvalidAppName", n, err)
		}
		if cfg != nil {
			t.Errorf("Build with name %q returned a config", n)
		}
	}
}

func TestBuild_RequiresName(t *testing.T) {
	cfg, err := NewAppConfigBuilder().WithConfigDir(t.TempDir()).Build()
	if !errors.Is(err, ErrInvalidAppName) {
		t.Fatalf("Build without WithName: err = %v, want ErrInvalidAppName", err)
	}
	if cfg != nil {
		t.Error("Build without WithName returned a config")
	}
}
