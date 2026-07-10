// Package myfactorycore embeds the reusable MyFactory templates into the
// binary so `myfactory init` works on machines that have no source checkout.
//
// This file must live at the repository root because go:embed cannot
// reference parent directories, and the templates/ tree is shared with the
// legacy Python CLI and the docs.
//
// The `all:` prefix is required: without it, embed skips files and
// directories whose names start with "." or "_", which would silently drop
// .ApplicationFactory, .github, .gitlab, and .claude template content.
// internal/assets/assets_test.go verifies the hidden files are present.
package myfactorycore

import "embed"

//go:embed all:templates/product-repo all:templates/project-overlays
var TemplatesFS embed.FS
