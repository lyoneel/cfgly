# cfgly Developer Reference

Developer guide for the cfgly package: architecture, resolution internals, builder flow, and extension guidance. The README holds the overview and quick start; this document holds the depth.

> Main repository: https://gitlab.com/lyoneel/cfgly
> Any other host that serves this repository is a mirror. Open issues
> and merge requests on GitLab.

## Architecture

The package is a single root package with these source files:

| File | Responsibility |
|------|----------------|
| `appcfg.go` | The `AppConfig` type: name plus resolved config and data `ConfigDir` instances |
| `builder.go` | The `ConfigBuilder` fluent API, OS dispatch, and automatic directory creation |
| `name.go` | App name validation (single plain path segment, Windows reserved names) |
| `errors.go` | Exported sentinel errors returned by `Build()` |
| `config_dir.go` | The `ConfigDir` wrapper around `dirly.Directory`: extension filter and permission defaults |
| `aux_fn.go` | OS-specific base directory resolution and environment variable overrides |
| `*_test.go` | Resolution, override, builder, and edge-case test suites |

Design patterns:

- Builder pattern for construction. `NewAppConfigBuilder` returns a chainable builder; `Build()` produces an immutable `AppConfig`.
- Instance-based API with no global state. Every instance carries its own resolved paths.
- Embedding: `ConfigDir` embeds `*dirly.Directory`, so the full dirly operation set is available on config and data directories.

Dependency model:

- cfgly depends on dirly for all directory operations. cfgly owns only the OS-specific resolution logic.

## Resolution Internals

`Build()` resolves paths in this order:

0. The builder error, if any, is returned first; then the app name is validated (`validateAppName`).
1. If `WithConfigDir` or `WithDataDir` set an explicit absolute base, that base wins. The setters store only the base; `Build()` appends the app name with `filepath.Join`, so call order does not matter.
2. Otherwise the base comes from `getConfigBaseForOS(goos, strategy)` or `getDataBaseForOS(goos, strategy)`, where `strategy` is the `WithMacOSStrategy` value (default `MacOSStrategyApple`).
3. The OS dispatch returns:
   - Linux: `$XDG_CONFIG_HOME` or `~/.config` for config; `$XDG_DATA_HOME` or `~/.local/share` for data.
   - macOS with the Apple strategy: `~/Library/Preferences` for config; `~/Library/Application Support` for data. The XDG strategy uses the Linux resolution, including `XDG_CONFIG_HOME` and `XDG_DATA_HOME`.
   - Windows: `%APPDATA%` for config; `%LOCALAPPDATA%` for data, with `USERPROFILE` as the fallback.
4. A missing home directory is an error; `Build()` wraps it and fails.
5. With `WithAutoCreate(true)`, both directories are created via `os.MkdirAll` with the `WithDirPerm` mode (default `0o700`). Existing directories are never chmod-ed.
6. The returned `ConfigDir`s write new files with the `WithFilePerm` mode (default `0o600`) when the caller passes perm `0`, and create missing parent directories with the directory mode before dirly writes.

Environment variable overrides, highest precedence first. Each is used only when it holds a non-empty absolute path for the platform; empty or relative values are ignored as if unset:

| Variable | Affects |
|----------|---------|
| `OSDIR_MACOS_APPLE_CONFIG_HOME` | macOS Apple config base |
| `OSDIR_MACOS_APPLE_DATA_HOME` | macOS Apple data base |
| `XDG_CONFIG_HOME` | Linux and macOS XDG config base |
| `XDG_DATA_HOME` | Linux and macOS XDG data base |
| `APPDATA` | Windows config base |
| `LOCALAPPDATA` | Windows data base |
| `USERPROFILE` | Windows fallback for both bases |

The resolved paths are fixed after `Build()`. Every `Build()` call resolves paths again and returns a new, independent `AppConfig`; the builder is never modified by `Build()`. The builder is not safe for concurrent use.

## Builder Flow

```go
b := cfgly.NewAppConfigBuilder()
b = b.WithName("myapp")                         // required; invalid names set the builder error
b = b.WithConfigDir("/custom")                  // optional absolute base override (any order)
b = b.WithDataDir("/custom")                    // optional absolute base override (any order)
b = b.WithMacOSStrategy(cfgly.MacOSStrategyXDG) // optional; Apple by default
b = b.WithDirPerm(0o700).WithFilePerm(0o600)    // optional; these are the defaults
b = b.WithAutoCreate(true)                      // optional directory creation
cfg, err := b.Build()                           // resolves, optionally creates, returns *AppConfig
```

The builder carries an internal error: the first failing step wins, and later steps keep the chainable shape. `Build()` returns that error.

## Extension Points

- Wrap `AppConfig` for application-level settings.
- Use `NewConfigDir(basePath, extensions...)` or `WithConfigExtensions(...)` to build a config directory that filters by extension. Listing, `Exists`, reads, and writes all honor the filter; disallowed names return `ErrExtensionNotAllowed`.
- Add builder options by setting a field on `ConfigBuilder`, returning the builder, and applying the value inside `Build()`. Follow the existing `WithAutoCreate` pattern.
- New platforms need a case in the OS dispatch functions in `aux_fn.go` plus table-driven tests in `aux_fn_test.go`.

## Testing

```bash
go test ./... -cover
```

Test layout:

- `aux_fn_test.go` covers OS dispatch and environment variable overrides, including the Windows variable chain.
- `builder_test.go` and `edge_cases_test.go` cover builder validation and failure paths.
- `config_dir_test.go` covers the ConfigDir convenience methods.

Windows paths in the fixtures use a generic `C:\Users\test` user and are test data only.
