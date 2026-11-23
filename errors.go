package cfgly

import "errors"

// Sentinel errors returned (wrapped) by ConfigBuilder.Build. Check them with
// errors.Is.
var (
	// ErrInvalidAppName reports an application name that is missing or is not
	// a single plain path segment (see ConfigBuilder.WithName).
	ErrInvalidAppName = errors.New("invalid app name")

	// ErrInvalidBaseDir reports an empty or relative base directory passed to
	// ConfigBuilder.WithConfigDir or ConfigBuilder.WithDataDir.
	ErrInvalidBaseDir = errors.New("invalid base directory")

	// ErrInvalidPerm reports a permission passed to ConfigBuilder.WithDirPerm
	// or ConfigBuilder.WithFilePerm that is zero, has bits outside 0o777, or
	// lacks the owner bits the library needs (rwx for directories, rw for
	// files).
	ErrInvalidPerm = errors.New("invalid permission")

	// ErrInvalidMacOSStrategy reports a value passed to
	// ConfigBuilder.WithMacOSStrategy other than MacOSStrategyApple or
	// MacOSStrategyXDG.
	ErrInvalidMacOSStrategy = errors.New("invalid macOS strategy")

	// ErrInvalidSymlinkPolicy reports a value passed to
	// ConfigBuilder.WithSymlinkPolicy other than FollowWithinRoot, FollowAny
	// or NoFollow.
	ErrInvalidSymlinkPolicy = errors.New("invalid symlink policy")
)
