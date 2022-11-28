package tls

import (
	"bytes"
	"strings"
	"text/template"

	configv1 "github.com/openshift/api/config/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	OpenSSLConfTemplate = `
{{define "openssl_conf_template" -}}

config_diagnostics = 1

openssl_conf = default_conf_section

[default_conf_section]
ssl_conf = ssl_section

[ssl_section]
system_default = system_default_section

[system_default_section]
MinProtocol = {{MinProtocol}}
CipherSuites = {{CipherSuites}}
{{end}}
`
)

func OpenSSLConf(k8client client.Client) string {
	pr, _ := FetchAPIServerTlsProfile(k8client)
	tls := GetTLSProfileSpec(pr)
	conf, _ := opensslConf(tls)
	return conf
}

func opensslConf(tlsConf configv1.TLSProfileSpec) (string, error) {
	t := template.New("openssl_conf_template")
	t.Funcs(template.FuncMap{
		"MinProtocol": func() string {
			switch MinTLSVersion(tlsConf) {
			case string(configv1.VersionTLS10):
				return "TLSv1.0"
			case string(configv1.VersionTLS11):
				return "TLSv1.1"
			case string(configv1.VersionTLS12):
				return "TLSv1.2"
			case string(configv1.VersionTLS13):
				return "TLSv1.3"
			}
			return "invalid"
		},
		"CipherSuites": func() string {
			return strings.TrimSpace(strings.Join(TLSCiphers(tlsConf), ", "))
		},
	})
	t, err := t.Parse(OpenSSLConfTemplate)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = t.ExecuteTemplate(buf, "openssl_conf_template", nil)
	return buf.String(), err
}
