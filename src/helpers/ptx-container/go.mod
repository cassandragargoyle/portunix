module portunix.ai/portunix/src/helpers/ptx-container

go 1.24.0

toolchain go1.24.2

// Use parent module for dependencies
replace portunix.ai/portunix => ../../..

require (
	github.com/spf13/cobra v1.8.1
	golang.org/x/term v0.38.0
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.39.0 // indirect
)
