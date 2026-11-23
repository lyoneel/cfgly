# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/2.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.0] - 2026-11-23

Security and correctness release that hardens directory resolution and defaults. Requires Go 1.26.11 or later and dirly v1.2.0.

### Added

- `ConfigBuilder.WithSymlinkPolicy(SymlinkPolicy)` passes a symlink policy to dirly for the config and data directories: `FollowWithinRoot` (default), `FollowAny` or `NoFollow`. `SymlinkPolicy`, the policy constants, `DefaultSymlinkPolicy`, `ErrPathEscapesRoot` and `ErrSymlinkNotAllowed` are re-exported from dirly.
- `ConfigBuilder.WithMacOSStrategy(MacOSStrategy)` selects the Apple (default) or XDG directory layout on macOS; the XDG strategy honors `XDG_CONFIG_HOME` and `XDG_DATA_HOME`. Adds the `DefaultMacOSStrategy` constant.
- `ConfigBuilder.WithDirPerm(os.FileMode)` and `ConfigBuilder.WithFilePerm(os.FileMode)`, plus `DefaultDirPerm` (`0o700`) and `DefaultFilePerm` (`0o600`).
- `ConfigBuilder.WithConfigExtensions(extensions ...string)` restricts the config directory to the given file extensions, and `ConfigDir.Extensions()` returns them.
- Sentinel errors `ErrInvalidAppName`, `ErrInvalidBaseDir`, `ErrInvalidPerm`, `ErrInvalidMacOSStrategy`, `ErrInvalidSymlinkPolicy` and `ErrExtensionNotAllowed`.

### Changed

- **Breaking:** cfgly now requires dirly v1.2.0. Writes are atomic (temp file, fsync, rename), so replacing a file needs write permission on its directory, and ownership errors are returned instead of ignored.
- **Breaking:** `WithConfigDir` and `WithDataDir` reject empty or relative base directories with `ErrInvalidBaseDir`. To upgrade, pass an absolute path.
- **Breaking:** A filtered `ConfigDir` (see `WithConfigExtensions`) rejects reads and writes of disallowed files with `ErrExtensionNotAllowed`; previously only `Exists` and listing honored the filter. To upgrade, include every extension you read or write.
- The README no longer hard-codes a coverage percentage.

### Fixed

- **Breaking:** Environment variable overrides (`XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `APPDATA`, `LOCALAPPDATA`, `USERPROFILE`, `OSDIR_MACOS_APPLE_CONFIG_HOME`, `OSDIR_MACOS_APPLE_DATA_HOME`) are used only when they hold a non-empty absolute path; empty or relative values fall back as if unset, per the XDG Base Directory Specification. To upgrade, set these variables to absolute paths.
- `Build()` returns a new, independent `AppConfig` on every call and no longer modifies the builder, so reusing a builder cannot change previously built configs. `AppConfig` is now actually immutable and safe for concurrent use; the builder is documented as not goroutine-safe.
- Extensions given to `WithConfigExtensions` with a leading dot or in upper case now match (previously `".yaml"` hid every file).
- `WithConfigDir` and `WithDataDir` work in any order relative to `WithName`; `Build()` appends the app name (previously calling them before `WithName` silently dropped the name).
- The macOS strategy advertised in the README can now actually be configured (previously the Apple layout was hard-coded).

### Security

- GO-2026-4602: **Breaking:** the minimum Go version is now 1.26.11 (`go.mod`), which clears this and other standard-library advisories reported against earlier 1.26.x releases. CI pins `golang:1.26.11`. To upgrade, build with Go 1.26.11 or later.
- **Breaking:** Symlinks inside the config or data directory that resolve outside it are refused by default with `ErrPathEscapesRoot` (through dirly v1.2.0), and missing parent directories created by cfgly writes go through an `os.Root` under `FollowWithinRoot` and `NoFollow`. If your app deliberately links files there to outside locations, use `WithSymlinkPolicy(cfgly.FollowAny)`; a config directory that is itself a symlink needs no change.
- **Breaking:** The app name must be a single plain path segment, so it can no longer move the config or data directory outside its base. `.`, `..`, names containing `/`, `\`, `:`, NUL or other control characters, and Windows reserved device names (`CON`, `NUL`, `COM1`, ...) are rejected with `ErrInvalidAppName`, and `Build()` fails when `WithName` was never called (previously it silently returned the bare base directory). To upgrade, use a name such as `myapp` or `com.example.app`.
- **Breaking:** Directories created by cfgly (`WithAutoCreate`, and missing parents created by `ConfigDir` writes) now use `0o700` instead of `0o755`, and new files written with perm `0` (including batch writes) use `0o600` instead of `0o644`. Existing files and directories are never chmod-ed. To keep the old modes, use `WithDirPerm(0o755)` and `WithFilePerm(0o644)`.

## [1.1.0] - 2026-08-17

### Changed

- **Breaking:** The package is flattened to a single root package `cfgly`; the former `lycfg` subpackage and its leftover directory are gone. To upgrade, change imports from `gitlab.com/lyoneel/cfgly/lycfg` to `gitlab.com/lyoneel/cfgly`.
- Documentation now refers to the renamed dirly package.

## [1.0.0] - 2026-08-13

### Changed

- cfgly now builds on the dirly v1.0.0 release.

## [0.0.1] - 2026-08-10

### Added

- `WithAutoCreate` builder option for automatic config and data directory creation during `Build()`.

## [0.0.0] - 2026-07-02

### Added

- Initial public snapshot with cross-platform application configuration.
- OS-specific directory resolution for Linux (XDG), macOS (Apple and XDG strategies), and Windows (APPDATA, LOCALAPPDATA).
- Builder pattern with `WithName`, `WithConfigDir`, `WithDataDir`.
- Environment variable overrides for all OS-specific defaults.
- Instance-based API with no global state.
- `ConfigDir` wrapper around dirly with glob search, text reading, and existence checks.

[Unreleased]: https://gitlab.com/lyoneel/cfgly/-/compare/v1.2.0...HEAD
[1.2.0]: https://gitlab.com/lyoneel/cfgly/-/compare/v1.1.0...v1.2.0
[1.1.0]: https://gitlab.com/lyoneel/cfgly/-/tags/v1.1.0
