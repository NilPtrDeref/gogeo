package frontend

import (
	"embed"
)

//go:generate npm run build
//go:embed all:build
var Files embed.FS
