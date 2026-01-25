package public

import (
	"embed"
)

//go:embed css images js favicon.ico favicon.svg robots.txt sitemap.xml
var FS embed.FS
