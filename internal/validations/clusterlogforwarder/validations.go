package clusterlogforwarder

import (
	v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Validate(clf v1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *v1.ClusterLogForwarderStatus) {
	returnStatus := v1.ClusterLogForwarderStatus{}
	for _, validate := range validations {
		if err, status := validate(clf, k8sClient, extras); err != nil {
			return err, status
		} else if status != nil {
			returnStatus.Conditions = append(returnStatus.Conditions, status.Conditions...)
		}
	}
	return nil, &returnStatus
}

// validations are the set of admission rules for validating
// a ClusterLogForwarder
var validations = []func(clf v1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *v1.ClusterLogForwarderStatus){
	ValidateClusterLoggingDependency,
	ValidateInputsOutputsPipelines,
	validateJsonParsingToElasticsearch,
	validateUrlAccordingToTls,
	validateHttpContentTypeHeaders,
}
