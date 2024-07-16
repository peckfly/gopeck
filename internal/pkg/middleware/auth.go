package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/pkg/log"
)

type AuthConfig struct {
	AllowedPathPrefixes []string
	SkippedPathPrefixes []string
	RootID              string
	Skipper             func(c *gin.Context) bool
	ParseUserID         func(c *gin.Context) (string, error)
}

func AuthWithConfig(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) ||
			(config.Skipper != nil && config.Skipper(c)) {
			c.Next()
			return
		}

		userID, err := config.ParseUserID(c)
		if err != nil {
			common.ResError(c, err)
			return
		}

		ctx := common.NewUserID(c.Request.Context(), userID)
		ctx = log.NewUserID(ctx, userID)
		if userID == config.RootID {
			ctx = common.NewIsRootUser(ctx)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
