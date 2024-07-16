package middleware

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type StaticConfig struct {
	SkippedPathPrefixes []string
	Root                string
}

func StaticWithConfig(config StaticConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if SkippedPathPrefixes(c, config.SkippedPathPrefixes...) {
			c.Next()
			return
		}

		p := c.Request.URL.Path
		fPath := filepath.Join(config.Root, filepath.FromSlash(p))
		_, err := os.Stat(fPath)
		if err != nil && os.IsNotExist(err) {
			fPath = filepath.Join(config.Root, "index.html")
		}
		c.File(fPath)
		c.Abort()
	}
}
