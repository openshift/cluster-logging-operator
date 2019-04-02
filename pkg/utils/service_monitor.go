package utils

import (
	"fmt"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/apimachinery/pkg/api/errors"
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

func RemoveServiceMonitor(namespace string, smName string) error {

	serviceMonitor := NewServiceMonitor(smName, namespace)

	err := sdk.Delete(serviceMonitor)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v service monitor: %v", serviceMonitor, err)
	}

	return nil
}
