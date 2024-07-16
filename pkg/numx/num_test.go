package numx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCeilDiv(t *testing.T) {
	assert.Equal(t, 2, CeilDiv[int](2, 1))
	assert.Equal(t, 3, CeilDiv[int](3, 1))
	assert.Equal(t, 3, CeilDiv[int](5, 2))
}
