package common

import "github.com/peckfly/gopeck/internal/pkg/errors"

const (
	ReqBodyKey                  = "req-body"
	ResBodyKey                  = "res-body"
	TreePathDelimiter           = "."
	ASC               Direction = "ASC"
	DESC              Direction = "DESC"
)

type Direction string

type ResponseResult struct {
	Success bool          `json:"success"`
	Data    interface{}   `json:"data,omitempty"`
	Total   int64         `json:"total,omitempty"`
	Error   *errors.Error `json:"error,omitempty"`
}

type PaginationResult struct {
	Total    int64 `json:"total"`
	Current  int   `json:"current"`
	PageSize int   `json:"pageSize"`
}
type PaginationParam struct {
	Pagination bool `form:"-"`
	OnlyCount  bool `form:"-"`
	Current    int  `form:"page"`
	PageSize   int  `form:"pageSize" binding:"max=100"`
}

type QueryOptions struct {
	SelectFields []string
	OmitFields   []string
	OrderFields  OrderByParams
}

type OrderByParam struct {
	Field     string
	Direction Direction
}

type OrderByParams []OrderByParam

func (a OrderByParams) ToSQL() string {
	if len(a) == 0 {
		return ""
	}

	var sql string
	for _, v := range a {
		sql += v.Field + " " + string(v.Direction) + ","
	}
	return sql[:len(sql)-1]
}
