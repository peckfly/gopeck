package middleware

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/internal/pkg/errors"
)

var ErrCasbinDenied = errors.Unauthorized("com.casbin.denied", "Permission denied")

type CasbinConfig struct {
	AllowedPathPrefixes []string
	SkippedPathPrefixes []string
	Skipper             func(c *gin.Context) bool
	GetEnforcer         func(c *gin.Context) *casbin.Enforcer
	GetSubjects         func(c *gin.Context) []string
}

func CasbinWithConfig(config CasbinConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) ||
			(config.Skipper != nil && config.Skipper(c)) {
			c.Next()
			return
		}

		enforcer := config.GetEnforcer(c)
		if enforcer == nil {
			common.ResError(c, ErrCasbinDenied)
			return
		}

		for _, sub := range config.GetSubjects(c) {
			if b, err := enforcer.Enforce(sub, c.Request.URL.Path, c.Request.Method); err != nil {
				common.ResError(c, err)
				return
			} else if b {
				c.Next()
				return
			}
		}
		common.ResError(c, ErrCasbinDenied)
	}
}
