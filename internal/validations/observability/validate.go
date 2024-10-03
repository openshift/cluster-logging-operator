package observability

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/filters"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/inputs"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/outputs"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/pipelines"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	clfValidators = []func(internalcontext.ForwarderContext){
		validateAnnotations,
		ValidatePermissions,
		inputs.Validate,
		outputs.Validate,
		filters.Validate,
		pipelines.Validate,
	}

	clfPreValidators = []func(internalcontext.ForwarderContext){
		outputs.ValidateLokistackOTLPForAnnotation,
	}
)

// ValidateClusterLogForwarder validates the forwarder spec that can not be accomplished using api attributes and returns a set of conditions that apply to the spec
func ValidateClusterLogForwarder(context internalcontext.ForwarderContext) {
	for _, validate := range clfValidators {
		validate(context)
	}
}

// PreValidateClusterLogForwarder validates the forwarder spec before initialization
// Currently supporting one use-case, forwarding OTEL data format to Lokistack
func PreValidateClusterLogForwarder(context internalcontext.ForwarderContext) {
	for _, preValidate := range clfPreValidators {
		preValidate(context)
	}
}

func MustUndeployCollector(conditions []metav1.Condition) bool {
	for _, condition := range conditions {
		if condition.Type == obs.ConditionTypeAuthorized && condition.Status == obs.ConditionFalse {
			return true
		}
	}
	return false
}
