package main

import (
	"embed"
)

// Embed assets directory - path relative to this file location
//go:embed assets/packages/*.json
var embeddedAssets embed.FS

// Embed installation scripts for Windows
//go:embed assets/scripts/windows/*.ps1
//go:embed assets/scripts/windows/*.cmd
var embeddedScripts embed.FS
