package templates

import "embed"

//go:embed qfd/*
var QFDTemplates embed.FS

//go:embed email/*
var EmailTemplates embed.FS

//go:embed fider/*
var FiderTemplates embed.FS
