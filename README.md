# cfgly - Cross-Platform Application Configuration with OS-Specific Directory Resolution

A standalone Go library for cross-platform application configuration, providing OS-specific directory resolution following platform conventions.

> Main repository: https://gitlab.com/lyoneel/cfgly
> Any other host that serves this repository is a mirror. Open issues
> and merge requests on GitLab.

## Features

- **Cross-platform** directory resolution (Linux XDG, macOS Apple/XDG, Windows APPDATA)
- **Builder pattern** for immutable configuration
- **Configurable macOS strategy** — choose between traditional Apple directories and XDG-style paths
- **Environment variable overrides** — all OS-specific defaults can be overridden via environment variables
- **Instance-based API** with no global state

## Installation

```bash
go get gitlab.com/lyoneel/cfgly
```

```go
import "gitlab.com/lyoneel/cfgly"
```

Requires Go 1.26 or later. The library depends on [dirly](https://gitlab.com/lyoneel/dirly) for directory operations.

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

Directory resolution happens once, at build time, and never changes afterwards. All methods are safe for concurrent use because the resolved paths are immutable.

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
| `Exists(filename string) bool` | Check if a file exists in the directory. |

Because `ConfigDir` embeds `dirly.Directory`, every dirly operation is available directly, including batch reads and writes, glob search, and file management. See the [dirly documentation](https://gitlab.com/lyoneel/dirly) for the full operation set.

### macOS Strategy

macOS supports two directory resolution strategies:

| Strategy | Config Dir | Data Dir |
|----------|-----------|----------|
| `MacOSStrategyApple` (default) | `~/Library/Preferences/appName` | `~/Library/Application Support/appName` |
| `MacOSStrategyXDG` | `~/.config/appName` | `~/.local/share/appName` |

### Environment Variable Overrides

All OS-specific defaults can be overridden via environment variables:

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

`Build()` returns an error when the home directory is inaccessible, when the app name is empty, or when automatic directory creation fails. The error wraps the underlying cause with `fmt.Errorf` and `%w`, so callers can inspect it with `errors.Is` and `errors.As`. Read and write operations return the errors from the underlying dirly operations unchanged.

## Tech Stack and Dependencies

- Go 1.26 or later
- [dirly](https://gitlab.com/lyoneel/dirly) for directory and file operations
- Go standard library for OS detection and environment lookup

## Extensibility

- Wrap `AppConfig` to add application-specific settings on top of the resolved directories.
- Compose `ConfigDir` with dirly filters: `NewConfigDir(basePath, extensions...)` builds a config directory that only surfaces matching files.
- Add new builder options by following the existing pattern: set a field, keep the builder chainable, apply the value in `Build()`.

## Testing

```bash
go test ./... -cover
```

The suite covers OS-specific resolution, environment variable overrides, both macOS strategies, builder validation, and the ConfigDir convenience methods. Current coverage is 75.2 percent of statements.

## License

MIT License - See [LICENSE](LICENSE) file for details.
