package biz

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGeneratePlanId(t *testing.T) {
	pId := uint64(0)
	for i := 0; i < 10000; i++ {
		id, _ := generatePlanId()
		assert.True(t, id > pId)
		pId = id
	}
}
