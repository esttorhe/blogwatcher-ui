// Package embed provides embedded filesystems for templates and static assets
package embed

import "embed"

//go:embed ../static/*
var StaticFS embed.FS

//go:embed ../templates/*.gohtml ../templates/pages/*.gohtml ../templates/partials/*.gohtml
var TemplateFS embed.FS
