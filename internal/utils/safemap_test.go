package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringMap(t *testing.T) {
	var y float64 = 0.04
	m := StringMap{M: map[string]string{"x": "y"}}
	assert.Equal(t, "y", m.Get("x"))
	m.Set("x", "z")
	assert.Equal(t, "z", m.Get("x"))
	f := Float64Map{M: map[string]float64{"x": y}}
	f.Inc("x")
	newy := f.Get("x")
	assert.Equal(t, y+1, newy)
}
