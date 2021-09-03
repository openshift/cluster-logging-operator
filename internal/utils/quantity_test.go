package utils

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestParseQuantity(t *testing.T) {
	if _, err := ParseQuantity(""); err == nil {
		t.Errorf("expected empty string to return error")
	}

	table := []struct {
		input  string
		expect resource.Quantity
	}{
		{"750k", resource.MustParse("750Ki")},
		{"750K", resource.MustParse("750Ki")},
		{"750Ki", resource.MustParse("750Ki")},
		{"8m", resource.MustParse("8Mi")},
		{"8M", resource.MustParse("8Mi")},
		{"8Mi", resource.MustParse("8Mi")},
		{"1g", resource.MustParse("1Gi")},
		{"1G", resource.MustParse("1Gi")},
		{"1Gi", resource.MustParse("1Gi")},
	}

	for _, item := range table {
		got, err := ParseQuantity(item.input)
		if err != nil {
			t.Errorf("%v: unexpected error: %v", item.input, err)
			continue
		}
		if actual, expect := got.Value(), item.expect.Value(); actual != expect {
			t.Errorf("%v: expected %v, got %v", item.input, expect, actual)
		}
	}
}
