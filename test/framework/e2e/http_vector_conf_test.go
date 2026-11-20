package e2e

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"sigs.k8s.io/yaml"
)

// TestVectorHttpConfYAML guards the non-TLS e2e http receiver config: it must
// be valid YAML since it is stored in the ConfigMap under vector.ConfigFile
// ("vector.yaml") and the receiver runs vector --config-yaml against it.
func TestVectorHttpConfYAML(t *testing.T) {
	var generic map[string]interface{}
	if err := yaml.Unmarshal([]byte(vectorHttpConf), &generic); err != nil {
		t.Fatalf("vectorHttpConf is not valid YAML:\n%s\nerror: %v", vectorHttpConf, err)
	}
	sources, ok := generic["sources"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected sources map in:\n%s", vectorHttpConf)
	}
	if _, ok := sources["my_source"]; !ok {
		t.Fatalf("expected my_source in:\n%s", vectorHttpConf)
	}
	for _, section := range []string{"transforms", "sinks"} {
		if _, ok := generic[section]; !ok {
			t.Fatalf("expected %s section in:\n%s", section, vectorHttpConf)
		}
	}
}

// TestVectorHttpConfWithTLSProfileYAML guards the e2e http receiver config: it
// must render valid YAML (with and without curves) since the receiver container
// is started against vector.ConfigFile.
func TestVectorHttpConfWithTLSProfileYAML(t *testing.T) {
	for name, profileSpec := range map[string]configv1.TLSProfileSpec{
		"without groups": {MinTLSVersion: configv1.VersionTLS12, Ciphers: []string{"TLS_AES_128_GCM_SHA256"}},
		"with groups":    {MinTLSVersion: configv1.VersionTLS12, Ciphers: []string{"TLS_AES_128_GCM_SHA256"}, Groups: []configv1.TLSGroup{"X25519"}},
	} {
		t.Run(name, func(t *testing.T) {
			conf := vectorHttpConfWithTLSProfile(profileSpec)
			var generic map[string]interface{}
			if err := yaml.Unmarshal([]byte(conf), &generic); err != nil {
				t.Fatalf("rendered config is not valid YAML:\n%s\nerror: %v", conf, err)
			}
			sources, ok := generic["sources"].(map[string]interface{})
			if !ok {
				t.Fatalf("expected sources map in:\n%s", conf)
			}
			mySource, ok := sources["my_source"].(map[string]interface{})
			if !ok {
				t.Fatalf("expected my_source in:\n%s", conf)
			}
			tls, ok := mySource["tls"].(map[string]interface{})
			if !ok {
				t.Fatalf("expected my_source.tls in:\n%s", conf)
			}
			if _, ok := tls["min_tls_version"]; !ok {
				t.Fatalf("expected my_source.tls.min_tls_version in:\n%s", conf)
			}
		})
	}
}
