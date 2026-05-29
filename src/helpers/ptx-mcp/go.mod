module portunix.ai/portunix/src/helpers/ptx-mcp

go 1.24.0

toolchain go1.24.2

// Use parent module's app submodule (still needed by other ptx-mcp files).
replace portunix.ai/app => ../../app

require (
	github.com/spf13/cobra v1.10.1
	portunix.ai/app v0.0.0
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/term v0.39.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
