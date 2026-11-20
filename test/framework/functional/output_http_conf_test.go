package functional

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"sigs.k8s.io/yaml"
)

// TestVectorConfFactoryYAML guards the Vector receiver config template: it must
// render valid YAML for every option combination the framework uses, since the
// receiver container is started with --config-yaml.
func TestVectorConfFactoryYAML(t *testing.T) {
	cases := []struct {
		name    string
		profile configv1.TLSProfileType
		options []Option
	}{
		{"default", "", nil},
		{"with tls profile", configv1.TLSProfileModernType, nil},
		{"with credentials", "", []Option{{"username", "u"}, {"password", "p"}}},
		{"benchmark template", "", []Option{{"path", "/tmp/{{.container_name}}.log"}, {"template", VectorHttpSourceBenchmarkConfTemplate}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			conf := VectorConfFactory(c.profile, c.options...)
			var generic map[string]interface{}
			if err := yaml.Unmarshal([]byte(conf), &generic); err != nil {
				t.Fatalf("rendered config is not valid YAML:\n%s\nerror: %v", conf, err)
			}
			sources, ok := generic["sources"].(map[string]interface{})
			if !ok {
				t.Fatalf("expected sources map in:\n%s", conf)
			}
			if _, ok := sources["my_source"]; !ok {
				t.Fatalf("expected my_source in:\n%s", conf)
			}
		})
	}
}
