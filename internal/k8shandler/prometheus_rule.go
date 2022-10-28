package k8shandler

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (clusterRequest *ClusterLoggingRequest) RemovePrometheusRule(ruleName string) error {

	promRule := runtime.NewPrometheusRule(clusterRequest.Cluster.Namespace, ruleName)

	err := clusterRequest.Delete(promRule)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v prometheus rule: %v", promRule, err)
	}

	return nil
}
