package clusterlogforwarder

import (
	v1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"github.com/openshift/cluster-logging-operator/internal/validations/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	vector = sets.NewString(
		v1.OutputTypeElasticsearch,
		v1.OutputTypeGoogleCloudLogging,
		v1.OutputTypeHttp,
		v1.OutputTypeKafka,
		v1.OutputTypeLoki,
		v1.OutputTypeSplunk,
		v1.OutputTypeSyslog,
		v1.OutputTypeCloudwatch,
		v1.OutputTypeAzureMonitor,
	)

	fluentd = sets.NewString(
		v1.OutputTypeElasticsearch,
		v1.OutputTypeFluentdForward,
		v1.OutputTypeHttp,
		v1.OutputTypeKafka,
		v1.OutputTypeLoki,
		v1.OutputTypeSyslog,
		v1.OutputTypeCloudwatch,
	)
)

// validateCollectorCompatibility checking is given collector support proposed output if not error will return, nil otherwise
//
// Output type         | Supported collector type
// Elasticsearch       | Fluentd, Vector
// Fluent Forward      | Fluentd
// Google Cloud Logging| Vector
// HTTP                | Fluentd, Vector
// Kafka               | Fluentd, Vector
// Loki                | Fluentd, Vector
// Splunk              | Vector
// Syslog              | Fluentd, Vector
// Amazon CloudWatch   | Fluentd, Vector
// Azure Monitor Logs  | Vector
// TODO: Might we need this when we write migration code?
func validateCollectorCompatibility(clf v1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *v1.ClusterLogForwarderStatus) {
	collector := constants.VectorName
	supportedOutputs := vector
	if !extras[constants.VectorName] {
		collector = constants.FluentdName
		supportedOutputs = fluentd
	}

	for _, output := range clf.Spec.Outputs {
		if !supportedOutputs.Has(output.Type) {
			return errors.NewValidationError("ClusterLogForwarder '%s/%s' configured with %q collector does not support output type %q",
				clf.Namespace, clf.Name, collector, output.Type), nil
		}
	}

	return nil, nil
}
