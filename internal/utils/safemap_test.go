package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeMap(t *testing.T) {
	var y float64 = 0.04
	m := InitStringMap(map[string]string{"x": "y"})
	assert.Equal(t, "y", m.Get("x"))
	m.Set("x", "z")
	assert.Equal(t, "z", m.Get("x"))
	f := InitFloat64Map(map[string]float64{"x": y})
	f.Inc("x")
	newy := f.Get("x")
	assert.Equal(t, y+1, newy)
}
