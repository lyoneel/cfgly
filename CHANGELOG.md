# Changelog

All notable changes in the cfgly project are documented in this file. The format is based on Keep a Changelog, and the project follows Semantic Versioning.

## Unreleased

### Added

- `ConfigBuilder.WithConfigExtensions(extensions ...string)` restricts the config directory to the given file extensions
- `ConfigDir.Extensions()` returns the allowed extensions
- `ErrExtensionNotAllowed` sentinel error for reads and writes of disallowed files on a filtered `ConfigDir`

### Changed

- Minimum Go version is now 1.26.11 (`go.mod`); earlier 1.26.x releases have known standard-library vulnerabilities. CI pins `golang:1.26.11`
- README no longer hard-codes a coverage percentage

### Fixed

- Environment variable overrides (`XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `APPDATA`, `LOCALAPPDATA`, `USERPROFILE`, `OSDIR_MACOS_APPLE_CONFIG_HOME`, `OSDIR_MACOS_APPLE_DATA_HOME`) are used only when they hold a non-empty absolute path; empty or relative values fall back as if unset, per the XDG Base Directory Specification
- `Build()` returns a new, independent `AppConfig` on every call and no longer modifies the builder, so reusing a builder cannot change previously built configs. `AppConfig` is now actually immutable and safe for concurrent use; the builder is documented as not goroutine-safe
- Extension filters are applied consistently: `ReadText` and all other read, write, and batch methods on a filtered `ConfigDir` now reject disallowed files (previously only `Exists` and listing honored the filter), and extensions given with a leading dot or in upper case now match (previously `".yaml"` hid every file)

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
