package kibana

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewServiceMonitor(serviceMonitorName, namespace string) *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.ServiceMonitorsKind,
			APIVersion: monitoringv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceMonitorName,
			Namespace: namespace,
		},
	}
}

func (clusterRequest *KibanaRequest) RemoveServiceMonitor(smName string) error {

	serviceMonitor := NewServiceMonitor(smName, clusterRequest.cluster.Namespace)

	err := clusterRequest.Delete(serviceMonitor)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v service monitor: %v", serviceMonitor, err)
	}

	return nil
}
