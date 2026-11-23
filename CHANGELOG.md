# Changelog

All notable changes in the cfgly project are documented in this file. The format is based on Keep a Changelog, and the project follows Semantic Versioning.

## v1.2.0 - Unreleased

### Added

- `ConfigBuilder.WithMacOSStrategy(MacOSStrategy)` selects the Apple (default) or XDG directory layout on macOS; the XDG strategy honors `XDG_CONFIG_HOME` and `XDG_DATA_HOME`. `DefaultMacOSStrategy` constant
- `ConfigBuilder.WithDirPerm(os.FileMode)` and `ConfigBuilder.WithFilePerm(os.FileMode)`, plus `DefaultDirPerm` (`0o700`) and `DefaultFilePerm` (`0o600`)
- Sentinel errors `ErrInvalidAppName`, `ErrInvalidBaseDir`, `ErrInvalidPerm` and `ErrInvalidMacOSStrategy`
- `ConfigBuilder.WithConfigExtensions(extensions ...string)` restricts the config directory to the given file extensions
- `ConfigDir.Extensions()` returns the allowed extensions
- `ErrExtensionNotAllowed` sentinel error for reads and writes of disallowed files on a filtered `ConfigDir`

### Changed

- Minimum Go version is now 1.26.11 (`go.mod`); earlier 1.26.x releases have known standard-library vulnerabilities. CI pins `golang:1.26.11`
- README no longer hard-codes a coverage percentage
- Directories created by cfgly (`WithAutoCreate`, and missing parents created by `ConfigDir` writes) now use `0o700` instead of `0o755`. New files written with perm `0` (including batch writes) now use `0o600` instead of `0o644`. Existing files and directories are never chmod-ed
- `WithConfigDir` and `WithDataDir` reject empty or relative base directories (`ErrInvalidBaseDir`)

### Fixed

- Environment variable overrides (`XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `APPDATA`, `LOCALAPPDATA`, `USERPROFILE`, `OSDIR_MACOS_APPLE_CONFIG_HOME`, `OSDIR_MACOS_APPLE_DATA_HOME`) are used only when they hold a non-empty absolute path; empty or relative values fall back as if unset, per the XDG Base Directory Specification
- `Build()` returns a new, independent `AppConfig` on every call and no longer modifies the builder, so reusing a builder cannot change previously built configs. `AppConfig` is now actually immutable and safe for concurrent use; the builder is documented as not goroutine-safe
- Extension filters are applied consistently: `ReadText` and all other read, write, and batch methods on a filtered `ConfigDir` now reject disallowed files (previously only `Exists` and listing honored the filter), and extensions given with a leading dot or in upper case now match (previously `".yaml"` hid every file)
- The app name must be a single plain path segment. `.`, `..`, names containing `/`, `\`, `:`, NUL or other control characters, and Windows reserved device names (`CON`, `NUL`, `COM1`, ...) are rejected with `ErrInvalidAppName`, so a name can no longer move the config or data directory outside its base. `Build()` also fails when `WithName` was never called (previously it silently returned the bare base directory)
- `WithConfigDir` and `WithDataDir` now work in any order relative to `WithName`; `Build()` appends the app name (previously calling them before `WithName` silently dropped the name)
- The macOS strategy advertised in the README can now actually be configured (previously the Apple layout was hard-coded)

### Upgrade Notes

This release tightens validation and defaults. Code that relied on the old behavior may need changes:

- App names containing path separators, `:`, control characters, or Windows reserved names, as well as `.` and `..`, now fail `Build()`.
- `WithConfigDir("")` and relative bases such as `WithConfigDir("cfg")` now fail `Build()`; pass an absolute path.
- Relative or empty override environment variables are ignored instead of being used.
- New directories are `0700` and new files `0600`. To keep the old modes, use `WithDirPerm(0o755)` and `WithFilePerm(0o644)`.
- A filtered `ConfigDir` now rejects reads and writes of disallowed files.
- Go 1.26.11 or later is required.

## v1.1.0 - 2026-08-17

### Changed

- Breaking: flatten the package to a single root package `cfgly` (the former `lycfg` subpackage is gone)
- Documentation now refers to the renamed dirly package
- Remove the leftover `lycfg` directory after flattening

### Upgrade Notes

Code that imported the subpackage changes its import path:

```go
// before
import "gitlab.com/lyoneel/cfgly/lycfg"
// after
import "gitlab.com/lyoneel/cfgly"
```

## v1.0.0 - 2026-08-13

### Changed

- Dependency on dirly moves to its v1.0.0 release

## v0.0.1 - 2026-08-10

### Added

- `WithAutoCreate` builder option for automatic config and data directory creation during `Build()`

## v0.0.0 - 2026-07-02

### Added

- Initial public snapshot with cross-platform application configuration
- OS-specific directory resolution for Linux (XDG), macOS (Apple and XDG strategies), and Windows (APPDATA, LOCALAPPDATA)
- Builder pattern with `WithName`, `WithConfigDir`, `WithDataDir`
- Environment variable overrides for all OS-specific defaults
- Instance-based API with no global state
- `ConfigDir` wrapper around dirly with glob search, text reading, and existence checks

## Statistics

- 9 commits across all tags
- 12 tracked files, 7 Go source files, about 800 lines of Go
- Test coverage 75.2 percent of statements
