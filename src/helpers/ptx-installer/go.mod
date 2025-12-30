module portunix.ai/portunix/src/helpers/ptx-installer

go 1.24.0

toolchain go1.24.2

// Use parent module for dependencies
replace portunix.ai/portunix => ../../..

require (
	github.com/spf13/cobra v1.10.1
	portunix.ai/portunix v0.0.0-00010101000000-000000000000
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
)
