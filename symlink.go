package cfgly

import "gitlab.com/lyoneel/dirly"

// SymlinkPolicy controls how the config and data directories treat symbolic
// links found inside them. It is dirly's SymlinkPolicy, re-exported so
// callers do not need to import dirly. The directory root itself is always
// resolved once, so a config directory that is itself a symlink (for example
// ~/.config/app -> ~/Sync/app) works under every policy.
type SymlinkPolicy = dirly.SymlinkPolicy

const (
	// FollowWithinRoot follows symlinks inside the directory but refuses any
	// operation whose resolved target lands outside it (errors wrap
	// ErrPathEscapesRoot). This is the default.
	FollowWithinRoot = dirly.FollowWithinRoot

	// FollowAny follows symlinks wherever they point, including outside the
	// directory (the behavior of cfgly v1.1.0 and earlier).
	FollowAny = dirly.FollowAny

	// NoFollow refuses any path that goes through a symlink below the
	// directory root (errors wrap ErrSymlinkNotAllowed).
	NoFollow = dirly.NoFollow

	// DefaultSymlinkPolicy is the policy used when WithSymlinkPolicy is not
	// called.
	DefaultSymlinkPolicy = FollowWithinRoot
)

var (
	// ErrPathEscapesRoot is dirly.ErrPathEscapesRoot: a path resolves through a
	// symlink to a location outside the directory.
	ErrPathEscapesRoot = dirly.ErrPathEscapesRoot

	// ErrSymlinkNotAllowed is dirly.ErrSymlinkNotAllowed: a path goes through a
	// symlink while the NoFollow policy is active.
	ErrSymlinkNotAllowed = dirly.ErrSymlinkNotAllowed
)

func validSymlinkPolicy(p SymlinkPolicy) bool {
	switch p {
	case FollowWithinRoot, FollowAny, NoFollow:
		return true
	}
	return false
}
