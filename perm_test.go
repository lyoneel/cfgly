package cfgly

import (
	"errors"
	"os"
	"testing"
)

func TestBuilder_PermValidation(t *testing.T) {
	base := t.TempDir()
	for _, p := range []os.FileMode{0, 0o600, 0o500, 0o1777, os.ModeDir | 0o700} {
		_, err := NewAppConfigBuilder().WithName("a").WithConfigDir(base).WithDataDir(base).WithDirPerm(p).Build()
		if !errors.Is(err, ErrInvalidPerm) {
			t.Errorf("WithDirPerm(%#o): err = %v, want ErrInvalidPerm", uint32(p), err)
		}
	}
	for _, p := range []os.FileMode{0, 0o400, 0o200, 0o4600, os.ModeDir | 0o600} {
		_, err := NewAppConfigBuilder().WithName("a").WithConfigDir(base).WithDataDir(base).WithFilePerm(p).Build()
		if !errors.Is(err, ErrInvalidPerm) {
			t.Errorf("WithFilePerm(%#o): err = %v, want ErrInvalidPerm", uint32(p), err)
		}
	}
	for _, p := range []os.FileMode{0o700, 0o750, 0o755, 0o777} {
		if _, err := NewAppConfigBuilder().WithName("a").WithConfigDir(base).WithDirPerm(p).Build(); err != nil {
			t.Errorf("WithDirPerm(%#o): unexpected error %v", uint32(p), err)
		}
	}
	for _, p := range []os.FileMode{0o600, 0o640, 0o644, 0o666} {
		if _, err := NewAppConfigBuilder().WithName("a").WithConfigDir(base).WithFilePerm(p).Build(); err != nil {
			t.Errorf("WithFilePerm(%#o): unexpected error %v", uint32(p), err)
		}
	}
}
