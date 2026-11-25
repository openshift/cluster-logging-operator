package tls

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	collectorcommon "github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

// valuePath resolves a ValueReference for the otelc collector: configmap
// references use the OpenTelemetry operator's mount convention, secret
// references use the well-known collector secrets directory mounted by the
// ClusterLogForwarder controller
func valuePath(resource *obs.ValueReference) string {
	if resource == nil {
		return ""
	}
	if resource.SecretName != "" {
		return helpers.SecretPath(&obs.SecretReference{
			SecretName: resource.SecretName,
			Key:        resource.Key,
		}, "%s")
	}
	if resource.ConfigMapName != "" {
		return collectorcommon.OtelConfigMapPath(resource.ConfigMapName, resource.Key)
	}
	return ""
}

// NewTlsClientConfig converts observability OutputTLSSpec to OTLP TlsClientConfig.
func NewTlsClientConfig(spec *obs.OutputTLSSpec, op utils.Options) *types.TlsClientConfig {
	if spec == nil {
		return nil
	}

	config := &types.TlsClientConfig{
		InsecureSkipVerify: spec.InsecureSkipVerify,
		CAFile:             valuePath(spec.CA),
		CertFile:           valuePath(spec.Certificate),
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
