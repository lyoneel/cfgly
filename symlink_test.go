package cfgly

import (
	"errors"
	"testing"

	"gitlab.com/lyoneel/dirly"
)

func TestSymlinkPolicy_DefaultAndPassThrough(t *testing.T) {
	base := t.TempDir()
	cfg, err := NewAppConfigBuilder().WithName("app").WithConfigDir(base).WithDataDir(base).Build()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Config().SymlinkPolicy() != dirly.FollowWithinRoot || cfg.Data().SymlinkPolicy() != dirly.FollowWithinRoot {
		t.Errorf("default policy = %v/%v, want FollowWithinRoot", cfg.Config().SymlinkPolicy(), cfg.Data().SymlinkPolicy())
	}
	if DefaultSymlinkPolicy != dirly.FollowWithinRoot {
		t.Errorf("DefaultSymlinkPolicy = %v", DefaultSymlinkPolicy)
	}
	if NewConfigDir(base).SymlinkPolicy() != FollowWithinRoot {
		t.Error("NewConfigDir should default to FollowWithinRoot")
	}

	for _, p := range []SymlinkPolicy{FollowWithinRoot, FollowAny, NoFollow} {
		cfg, err := NewAppConfigBuilder().WithName("app").WithConfigDir(base).WithDataDir(base).WithSymlinkPolicy(p).Build()
		if err != nil {
			t.Fatalf("WithSymlinkPolicy(%v): %v", p, err)
		}
		if cfg.Config().SymlinkPolicy() != p || cfg.Data().SymlinkPolicy() != p {
			t.Errorf("WithSymlinkPolicy(%v): got %v/%v", p, cfg.Config().SymlinkPolicy(), cfg.Data().SymlinkPolicy())
		}
	}
}

func TestSymlinkPolicy_Invalid(t *testing.T) {
	for _, p := range []SymlinkPolicy{SymlinkPolicy(-1), SymlinkPolicy(3), SymlinkPolicy(99)} {
		_, err := NewAppConfigBuilder().WithName("app").WithConfigDir(t.TempDir()).WithSymlinkPolicy(p).Build()
		if !errors.Is(err, ErrInvalidSymlinkPolicy) {
			t.Errorf("WithSymlinkPolicy(%v): err = %v, want ErrInvalidSymlinkPolicy", p, err)
		}
	}
}

func TestSymlinkPolicy_ReexportsMatchDirly(t *testing.T) {
	if ErrPathEscapesRoot != dirly.ErrPathEscapesRoot || ErrSymlinkNotAllowed != dirly.ErrSymlinkNotAllowed {
		t.Error("re-exported errors differ from dirly's")
	}
	if FollowAny != dirly.FollowAny || NoFollow != dirly.NoFollow {
		t.Error("re-exported policies differ from dirly's")
	}
}
