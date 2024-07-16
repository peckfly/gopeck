package netx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPing(t *testing.T) {
	tests := []struct {
		url    string
		expect bool
	}{
		{
			url:    "https://www.baidu.com",
			expect: true,
		},
		{
			url:    "https://www.google.com",
			expect: true,
		},
		{
			url:    "https://cccgg.com",
			expect: false,
		},
	}
	for _, tt := range tests {
		err := Ping(tt.url)
		assert.Equal(t, tt.expect, err == nil)
	}
}
