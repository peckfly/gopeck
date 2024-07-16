package middleware

import (
	"github.com/gin-contrib/cors"
	"time"

	"github.com/gin-gonic/gin"
)

type CORSConfig struct {
	Enable          bool
	AllowAllOrigins bool
	// AllowOrigins is a list of origins a cross-domain request can be executed from.
	// If the special "*" value is present in the list, all origins will be allowed.
	// Default value is []
	AllowOrigins []string
	// AllowMethods is a list of methods the client is allowed to use with
	// cross-domain requests. Default value is simple methods (GET, POST, PUT, PATCH, DELETE, HEAD, and OPTIONS)
	AllowMethods []string
	// AllowHeaders is list of non simple headers the client is allowed to use with
	// cross-domain requests.
	AllowHeaders []string
	// AllowCredentials indicates whether the request can include user credentials like
	// cookies, HTTP authentication or client side SSL certificates.
	AllowCredentials bool
	// ExposeHeaders indicates which headers are safe to expose to the API of a CORS
	// API specification
	ExposeHeaders []string
	// MaxAge indicates how long (with second-precision) the results of a preflight request
	// can be cached
	MaxAge int
	// Allows to add origins like http://some-domain/*, https://api.* or http://some.*.subdomain.com
	AllowWildcard bool
	// Allows usage of popular browser extensions schemas
	AllowBrowserExtensions bool
	// Allows usage of WebSocket protocol
	AllowWebSockets bool
	// Allows usage of file:// schema (dangerous!) use it only when you 100% sure it's needed
	AllowFiles bool
}

var defaultCORSConfig = CORSConfig{
	Enable:           true,
	AllowAllOrigins:  false,
	AllowOrigins:     []string{"*"},
	AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
	AllowHeaders:     []string{"*"},
	AllowCredentials: true,
	ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Cache-Control", "Content-Language", "Content-Type"},
	MaxAge:           86400,
	AllowWildcard:    true,
	AllowWebSockets:  true,
	AllowFiles:       true,
}

func Cors() gin.HandlerFunc {
	return CORSWithConfig(defaultCORSConfig)
}

func CORSWithConfig(cfg CORSConfig) gin.HandlerFunc {
	if !cfg.Enable {
		return Empty()
	}
	return cors.New(cors.Config{
		AllowAllOrigins:        cfg.AllowAllOrigins,
		AllowOrigins:           cfg.AllowOrigins,
		AllowMethods:           cfg.AllowMethods,
		AllowHeaders:           cfg.AllowHeaders,
		AllowCredentials:       cfg.AllowCredentials,
		ExposeHeaders:          cfg.ExposeHeaders,
		MaxAge:                 time.Second * time.Duration(cfg.MaxAge),
		AllowWildcard:          cfg.AllowWildcard,
		AllowBrowserExtensions: cfg.AllowBrowserExtensions,
		AllowWebSockets:        cfg.AllowWebSockets,
		AllowFiles:             cfg.AllowFiles,
	})
}
