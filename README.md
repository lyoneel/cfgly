# cfgly - Cross-Platform Application Configuration with OS-Specific Directory Resolution

A standalone Go library for cross-platform application configuration, providing OS-specific directory resolution following platform conventions.

> Main repository: https://gitlab.com/lyoneel/cfgly
> Any other host that serves this repository is a mirror. Open issues
> and merge requests on GitLab.

## Features

- **Cross-platform** directory resolution (Linux XDG, macOS Apple/XDG, Windows APPDATA)
- **Builder pattern** producing immutable configuration values
- **Configurable macOS strategy** — choose between traditional Apple directories and XDG-style paths
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
| macOS (XDG strategy) | `~/.config/myapp` | `~/.local/share/myapp` |
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
| `WithName(name string)` | Set the application name (required). Cannot be empty. |
| `WithConfigDir(dir string)` | Override the configuration directory base path. The app name is appended automatically. |
| `WithDataDir(dir string)` | Override the data directory base path. The app name is appended automatically. |
| `WithConfigExtensions(extensions ...string)` | Restrict the configuration directory to files with these extensions (case-insensitive, leading dot optional). Each call replaces the previous set; no arguments removes the filter. The data directory is never filtered. See [Extension Filtering](#extension-filtering). |
| `WithAutoCreate(create bool)` | Create config and data directories during `Build()`. Defaults to false. |
| `Build() (*AppConfig, error)` | Resolve OS-specific directories and create the final instance. Returns an error if directory resolution fails (e.g., home directory inaccessible). |

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
| `MacOSStrategyApple` (default) | `~/Library/Preferences/appName` | `~/Library/Application Support/appName` |
| `MacOSStrategyXDG` | `~/.config/appName` | `~/.local/share/appName` |

### Environment Variable Overrides

All OS-specific defaults can be overridden via environment variables. A value is only used when it is a non-empty absolute path for that platform (`/...` on Linux and macOS; `C:\...` or a UNC path such as `\\server\share` on Windows). Empty or relative values are ignored and resolution falls back as if the variable were unset, as the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/latest/) requires.

| Variable | Description |
|----------|-------------|
| `XDG_CONFIG_HOME` | Override Linux/macOS XDG config base directory |
| `XDG_DATA_HOME` | Override Linux/macOS XDG data base directory |
| `APPDATA` | Override Windows config base directory |
| `LOCALAPPDATA` | Override Windows data base directory |
| `USERPROFILE` | Fallback for Windows directories |
| `OSDIR_MACOS_APPLE_CONFIG_HOME` | Override macOS Apple config base directory |
| `OSDIR_MACOS_APPLE_DATA_HOME` | Override macOS Apple data base directory |

## Error Handling

`Build()` returns an error when the home directory is inaccessible, when the app name is empty, or when automatic directory creation fails. The error wraps the underlying cause with `fmt.Errorf` and `%w`, so callers can inspect it with `errors.Is` and `errors.As`. Read and write operations return the errors from the underlying dirly operations unchanged, except that on a filtered `ConfigDir` a disallowed file name returns an error wrapping `ErrExtensionNotAllowed`.

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
