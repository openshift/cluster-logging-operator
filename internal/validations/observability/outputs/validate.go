package outputs

import (
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Validate(context internalcontext.ForwarderContext) (_ common.AttributeConditionType, results []metav1.Condition) {

	for _, out := range context.Forwarder.Spec.Outputs {
		results = append(results, ValidateSecretsAndConfigMaps(out, context.Secrets, context.ConfigMaps)...)
	}

	return common.AttributeConditionOutputs, results
}

func ValidateSecretsAndConfigMaps(spec obsv1.OutputSpec, secrets map[string]*corev1.Secret, configMaps map[string]*corev1.ConfigMap) []metav1.Condition {
	configs := internalobs.ConfigMapOrSecretKeys(spec.TLS.TLSSpec)
	configs = append(configs, internalobs.SecretKeysAsConfigMapOrSecretKeys(spec)...)
	return common.ValidateConfigMapOrSecretKey(spec.Name, configs, secrets, configMaps)
}
