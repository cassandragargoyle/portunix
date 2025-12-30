module portunix.ai/portunix/src/helpers/ptx-mcp

go 1.24.0

toolchain go1.24.2

// Use parent module's app submodule for MCP server
replace portunix.ai/app => ../../app

replace portunix.ai/app/install => ../../app/install

require (
	github.com/spf13/cobra v1.8.1
	portunix.ai/app v0.0.0
	portunix.ai/app/install v0.0.0
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58 // indirect
	github.com/pkg/sftp v1.13.10 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/testcontainers/testcontainers-go v0.40.0 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/term v0.38.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
