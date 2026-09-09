[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/elfranne/check-haproxy-endpoint)
![Go Test](https://github.com/elfranne/check-haproxy-endpoint/actions/workflows/test.yml/badge.svg)
![Go Lint](https://github.com/elfranne/check-haproxy-endpoint/actions/workflows/lint.yml/badge.svg)
![goreleaser](https://github.com/elfranne/check-haproxy-endpoint/actions/workflows/release.yml/badge.svg)

# check-haproxy-endpoint

## Table of Contents
- [Overview](#overview)
- [Files](#files)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Check definition](#check-definition)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
- [Contributing](#contributing)

## Overview

check-haproxy-endpoint is a [Sensu Check][1] built with the [Sensu Plugin SDK][2] that reports
on the health of an HAProxy instance's backends and servers. Rather than scraping HAProxy's HTTP
stats page, it connects directly to the HAProxy stats socket (the `unix://` (or `tcp://`) address
configured with HAProxy's `stats socket` directive) and reads the CSV stats HAProxy returns over
it.

By default the check evaluates every backend and every server behind it in one pass:

- Any backend reported as `DOWN` makes the check **CRITICAL**.
- Any server reported as `DOWN` makes the check **WARNING** (unless a backend is also down, in
  which case CRITICAL wins).

It can also be scoped to only backends or only servers, narrowed to a single named backend or
server, asked to list what it currently sees (handy when writing a check definition), or asked to
fail if a specific backend/server it expects is missing entirely.

## Files

This project builds a single, self-contained Go binary; there are no other runtime files.

| File | Description |
|------|-------------|
| `main.go` | The check itself: registers the plugin's flags with the Sensu Plugin SDK, connects to the HAProxy stats socket, and evaluates backend/server status. |
| `main_test.go` | Unit tests covering the check logic. |

## Usage examples

### Flags

Every flag below can also be set via its environment variable, or via a Sensu annotation in the
`sensu.io/plugins/check-haproxy-endpoint/config` keyspace (the annotation key is the flag name
without the leading `--`, e.g. `socket`).

| Flag | Short | Environment variable | Default | Description |
|------|-------|-----------------------|---------|-------------|
| `--socket` | | `HAPROXY_SOCKET` | `unix:///var/run/haproxy.sock` | Socket to query for HAProxy stats. |
| `--backends` | `-B` | `HAPROXY_BACKENDS` | `false` | Check only backends. |
| `--backend` | `-b` | `HAPROXY_BACKEND` | `""` | Check only the specified backend. |
| `--servers` | `-S` | `HAPROXY_SERVERS` | `false` | Check only servers. |
| `--server` | `-s` | `HAPROXY_SERVER` | `""` | Check only the specified server. |
| `--check-missing` | `-m` | `HAPROXY_CHECK-MISSING` | `[]` | Combined with `--backends` or `--servers`, fail if one of the named entries given here is not present in HAProxy's stats. Repeatable. |
| `--list` | `-l` | `HAPROXY_LIST` | `false` | Combined with `--backends` or `--servers`, list every entry instead of evaluating status; useful for debugging, or for generating the list to pass to `--check-missing`. |

`--backends` and `--servers` are mutually exclusive; passing both returns **UNKNOWN**.

The check follows standard Sensu/Nagios exit codes: `0` OK, `1` WARNING, `2` CRITICAL, `3` UNKNOWN.

### Examples

Check every backend and server on the default socket:

```
$ check-haproxy-endpoint
Haproxy at unix:///var/run/haproxy.sock: all systems UP
```

If a backend is down, the check goes CRITICAL:

```
$ check-haproxy-endpoint
Service web_backend is DOWN!
```

If a server behind a backend is down (but the backend itself still has capacity), the check goes
WARNING:

```
$ check-haproxy-endpoint
Backend web01 for service web_backend is DOWN!
```

Point at a non-default socket:

```
$ check-haproxy-endpoint --socket unix:///var/lib/haproxy/stats.sock
Haproxy at unix:///var/lib/haproxy/stats.sock: all systems UP
```

Check only a single, named backend:

```
$ check-haproxy-endpoint --backends --backend web_backend
Service web_backend is UP
```

Check only a single, named server:

```
$ check-haproxy-endpoint --servers --server web01
Server web01 is UP:
```

List every backend currently known to HAProxy (no status is evaluated, exit is always OK):

```
$ check-haproxy-endpoint --backends --list
web_backend
api_backend
```

Fail if a specific backend is missing from HAProxy entirely (e.g. a config rollout dropped it),
in addition to listing what is actually present:

```
$ check-haproxy-endpoint --backends --check-missing web_backend --check-missing api_backend
Missing:
  []string{
- 	"web_backend",
  	"api_backend",
+ 	"other_backend",
  }
```

Passing both `--backends` and `--servers` is rejected before HAProxy is even queried:

```
$ check-haproxy-endpoint --backends --servers
--backends and --servers are mutually exclusive
```

A socket that can't be reached is reported as CRITICAL rather than UNKNOWN, since it usually means
HAProxy itself is down:

```
$ check-haproxy-endpoint --socket unix:///var/run/does-not-exist.sock
could not connect to socket: dial unix /var/run/does-not-exist.sock: connect: no such file or directory
```

## Configuration

### Asset registration

[Sensu Assets][3] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```
sensuctl asset add elfranne/check-haproxy-endpoint
```

If you're using an earlier version of sensuctl, you can find the asset on the
[Bonsai Asset Index][4].

### Check definition

```yml
---
type: CheckConfig
api_version: core/v2
metadata:
  name: check-haproxy-endpoint
  namespace: default
spec:
  command: check-haproxy-endpoint --backends --check-missing web_backend --check-missing api_backend
  subscriptions:
  - system
  runtime_assets:
  - elfranne/check-haproxy-endpoint
  interval: 30
  timeout: 10
```

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable script from this source.

From the local path of the check-haproxy-endpoint repository:

```
go build
```

## Additional notes

- The check talks to HAProxy over its stats socket, not over HTTP, so HAProxy must have a
  `stats socket` directive configured (e.g. `stats socket /var/run/haproxy.sock mode 660 level
  admin`) and the user running the check needs permission to connect to it — typically this means
  adding the Sensu agent's user to the socket's group, or adjusting the socket's `mode`/`level`.
- A `DOWN` backend and a `DOWN` server are treated differently: a down backend means the service
  behind it has no capacity left and is CRITICAL, while a down server behind an otherwise-healthy
  backend is only WARNING.
- `--list` and `--check-missing` never report a plain OK/backend-status line — they always either
  print the requested listing, print a diff of what's missing, or (on success) exit silently with
  status OK.
- `--check-missing` is most useful paired with `--list`: run once with `--list` to get the current
  set of backend or server names, then feed that list back in via `--check-missing` so future
  config changes that silently drop an entry are caught.

## Contributing

For more information about contributing to this plugin, see [Contributing][5].

[1]: https://docs.sensu.io/sensu-go/latest/reference/checks/
[2]: https://github.com/sensu/sensu-plugin-sdk
[3]: https://docs.sensu.io/sensu-go/latest/reference/assets/
[4]: https://bonsai.sensu.io/assets/elfranne/check-haproxy-endpoint
[5]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
