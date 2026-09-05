# Contributing to cfgly

Thank you for considering a contribution. This document explains how to set up the development environment, follow the project conventions, and submit changes.

> Main repository: https://gitlab.com/lyoneel/cfgly
> Any other host that serves this repository is a mirror. Open issues
> and merge requests on GitLab.

## Code of Conduct

This project keeps collaboration practical and respectful. Treat maintainers and contributors as professional peers.

## Development Environment Setup

1. Install Go 1.26 or later. Check with `go version`.
2. Clone the repository. The module depends on dirly, its sibling library:

```bash
git clone https://gitlab.com/lyoneel/cfgly.git
cd cfgly
```

3. Resolve the dependency. After the public release of dirly, `go mod tidy` fetches it from the module proxy. For local development against a sibling checkout, add a replace directive:

```bash
go mod edit -replace gitlab.com/lyoneel/dirly=../dirly
```

4. Confirm the toolchain works:

```bash
go build ./...
go test ./...
```

## Project Layout

| Path | Content |
|------|---------|
| `appcfg.go` | The `AppConfig` type and its accessors |
| `builder.go` | The `ConfigBuilder` fluent construction API |
| `config_dir.go` | The `ConfigDir` wrapper around `dirly.Directory` |
| `aux_fn.go` | OS-specific base directory resolution and env var overrides |
| `*_test.go` | Resolution, override, builder, and edge-case tests |

## Code Style Guidelines

1. Follow the conventions of Effective Go and the Go Code Review Comments guide.
2. Run `gofmt` on every changed file. The project has no custom formatter configuration.
3. Keep the public API stable within a major version. Breaking changes require a major version bump and a CHANGELOG entry.
4. Document every exported symbol with a doc comment that starts with the symbol name.
5. Wrap errors with `%w` so callers can use `errors.Is` and `errors.As`.
6. Keep the resolved paths immutable after `Build()`. Re-resolution requires a new builder.

## Testing Requirements

1. Run the full suite before you submit anything:

```bash
go test ./... -cover
```

2. New features need table-driven tests in the style of the existing `aux_fn_test.go` and `edge_cases_test.go` files.
3. Bug fixes need a regression test that fails without the fix.
4. Keep coverage at or above the current 75 percent. The CI pipeline runs vet, tests with coverage, and a build on every merge request.
5. Test fixtures use generic values only, for example the `C:\Users\test` Windows user. Never commit real personal paths.

## Git Workflow

1. The default branch is the integration target. Create a feature branch from it:

```bash
git checkout -b feat/my-change
```

2. Use conventional commit messages in the form `type: message` or `type!: message` for breaking changes. Known types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`, `ci`, `perf`.

```bash
git commit -m "feat: add WithXDGStrict builder option"
```

3. Keep each commit focused. One logical change per commit.
4. Rebase your branch on the default branch before you open a merge request.

## Pull Request Process

1. Push your branch to the same repository and open a merge request on GitLab.
2. Fill in the merge request template: describe the change, the type, and how you tested it.
3. The CI pipeline must pass: vet, tests with coverage, build.
4. Update documentation in the same change set:
   - README.md for user-facing behavior
   - DEVELOPMENT.md for resolution internals and extension points
   - CHANGELOG.md under an Unreleased heading

## Code Review Expectations

1. A maintainer reviews every merge request. Expect a review within a few days.
2. Reviewers check correctness, test coverage, API stability, and documentation accuracy.
3. Address review comments with new commits; keep the merge request history readable.
4. Squash only at the discretion of the maintainer who merges.

## Onboarding for New Contributors

1. Read the README quick start, then DEVELOPMENT.md for the resolution internals.
2. Pick an issue labeled with the bug label as a first change.
3. Build and run the suite first; the tests document the expected resolution behavior per platform.
4. Ask questions in the merge request or in an issue. Questions in the open are welcome.

## License

The project uses the MIT license. All contributions are submitted under the MIT license. See the LICENSE file for the full text.
