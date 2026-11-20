package apiaudit

import (
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
)

// TestVectorConfYAML guards the local Vector config renderer: a multi-line VRL
// program must embed as a YAML block scalar that parses back to the identical
// source, since the test pod runs `vector -c /tmp/vector.yaml`.
func TestVectorConfYAML(t *testing.T) {
	vrl := "if true {\n  .msg = '''not a toml literal'''\n}"
	conf := vectorConf(vrl)

	var generic map[string]interface{}
	if err := yaml.Unmarshal([]byte(conf), &generic); err != nil {
		t.Fatalf("rendered config is not valid YAML:\n%s\nerror: %v", conf, err)
	}
	transforms, ok := generic["transforms"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected transforms map in:\n%s", conf)
	}
	policy, ok := transforms["policy"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected transforms.policy in:\n%s", conf)
	}
	// A "|" block scalar clips to exactly one trailing newline; VRL ignores it.
	expected := strings.TrimRight(vrl, "\n") + "\n"
	if policy["source"] != expected {
		t.Fatalf("expected VRL source %q, got %q", expected, policy["source"])
	}
}
