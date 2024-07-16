package common

import "github.com/rs/xid"

func NewXID() string {
	return xid.New().String()
}
