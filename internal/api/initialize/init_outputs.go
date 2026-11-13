package initialize

import (
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (

	// DefaultMaxWriteLokiStack is the maximum Loki push size applied when a
	// lokiStack output does not set tuning.maxWrite.
	//
	// 1Mi stays inside the per-distributor burst of a 1x.extra-small LokiStack.
	// That size configures a 6MiB global ingestion burst, and Loki does not
	// divide the burst across distributors. Collectors push concurrently, so a
	// 3Mi batch consumes the burst and the next full batch is rejected with
	// HTTP 429 while the aggregate ingest rate is still under the global limit.
	// Larger sizes can raise tuning.maxWrite.
	DefaultMaxWriteLokiStack = "1Mi"
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
					panic(fmt.Sprintf("Failed to parmaxWritese Max Write value for lokistack output %s/%s: %v", forwarder.Namespace, forwarder.Name, err))
				}
				o.LokiStack.Tuning.MaxWrite = &maxWrite
			}
		}
	}
	return forwarder
}
