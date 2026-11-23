package cfgly

import (
	"fmt"
	"strings"
)

// windowsReservedNames are device names Windows reserves in every directory,
// with or without an extension ("NUL", "nul.txt"). They are rejected on all
// platforms so that a configuration that works on one OS works on all.
var windowsReservedNames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
	"COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
	"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
	"COM¹": true, "COM²": true, "COM³": true,
	"LPT¹": true, "LPT²": true, "LPT³": true,
	"CONIN$": true, "CONOUT$": true,
}

// validateAppName checks that name is a single plain path segment that is
// safe to append to a base directory on every supported OS. It rejects:
//   - the empty string, "." and "..";
//   - path separators ('/' and '\'), so the name can never add or climb levels;
//   - ':' (Windows drive letters such as "C:" and NTFS alternate data streams);
//   - NUL and other ASCII control characters;
//   - Windows reserved device names (CON, PRN, AUX, NUL, COM1-9, LPT1-9, ...),
//     case-insensitively and with or without an extension.
//
// Names such as "myapp", "my-app", "my app" and "com.example.app" are valid.
func validateAppName(name string) error {
	switch name {
	case "":
		return fmt.Errorf("%w: appName cannot be empty", ErrInvalidAppName)
	case ".", "..":
		return fmt.Errorf("%w: %q is not allowed", ErrInvalidAppName, name)
	}
	for _, r := range name {
		switch {
		case r == '/' || r == '\\':
			return fmt.Errorf("%w: %q must be a single path segment (contains %q)", ErrInvalidAppName, name, r)
		case r == ':':
			return fmt.Errorf("%w: %q must not contain ':' (drive or stream syntax)", ErrInvalidAppName, name)
		case r < 0x20 || r == 0x7f:
			return fmt.Errorf("%w: %q contains a control character", ErrInvalidAppName, name)
		}
	}
	stem := name
	if i := strings.IndexByte(stem, '.'); i >= 0 {
		stem = stem[:i]
	}
	stem = strings.TrimRight(stem, " ")
	if windowsReservedNames[strings.ToUpper(stem)] {
		return fmt.Errorf("%w: %q is a reserved device name on Windows", ErrInvalidAppName, name)
	}
	return nil
}
