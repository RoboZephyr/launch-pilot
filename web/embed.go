package web

import "embed"

//go:embed index.html app.js vendor styles components lib
var FS embed.FS
