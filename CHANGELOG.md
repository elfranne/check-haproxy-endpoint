# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

### Changed
- Migrated `.goreleaser.yml` to the GoReleaser v2 config schema.
- Modernized the GitHub Actions workflows: bumped actions to their current major
  versions, tightened workflow permissions to least-privilege, added pull request
  triggers, and added concurrency groups to cancel superseded runs.
- Rewrote `README.md` and `CHANGELOG.md` to remove leftover template boilerplate
  and reflect the plugin's actual flags, behavior, and release history.

### Added
- Real unit tests in `main_test.go`, replacing the placeholder `TestMain` stub.

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

[Unreleased]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.3.2...HEAD
[0.3.2]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.3.1...0.3.2
[0.3.1]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.3...0.3.1
[0.3]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.2.3...0.3
[0.2.3]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.2.2...0.2.3
[0.2.2]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.2.1...0.2.2
[0.2.1]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.2.0...0.2.1
[0.2.0]: https://github.com/elfranne/check-haproxy-endpoint/compare/0.1.0...0.2.0
[0.1.0]: https://github.com/elfranne/check-haproxy-endpoint/releases/tag/0.1.0
