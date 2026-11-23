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
)
