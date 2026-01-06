package templates

import "embed"

//go:embed examples/*
var TemplateFS embed.FS
