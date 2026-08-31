package web

import (
	_ "embed"
)

// IndexHTML contains the embedded dashboard HTML.
//go:embed index.html
var IndexHTML []byte