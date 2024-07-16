package stringx

import (
	"github.com/google/uuid"
	"strings"
)

func NewUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
