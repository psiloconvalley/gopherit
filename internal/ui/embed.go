package ui

import "embed"

// FS contains all static assets and templates.
//
//go:embed all:static all:templates
var FS embed.FS
