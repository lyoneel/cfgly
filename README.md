# cfgly - Cross-Platform Application Configuration with OS-Specific Directory Resolution

A standalone Go library for cross-platform application configuration, providing OS-specific directory resolution following platform conventions.

> Main repository: https://gitlab.com/lyoneel/cfgly
> Any other host that serves this repository is a mirror. Open issues
> and merge requests on GitLab.

## Features

- **Cross-platform** directory resolution (Linux XDG, macOS Apple/XDG, Windows APPDATA)
- **Builder pattern** producing immutable configuration values
- **Configurable macOS strategy** — choose between traditional Apple directories and XDG-style paths with `WithMacOSStrategy`
- **Safe defaults** — app names must be a single path segment, and directories and files are created owner-only (`0700`/`0600`)
- **Environment variable overrides** — all OS-specific defaults can be overridden via environment variables (absolute paths only)
- **Instance-based API** with no global state

## Installation

```bash
go get gitlab.com/lyoneel/cfgly
```

```go
import "gitlab.com/lyoneel/cfgly"
```

Requires Go 1.26.11 or later (the minimum in `go.mod`; earlier 1.26.x releases have known standard-library vulnerabilities). The library depends on [dirly](https://gitlab.com/lyoneel/dirly) for directory operations.

## Quick Start

```go
package main

import (
    "fmt"
    "gitlab.com/lyoneel/cfgly"
)

func main() {
    // Create an app config with OS-specific directory resolution
    cfg, err := cfgly.NewAppConfigBuilder().
        WithName("myapp").
        Build()
    if err != nil {
        panic(err)
    }

    fmt.Println(cfg.ConfigDir())  // ~/.config/myapp (Linux), ~/Library/Preferences/myapp (macOS), %APPDATA%\myapp (Windows)
    fmt.Println(cfg.DataDir())    // ~/.local/share/myapp (Linux), ~/Library/Application Support/myapp (macOS), %LOCALAPPDATA%\myapp (Windows)
}
```

## Architecture

The package has three cooperating parts:

1. `AppConfig` is the immutable result. It carries the application name and a `*ConfigDir` for configuration and data.
2. `ConfigBuilder` resolves the OS-specific base directories at `Build()` time and can create the directories on demand.
3. `ConfigDir` wraps a `dirly.Directory` and inherits the full dirly operation set, with config-file convenience methods on top.

Directory resolution happens once per `Build()` call and never changes afterwards. Each `Build()` returns a new, independent `*AppConfig`; reusing or reconfiguring the builder later never affects configs that were already built. `AppConfig` and `ConfigDir` methods are safe for concurrent use because their state is immutable. The `ConfigBuilder` itself is not safe for concurrent use: configure and build it from one goroutine, then share the result.

## Directory Resolution Matrix

| Platform | Config Directory | Data Directory |
|----------|------------------|----------------|
| Linux | `$XDG_CONFIG_HOME/myapp` (default `~/.config/myapp`) | `$XDG_DATA_HOME/myapp` (default `~/.local/share/myapp`) |
| macOS (Apple strategy, default) | `~/Library/Preferences/myapp` | `~/Library/Application Support/myapp` |
| macOS (XDG strategy, `WithMacOSStrategy(cfgly.MacOSStrategyXDG)`) | `$XDG_CONFIG_HOME/myapp` (default `~/.config/myapp`) | `$XDG_DATA_HOME/myapp` (default `~/.local/share/myapp`) |
| Windows | `%APPDATA%\myapp` | `%LOCALAPPDATA%\myapp` |

## API Reference

### AppConfig Creation

#### `NewAppConfigBuilder() *ConfigBuilder`

Creates a new builder for configuring an application instance.

```go
cfg, err := cfgly.NewAppConfigBuilder().
    WithName("myapp").
    Build()
```

### Builder Methods

All builder methods return the same builder instance for method chaining. Configuration is applied when `Build()` is called.

| Method | Description |
|--------|-------------|
| `WithName(name string)` | Set the application name (required). Must be a single plain path segment; see [App Name Rules](#app-name-rules). |
| `WithConfigDir(dir string)` | Override the configuration base directory. Must be a non-empty absolute path. `Build()` appends the app name, so call order does not matter. |
| `WithDataDir(dir string)` | Override the data base directory. Must be a non-empty absolute path. `Build()` appends the app name, so call order does not matter. |
| `WithMacOSStrategy(strategy MacOSStrategy)` | Choose `MacOSStrategyApple` (default) or `MacOSStrategyXDG` on macOS. No effect on other systems. See [macOS Strategy](#macos-strategy). |
| `WithDirPerm(perm os.FileMode)` | Mode for directories cfgly creates. Default `0o700`. Must include owner `rwx`. See [Permissions](#permissions). |
| `WithFilePerm(perm os.FileMode)` | Default mode for new files written with perm `0`. Default `0o600`. Must include owner `rw`. |
| `WithConfigExtensions(extensions ...string)` | Restrict the configuration directory to files with these extensions (case-insensitive, leading dot optional). Each call replaces the previous set; no arguments removes the filter. The data directory is never filtered. See [Extension Filtering](#extension-filtering). |
| `WithAutoCreate(create bool)` | Create config and data directories during `Build()`. Defaults to false. |
| `Build() (*AppConfig, error)` | Resolve OS-specific directories and create the final instance. Returns the first builder error (invalid name, base directory, permission or strategy), or an error if directory resolution or creation fails. |

### App Name Rules

The app name is appended to the base directories, so it must be a single plain path segment. `WithName` (and `Build()` when no name was set) rejects, with an error wrapping `cfgly.ErrInvalidAppName`:

- an empty name, `.` and `..`;
- names containing `/` or `\`;
- names containing `:` (Windows drive letters such as `C:` and NTFS stream syntax);
- NUL and other control characters;
- Windows reserved device names (`CON`, `PRN`, `AUX`, `NUL`, `COM1`-`COM9`, `LPT1`-`LPT9`, ...), case-insensitively and with or without an extension (`nul.txt`). These are rejected on every OS so a name that works on one platform works on all.

Names such as `myapp`, `my-app`, `my app` and `com.example.app` are valid.

### Permissions

Config directories often hold credentials, so cfgly creates everything owner-only by default:

- Directories created by `WithAutoCreate(true)`, and missing parent directories created by a `ConfigDir` write, use `0o700` (`cfgly.DefaultDirPerm`), or the mode set with `WithDirPerm`.
- New files written with perm `0`, including all batch writes, use `0o600` (`cfgly.DefaultFilePerm`), or the mode set with `WithFilePerm`. A non-zero perm passed to a write method is used as given.
- Existing directories and files are never chmod-ed, so their modes stay as they are. The process umask still applies.

`NewConfigDir` uses the same defaults.

### AppConfig Methods

| Method | Description |
|--------|-------------|
| `Name() string` | Returns the application name. |
| `ConfigDir() string` | Returns the resolved configuration directory path. |
| `DataDir() string` | Returns the resolved data directory path. |
| `Config() *ConfigDir` | Returns the configuration directory instance for file operations. |
| `Data() *ConfigDir` | Returns the data directory instance for file operations. |

### ConfigDir Methods

The `*ConfigDir` type wraps `dirly.Directory` with config-file-specific convenience methods.

| Method | Description |
|--------|-------------|
| `GetByGlob(pattern string) ([]string, error)` | Search for files matching a pattern and return absolute paths. |
| `ReadText(filename string) (string, error)` | Read a file and return its content as a string. |
| `Exists(filename string) bool` | Check if a file exists in the directory (and has an allowed extension). |
| `Extensions() []string` | Return a copy of the allowed extensions (lower-case, no dot), or nil when unfiltered. |

Because `ConfigDir` embeds `dirly.Directory`, every dirly operation is available directly, including batch reads and writes, glob search, and file management. See the [dirly documentation](https://gitlab.com/lyoneel/dirly) for the full operation set.

### Extension Filtering

A `ConfigDir` can be limited to a set of file extensions, either with `WithConfigExtensions` on the builder or with `NewConfigDir(basePath, extensions...)`. Extensions are case-insensitive and may be written with or without the leading dot (`"yaml"`, `".YAML"`). When a filter is set it applies consistently:

- Listing and search (`GetByGlob`, `GetAllByGlobAbs/Rel`, `GetAllAbs/Rel`) only return files with an allowed extension.
- `Exists` returns false for any other file.
- Every read and write (`ReadText`, `ReadTo*`, `WriteFrom*`, `BatchReadTo*`, `BatchWriteFrom*`) rejects other names with an error wrapping `cfgly.ErrExtensionNotAllowed`, before touching the file system. Batch operations validate all names first, so nothing is read or written if any name is rejected.
- Files without an extension are not allowed while a filter is set.
- Path-management operations (`Remove`, `BatchRemove`, `BatchRemoveAllFromDir`, `Chown`, `Chmod`, `AbsPath`, `SubDirectory`, `GetAllDirs*`) are not filtered.

```go
cfg, err := cfgly.NewAppConfigBuilder().
    WithName("myapp").
    WithConfigExtensions("yaml", "json").
    Build()
// ...
_, err = cfg.Config().ReadText("token.txt")
errors.Is(err, cfgly.ErrExtensionNotAllowed) // true
```

### macOS Strategy

macOS supports two directory resolution strategies:

| Strategy | Config Dir | Data Dir |
|----------|-----------|----------|
| `MacOSStrategyApple` (default) | `~/Library/Preferences/appName` (override base: `OSDIR_MACOS_APPLE_CONFIG_HOME`) | `~/Library/Application Support/appName` (override base: `OSDIR_MACOS_APPLE_DATA_HOME`) |
| `MacOSStrategyXDG` | `$XDG_CONFIG_HOME/appName` (default `~/.config/appName`) | `$XDG_DATA_HOME/appName` (default `~/.local/share/appName`) |

Select the strategy with `WithMacOSStrategy`; any other value makes `Build()` fail with `cfgly.ErrInvalidMacOSStrategy`:

```go
cfg, err := cfgly.NewAppConfigBuilder().
    WithName("myapp").
    WithMacOSStrategy(cfgly.MacOSStrategyXDG).
    Build()
```

### Environment Variable Overrides

All OS-specific defaults can be overridden via environment variables. A value is only used when it is a non-empty absolute path for that platform (`/...` on Linux and macOS; `C:\...` or a UNC path such as `\\server\share` on Windows). Empty or relative values are ignored and resolution falls back as if the variable were unset, as the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/latest/) requires.

| Variable | Description |
|----------|-------------|
| `XDG_CONFIG_HOME` | Override XDG config base directory (Linux, and macOS with `MacOSStrategyXDG`) |
| `XDG_DATA_HOME` | Override XDG data base directory (Linux, and macOS with `MacOSStrategyXDG`) |
| `APPDATA` | Override Windows config base directory |
| `LOCALAPPDATA` | Override Windows data base directory |
| `USERPROFILE` | Fallback for Windows directories |
| `OSDIR_MACOS_APPLE_CONFIG_HOME` | Override macOS Apple config base directory (`MacOSStrategyApple`) |
| `OSDIR_MACOS_APPLE_DATA_HOME` | Override macOS Apple data base directory (`MacOSStrategyApple`) |

## Error Handling

`Build()` returns an error when the app name is missing or invalid (`ErrInvalidAppName`), a base directory passed to `WithConfigDir`/`WithDataDir` is empty or relative (`ErrInvalidBaseDir`), a mode passed to `WithDirPerm`/`WithFilePerm` is invalid (`ErrInvalidPerm`), the macOS strategy is unknown (`ErrInvalidMacOSStrategy`), the home directory is inaccessible, or automatic directory creation fails. The builder keeps the first error and later setters do not clear it. Errors wrap the underlying cause with `fmt.Errorf` and `%w`, so callers can inspect them with `errors.Is` and `errors.As`. Read and write operations return the errors from the underlying dirly operations unchanged, except that on a filtered `ConfigDir` a disallowed file name returns an error wrapping `ErrExtensionNotAllowed`.

## Tech Stack and Dependencies

- Go 1.26.11 or later
- [dirly](https://gitlab.com/lyoneel/dirly) for directory and file operations
- Go standard library for OS detection and environment lookup

## Extensibility

- Wrap `AppConfig` to add application-specific settings on top of the resolved directories.
- Compose `ConfigDir` with extension filters: `NewConfigDir(basePath, extensions...)` builds a config directory that only surfaces and accepts matching files (see [Extension Filtering](#extension-filtering)).
- Add new builder options by following the existing pattern: set a field, keep the builder chainable, apply the value in `Build()`.

## Testing

```bash
go test ./... -cover
```

The suite covers OS-specific resolution, environment variable overrides (including rejection of empty and relative values), both macOS strategies, builder validation and immutability, extension filtering, and the ConfigDir convenience methods. Run `go test -race -cover ./...` to see current coverage.

## License

MIT License - See [LICENSE](LICENSE) file for details.
