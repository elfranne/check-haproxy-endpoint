# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

## [0.4.0] - 2026-09-09

### Changed
- Migrated `.goreleaser.yml` to the GoReleaser v2 config schema.
- Hardened the GitHub Actions workflows: tightened permissions to least-privilege,
  added pull request triggers, added concurrency groups to cancel superseded runs,
  and added `go vet`/`go build` steps to the test workflow.
- Rewrote `README.md` and `CHANGELOG.md` to remove leftover template boilerplate
  and reflect the plugin's actual flags, behavior, and release history.
- Updated the Go toolchain and Go module dependencies, and bumped the GitHub
  Actions used by the workflows to their current major versions. This includes
  `google.golang.org/grpc` 1.83.2, which resolves a high-severity HTTP/2 heap
  memory exhaustion (OOM) advisory affecting versions <= 1.83.0.

### Added
- Real unit tests in `main_test.go`, replacing the placeholder `TestMain` stub.
- A `.golangci.yml` config, required for golangci-lint v2 to run at all in CI.

### Fixed
- A server reported as `DOWN` now exits WARNING instead of CRITICAL, so server-level
  and backend-level failures are no longer indistinguishable to Sensu.
- Removed a trailing space from the plugin name, which rendered as a double space in
  `--help` output and reported the plugin under a name matching neither the repository
  nor the Bonsai asset. The config keyspace was already correct, so annotation-based
  configuration is unaffected.

## [0.3.2] - 2025-02-24

### Changed
- Updated the Go toolchain version and Go module dependencies.

## [0.3.1] - 2025-01-13

### Changed
- Updated the Go toolchain version and Go module dependencies.

## [0.3] - 2024-09-16

### Changed
- Updated the Go toolchain version, Go module dependencies, and GitHub Actions
  workflows.
- Bumped `golangci/golangci-lint-action` from 5 to 6.
- Bumped `github.com/google/go-cmp` from 0.5.9 to 0.6.0.

## [0.2.3] - 2024-04-30

### Changed
- Improved the OK output: checking a single `--backend` or `--server` now prints
  an explicit `... is UP` message instead of staying silent.

## [0.2.2] - 2024-04-30

### Fixed
- `--check-missing` and `--list` no longer print the generic "all systems UP"
  status line; they only print their own listing/diff output.

## [0.2.1] - 2024-04-30

### Fixed
- Backend/server DOWN checks no longer fire while `--check-missing` is in use,
  so the two modes don't produce conflicting status output.

## [0.2.0] - 2024-04-30

### Added
- `--backends`/`-B` and `--servers`/`-S` to restrict a check to only backends or
  only servers (mutually exclusive; using both returns UNKNOWN).
- `--backend`/`-b` and `--server`/`-s` to scope a check to a single named backend
  or server.
- `--check-missing`/`-m` to fail the check if an expected backend or server name
  is not present in HAProxy's stats.
- `--list`/`-l` to print every known backend or server name instead of
  evaluating status, for debugging or for building a `--check-missing` list.

### Changed
- Bumped `golangci/golangci-lint-action` from 4 to 5.

## [0.1.0] - 2024-04-26

### Changed
- Rewrote the check to read HAProxy stats from the local HAProxy stats socket
  (via the `ruansteve/go-haproxy` module) instead of scraping HAProxy's HTTP
  stats page. This replaced the `--url`/`--admin-user`/`--admin-pass` flags with
  a single `--socket` flag (`HAPROXY_SOCKET`, default
  `unix:///var/run/haproxy.sock`).

[Unreleased]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.4.0...HEAD
[0.4.0]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.3.2...0.4.0
[0.3.2]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.3.1...0.3.2
[0.3.1]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.3...0.3.1
[0.3]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.2.3...0.3
[0.2.3]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.2.2...0.2.3
[0.2.2]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.2.1...0.2.2
[0.2.1]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.2.0...0.2.1
[0.2.0]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.1.0...0.2.0
[0.1.0]: https://github.com/elfranne/check-haproxy-endpoint/releases/tag/0.1.0
