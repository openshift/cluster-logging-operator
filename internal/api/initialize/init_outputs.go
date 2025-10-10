package initialize

import (
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (

	// DefaultMaxWriteLokiStack is the maximum request size for LokiStack clusters when otherwise not defined by the output
	DefaultMaxWriteLokiStack = "3Mi"
)

// InitOutputs initializes outputs (i.e. defaulting output field attributes as needed)
func InitOutputs(forwarder obs.ClusterLogForwarder, _ utils.Options) obs.ClusterLogForwarder {
	for _, o := range forwarder.Spec.Outputs {
		if o.Type == obs.OutputTypeLokiStack {
			if o.LokiStack.Tuning == nil {
				o.LokiStack.Tuning = &obs.LokiTuningSpec{}
			}
			if o.LokiStack.Tuning.MaxWrite == nil {
				maxWrite, err := resource.ParseQuantity(DefaultMaxWriteLokiStack)
				if err != nil {
					panic(fmt.Sprintf("Failed to initialize  Max Write value for lokistack output %s/%s: %v", forwarder.Namespace, forwarder.Name, err))
				}
				o.LokiStack.Tuning.MaxWrite = &maxWrite
			}
		}
	}
	return forwarder
}
