# pkg

## Overview

`pkg/` holds Go library packages of the root module `portunix.ai/portunix`. They are imported
by path, for example `portunix.ai/portunix/pkg/logging`. They have no `main` and produce no
binary.

The name follows the common Go layout convention: `pkg/` is library code meant to be importable,
in contrast to `src/` (the program code of `portunix` and its `ptx-*` helpers).

**Current state**: none of the packages is imported by the `portunix` binary or a helper today.
They are used by their own tests, examples and the integration tests in `test/integration/`.
Treat them as building blocks prepared for adoption, not as code on the runtime path.

## Packages

| Package | Purpose | Origin |
| ------- | ------- | ------ |
| [`logging`](logging/) | Structured logging on top of zerolog | Issue #52 |
| [`logging/handlers`](logging/handlers/) | Stand-alone console, file (rotation) and syslog writers | Issue #52 |
| [`logging/middleware`](logging/middleware/) | Correlation IDs for HTTP and context propagation | Issue #52 |
| [`docker`](docker/) | Interfaces and types for testing container operations | Testing guide |

### logging

Structured, levelled logging for Portunix components, built on
[zerolog](https://github.com/rs/zerolog).

- **`Logger`** interface and `New(component)` for a quick logger with default settings
- **`Config`**: level, format (`text` / `json`), outputs (`console`, `stderr`, `file`, `syslog`),
  file path, per-module levels; loaded from code, a file (`LoadFromFile`) or the environment
- **`Factory`** creates configured loggers; `CreateMCPLogger` writes to a file only, so MCP
  stdio stays clean; `CreateTestLogger` adapts to `go test -v`
- **`ConfigManager`** and the global functions (`GetLogger`, `SetGlobalLogLevel`,
  `SetModuleLogLevel`, `UpdateGlobalConfig`) change the configuration at runtime
- **Context helpers** (`context.go`): logger, correlation ID and user ID carried in
  `context.Context`, `InfoContext` and friends
- Every entry carries `component`, `os`, `arch` and, inside a container, `container` /
  `container_id`

Log files are opened once per path and shared by all loggers of a factory, including the
factories derived for MCP, tests and configuration changes. Call `Close()` on the factory, or
`CloseGlobalLoggers()` for the global manager, at process shutdown.

Environment variables read by `Config.LoadFromEnv`:

| Variable | Meaning |
| -------- | ------- |
| `PORTUNIX_LOG_LEVEL` | `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic` |
| `PORTUNIX_LOG_FORMAT` | `text` or `json` |
| `PORTUNIX_LOG_OUTPUT` | Comma-separated outputs, e.g. `console,file` |
| `PORTUNIX_LOG_FILE` | Log file path for the `file` output |
| `PORTUNIX_LOG_NO_COLOR` | `true` or `1` disables colors |
| `PORTUNIX_LOG_MODULE_<MODULE>` | Level for one module, e.g. `PORTUNIX_LOG_MODULE_MCP=debug` |

```go
import "portunix.ai/portunix/pkg/logging"

factory := logging.NewFactory(&logging.Config{
    Level:    "info",
    Format:   "json",
    Output:   []string{"console", "file"},
    FilePath: "/tmp/portunix.log",
})
defer factory.Close()

log := factory.CreateLogger("installer")
log.Info("package installed", "package", "nodejs")
```

More runnable examples are in [example_test.go](logging/example_test.go).

### logging/handlers

Writers that can be used independently of `Factory`: `ConsoleHandler`, `FileHandler` (size and
age based rotation, backups, compression) and `SyslogHandler` (local or remote; on Windows
`NewSyslogHandler` returns a "not supported" error and `IsSyslogSupported()` is `false`).
Each has a builder, e.g. `NewFileBuilder(path).WithMaxSizeMB(10)`.
`Factory` does not use them yet; it opens plain files itself.

### logging/middleware

Correlation IDs for tracing a request across components: `HTTPMiddleware` reads or creates the ID
from `X-Correlation-ID`, `X-Request-ID` or `X-Trace-ID` and stores it in the request context;
`CreateChildContext` and `CreateSpan` derive nested IDs for sub-operations. Generators: default,
UUID-like and timestamp based.

### docker

Only interfaces and data types (`DockerClient`, `SystemDetector`, `NetworkManager`, `FileSystem`,
`ContainerConfig`, `ContainerInfo`, `PackageManagerInfo`) that make container operations
mockable in tests. It imports `portunix.ai/app/system` for `SystemInfo`. Most of the package is
its tests:

- `docker_test.go` (`unit` tag): mock-based tests of the interfaces
- `docker_integration_test.go` (`integration` tag): runs real containers via testcontainers and
  skips when no Docker daemon is available

The runtime container support lives in the `ptx-container` helper
(`src/helpers/ptx-container/`), not here.

## Where Other Code Lives

| Directory | Contents |
| --------- | -------- |
| `src/app/`, `src/cmd/`, `src/parser/`, `src/dispatcher/` | Main program (modules mapped by `replace` in `go.mod`) |
| `src/helpers/ptx-*/` | Helper binaries invoked through the dispatcher |
| `src/pkg/` | Libraries used by the program itself (archive, installers, package registry, platform) |
| `src/shared/` | Small code shared by helpers (version, clipboard) |
| `internal/testutils/` | Test helpers, importable only inside this module |

Rule of thumb: code that the program or a helper needs goes to `src/`; code meant as a reusable
library of the root module goes to `pkg/`.

## Testing

```bash
# logging (no build tag)
go test ./pkg/logging/...

# docker, mock-based
go test -tags=unit ./pkg/docker/...

# docker against a real daemon, skipped without one
go test -tags=integration ./pkg/docker/...
```

The CI Unit Tests job runs `go test -tags=unit -race ./...`, which includes both packages.

## Adding a Package

- Put it under `pkg/<name>/` with a package doc comment (`// Package <name> ...`)
- Keep it free of dependencies on `src/` where possible, so it stays importable as a library
  (`logging` has none, `docker` depends on `portunix.ai/app/system`)
- Add tests and, for public APIs, `Example` functions
- Add a row to the table in [Packages](#packages)
