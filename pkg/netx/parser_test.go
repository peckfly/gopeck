package netx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_parseQueryMap(t *testing.T) {
	tests := []struct {
		data   map[string]string
		expect string
	}{
		{
			data:   map[string]string{"a": "1", "b": "2", "c": "3"},
			expect: "a=1&b=2&c=3",
		},
		{
			data:   map[string]string{"a": "1", "b": "2"},
			expect: "a=1&b=2",
		},
	}
	for _, tt := range tests {
		result := ParseQueryMap(tt.data)
		assert.Equal(t, tt.expect, result)
	}
}

func Test_parseQuery(t *testing.T) {
	tests := []struct {
		data   string
		expect map[string]string
	}{
		{
			data:   "a=1&b=2&c=3",
			expect: map[string]string{"a": "1", "b": "2", "c": "3"},
		},
		{
			data:   "a=1&b=2",
			expect: map[string]string{"a": "1", "b": "2"},
		},
		{
			data:   "a=1",
			expect: map[string]string{"a": "1"},
		},
	}
	for _, tt := range tests {
		result := ParseQuery(tt.data)
		assert.Equal(t, tt.expect, result)
	}
}
