package web

import "embed"

//go:embed index.html app.js vendor styles
var FS embed.FS
