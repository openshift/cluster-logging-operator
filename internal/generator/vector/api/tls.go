package api

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

type TLS struct {
	Enabled           bool   `json:"enabled" yaml:"enabled,omitempty" toml:"enabled,omitempty"`
	MinTlsVersion     string `json:"min_tls_version,omitempty" yaml:"min_tls_version,omitempty" toml:"min_tls_version,omitempty"`
	CipherSuites      string `json:"ciphersuites,omitempty" yaml:"ciphersuites,omitempty" toml:"ciphersuites,omitempty"`
	VerifyCertificate bool   `json:"verify_certificate,omitempty" yaml:"verify_certificate,omitempty" toml:"verify_certificate,omitempty"`
	VerifyHostname    bool   `json:"verify_hostname,omitempty" yaml:"verify_hostname,omitempty" toml:"verify_hostname,omitempty"`
	KeyFile           string `json:"key_file,omitempty" yaml:"key_file,omitempty" toml:"key_file,omitempty"`
	CRTFile           string `json:"crt_file,omitempty" yaml:"crt_file,omitempty" toml:"crt_file,omitempty"`
	CAFile            string `json:"ca_file,omitempty" yaml:"ca_file,omitempty" toml:"ca_file,omitempty"`
	KeyPass           string `json:"key_pass,omitempty" yaml:"key_pass,omitempty" toml:"key_pass,omitempty"`
}

// SetTLSProfile updates the tls and cipher specs from the options given
// TODO: Remove internal/generator/vector/output/common/tls
func (t *TLS) SetTLSProfile(op utils.Options) *TLS {
	if version, found := op[framework.MinTLSVersion]; found {
		t.MinTlsVersion = version.(string)
	}
	if ciphers, found := op[framework.Ciphers]; found {
		t.CipherSuites = ciphers.(string)
	}
	return t
}
