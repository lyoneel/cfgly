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
| `config_dir.go` | The `ConfigDir` wrapper around `dirly.Directory` |
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

1. If `WithConfigDir` or `WithDataDir` set an explicit base, that base wins and the app name is appended with `filepath.Join`.
2. Otherwise the base comes from `getConfigBaseForOS(runtime.GOOS)` or `getDataBaseForOS(runtime.GOOS)`.
3. The OS dispatch returns:
   - Linux: `$XDG_CONFIG_HOME` or `~/.config` for config; `$XDG_DATA_HOME` or `~/.local/share` for data.
   - macOS with the Apple strategy: `~/Library/Preferences` for config; `~/Library/Application Support` for data. The XDG strategy uses the Linux paths.
   - Windows: `%APPDATA%` for config; `%LOCALAPPDATA%` for data, with `USERPROFILE` as the fallback.
4. A missing home directory is an error; `Build()` wraps it and fails.
5. With `WithAutoCreate(true)`, both directories are created with `0o755` permissions via `os.MkdirAll`.

Environment variable overrides, highest precedence first:

| Variable | Affects |
|----------|---------|
| `OSDIR_MACOS_APPLE_CONFIG_HOME` | macOS Apple config base |
| `OSDIR_MACOS_APPLE_DATA_HOME` | macOS Apple data base |
| `XDG_CONFIG_HOME` | Linux and macOS XDG config base |
| `XDG_DATA_HOME` | Linux and macOS XDG data base |
| `APPDATA` | Windows config base |
| `LOCALAPPDATA` | Windows data base |
| `USERPROFILE` | Windows fallback for both bases |

The resolved paths are fixed after `Build()`. Re-resolving requires a new builder.

## Builder Flow

```go
b := cfgly.NewAppConfigBuilder()
b = b.WithName("myapp")        // required; empty name sets the builder error
b = b.WithConfigDir("/custom") // optional base override
b = b.WithDataDir("/custom")   // optional base override
b = b.WithAutoCreate(true)     // optional directory creation
cfg, err := b.Build()          // resolves, optionally creates, returns *AppConfig
```

The builder carries an internal error: the first failing step wins, and later steps keep the chainable shape. `Build()` returns that error.

## Extension Points

- Wrap `AppConfig` for application-level settings.
- Use `NewConfigDir(basePath, extensions...)` to build a config directory that filters by extension through dirly.
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
