#!/usr/bin/env bash
# Ensure the Delve debugger (dlv) is built with a Go toolchain new enough to
# read the DWARF debug info emitted by the toolchain that compiles Portunix.
#
# Why: Go 1.25+ emits DWARFv5. A dlv built with an older Go fails every debug
# session with "Delve must be built with Go version 1.25.0 or later". The Go
# version used to compile the program is defined by go.mod (the `go` directive)
# + GOTOOLCHAIN, NOT by which Go built dlv, so dlv is a separate concern that
# no project file can pin. This script closes that gap: it rebuilds dlv with
# the exact project toolchain, but only when the installed dlv is too old.
#
# Wired as a VS Code preLaunchTask (see .vscode/tasks.json). Fast no-op when
# dlv already matches; only touches the network when a rebuild is required.
set -euo pipefail

command -v go >/dev/null 2>&1 || { echo "ensure-delve: 'go' not on PATH" >&2; exit 1; }

# Toolchain that compiles the project (honours go.mod `go`/`toolchain` + GOTOOLCHAIN).
tool_go="$(go env GOVERSION)"                                   # e.g. go1.25.0

# Locate the real dlv in GOBIN (fallback: GOPATH/bin).
gobin="$(go env GOBIN)"; [ -n "$gobin" ] || gobin="$(go env GOPATH)/bin"
real_dlv="$gobin/dlv"

dlv_go=""
if [ -x "$real_dlv" ]; then
	dlv_go="$(go version -m "$real_dlv" 2>/dev/null | grep -oE 'go1\.[0-9]+(\.[0-9]+)?' | head -1)"
fi

# Rebuild unless dlv exists AND was built with Go >= the project toolchain.
need_build=1
if [ -n "$dlv_go" ]; then
	oldest="$(printf '%s\n%s\n' "$dlv_go" "$tool_go" | sort -V | head -1)"
	[ "$oldest" = "$tool_go" ] && need_build=0
fi

if [ "$need_build" -eq 0 ]; then
	echo "ensure-delve: dlv ($dlv_go) matches toolchain ($tool_go) — OK" >&2
	exit 0
fi

echo "ensure-delve: rebuilding dlv (was: ${dlv_go:-missing}) with $tool_go ..." >&2
# Force the exact project toolchain so dlv's DWARF support matches the program.
GOTOOLCHAIN="$tool_go" go install github.com/go-delve/delve/cmd/dlv@latest >&2
echo "ensure-delve: done" >&2
