package cfgly

import "errors"

// Sentinel errors returned (wrapped) by ConfigBuilder.Build. Check them with
// errors.Is.
var (
	// ErrInvalidAppName reports an application name that is missing or is not
	// a single plain path segment (see ConfigBuilder.WithName).
	ErrInvalidAppName = errors.New("invalid app name")
)
