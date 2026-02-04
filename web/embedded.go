package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var staticFiles embed.FS

func RegisterStaticRoutes(r *gin.Engine) {
	distFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		panic(err)
	}

	// Serve static assets
	r.StaticFS("/assets", http.FS(distFS))

	// Catch-all for SPA routing
	r.NoRoute(func(c *gin.Context) {
		// If requesting a file (e.g., .js, .css, .png), and it's not found in /assets,
		// it might be at the root of dist (like favicon.ico or index.html)
		path := c.Request.URL.Path
		if path == "/" {
			path = "index.html"
		} else {
			// Remove leading slash for FS lookup
			path = path[1:]
		}

		file, err := distFS.Open(path)
		if err == nil {
			file.Close()
			http.FileServer(http.FS(distFS)).ServeHTTP(c.Writer, c.Request)
			return
		}

		// Fallback to index.html for SPA
		c.FileFromFS("index.html", http.FS(distFS))
	})
}
