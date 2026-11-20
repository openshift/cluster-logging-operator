package functional

import (
	"testing"

	"sigs.k8s.io/yaml"
)

// TestSetS3BatchTimeout guards the YAML set helper: it must set the nested
// timeout, create missing intermediate maps, and refuse to silently replace a
// non-map intermediate value with a fresh map (that would swallow a malformed
// config instead of surfacing it).
func TestSetS3BatchTimeout(t *testing.T) {
	t.Run("sets an existing nested value", func(t *testing.T) {
		in := "sinks:\n  output_s3:\n    batch:\n      timeout_secs: 5\n"
		out, err := SetS3BatchTimeout(in, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var m map[string]interface{}
		if err := yaml.Unmarshal([]byte(out), &m); err != nil {
			t.Fatalf("output not valid YAML: %v", err)
		}
		got := m["sinks"].(map[string]interface{})["output_s3"].(map[string]interface{})["batch"].(map[string]interface{})["timeout_secs"]
		if got != float64(10) {
			t.Fatalf("expected timeout_secs 10, got %v", got)
		}
	})

	t.Run("creates a missing path", func(t *testing.T) {
		out, err := SetS3BatchTimeout("sinks: {}\n", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var m map[string]interface{}
		if err := yaml.Unmarshal([]byte(out), &m); err != nil {
			t.Fatalf("output not valid YAML: %v", err)
		}
		got := m["sinks"].(map[string]interface{})["output_s3"].(map[string]interface{})["batch"].(map[string]interface{})["timeout_secs"]
		if got != float64(10) {
			t.Fatalf("expected timeout_secs 10, got %v", got)
		}
	})

	t.Run("errors when an intermediate key holds a non-map value", func(t *testing.T) {
		if _, err := SetS3BatchTimeout("sinks: not-a-map\n", 10); err == nil {
			t.Fatal("expected an error for non-map intermediate value, got nil")
		}
	})
}
