package common

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/peckfly/gopeck/internal/pkg/errors"
	"github.com/peckfly/gopeck/pkg/log/logc"
	"go.uber.org/zap"
	"net/http"
	"reflect"
	"strings"
)

// GetToken Get access token from header or query parameter
func GetToken(c *gin.Context) string {
	var token string
	auth := c.GetHeader("Authorization")
	prefix := "Bearer "
	if auth != "" && strings.HasPrefix(auth, prefix) {
		token = auth[len(prefix):]
	} else {
		token = auth
	}
	if token == "" {
		token = c.Query("accessToken")
	}
	return token
}

// GetBodyData Get body data from context
func GetBodyData(c *gin.Context) []byte {
	if v, ok := c.Get(ReqBodyKey); ok {
		if b, ok := v.([]byte); ok {
			return b
		}
	}
	return nil
}

// ParseJSON Parse body json data to struct
func ParseJSON(c *gin.Context, obj any) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return errors.BadRequest("", "Failed to parse json: %s", err.Error())
	}
	return nil
}

// ParseQuery Parse query parameter to struct
func ParseQuery(c *gin.Context, obj any) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return errors.BadRequest("", "Failed to parse query: %s", err.Error())
	}
	return nil
}

// ParseForm Parse body form data to struct
func ParseForm(c *gin.Context, obj any) error {
	if err := c.ShouldBindWith(obj, binding.Form); err != nil {
		return errors.BadRequest("", "Failed to parse form: %s", err.Error())
	}
	return nil
}

// ResJSON Response json data with status code
func ResJSON(c *gin.Context, status int, v any) {
	buf, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	c.Set(ResBodyKey, buf)
	c.Data(status, "application/json; charset=utf-8", buf)
	c.Abort()
}

func ResSuccess(c *gin.Context, v any) {
	ResJSON(c, http.StatusOK, ResponseResult{
		Success: true,
		Data:    v,
	})
}

func ResOK(c *gin.Context) {
	ResJSON(c, http.StatusOK, ResponseResult{
		Success: true,
	})
}

func ResPage(c *gin.Context, v any, pr *PaginationResult) {
	var total int64
	if pr != nil {
		total = pr.Total
	}

	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	if reflectValue.IsNil() {
		v = make([]any, 0)
	}

	ResJSON(c, http.StatusOK, ResponseResult{
		Success: true,
		Data:    v,
		Total:   total,
	})
}

func ResError(c *gin.Context, err error, status ...int) {
	var err0 *errors.Error
	if e, ok := errors.As(err); ok {
		err0 = e
	} else {
		err0 = errors.FromError(errors.InternalServerError("", err.Error()))
	}
	code := int(err0.Code)
	if len(status) > 0 {
		code = status[0]
	}
	if code >= 500 {
		ctx := c.Request.Context()
		logc.Error(ctx, "Internal server error", zap.Error(err))
		err0.Detail = http.StatusText(http.StatusInternalServerError)
	}
	err0.Code = int32(code)
	ResJSON(c, code, ResponseResult{Error: err0})
}
