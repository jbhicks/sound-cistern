package public

import (
	"embed"
	"io/fs"

	"github.com/gobuffalo/buffalo"
)

//go:embed css images js favicon.ico favicon.svg robots.txt sitemap.xml
var files embed.FS

func FS() fs.FS {
	return buffalo.NewFS(files, "public")
}
