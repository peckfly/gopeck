package bufferx

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestBuffer_Push(t *testing.T) {
	b := NewBuffer[string](2)
	s := []string{"a", "b", "c", "d", "e"}
	var result string
	f := func(bs []string) {
		result += strings.Join(bs, "-")
	}
	for _, e := range s {
		b.Push(e, f)
	}
	b.Flush(f)
	assert.Equal(t, "a-bc-de", result)
}
