package assets

import "embed"

//go:embed *.css *.ico *.html *.png *.svg
var Assets embed.FS
