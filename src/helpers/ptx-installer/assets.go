package main

import (
	"embed"
)

// Embed assets directory - path relative to this file location
//go:embed assets/packages/*.json
var embeddedAssets embed.FS
