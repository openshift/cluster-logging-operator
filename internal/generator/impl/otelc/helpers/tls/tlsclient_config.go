package tls

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

// NewTlsClientConfig converts observability OutputTLSSpec to OTLP TlsClientConfig.
func NewTlsClientConfig(spec *obs.OutputTLSSpec, op utils.Options) *types.TlsClientConfig {
	if spec == nil {
		return nil
	}

	config := &types.TlsClientConfig{
		InsecureSkipVerify: spec.InsecureSkipVerify,
		CAFile:             helpers.ValuePath(spec.CA, "%s"),
		CertFile:           helpers.ValuePath(spec.Certificate, "%s"),
		KeyFile:            helpers.SecretPath(spec.Key, "%s"),
	}

	return config
}

// SetTLSProfile updates the tls and cipher specs from the options given
// TODO: Remove internal/generator/vector/output/common/tls
func SetTLSProfile(t *types.TlsClientConfig, op utils.Options) *types.TlsClientConfig {
	if version, found := op[framework.MinTLSVersion]; found {
		t.MinVersion = version.(string)
	}
	//if ciphers, found := op[framework.Ciphers]; found {
	//	t.CipherSuites = ciphers.(string)
	//}
	return t
}
