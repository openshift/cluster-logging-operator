package utils

import (
	"os"
	"testing"
)

const (
	envKey   = "TEST"
	envValue = "value"
)

func TestLookupEnvWithDefaultDefined(t *testing.T) {
	os.Setenv(envKey, envValue)
	res := LookupEnvWithDefault(envKey, "should be ignored")
	if res != envValue {
		t.Errorf("Expected %s=%s but got %s=%s", envKey, envValue, envKey, res)
	}
}

func TestLookupEnvWithDefaultUndefined(t *testing.T) {
	expected := "defaulted"
	os.Unsetenv(envKey)
	res := LookupEnvWithDefault(envKey, expected)
	if res != expected {
		t.Errorf("Expected %s=%s but got %s=%s", envKey, expected, envKey, res)
	}
}
