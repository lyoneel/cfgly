# Changelog

All notable changes in the cfgly project are documented in this file. The format is based on Keep a Changelog, and the project follows Semantic Versioning.

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
