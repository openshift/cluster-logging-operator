package utils

// monitoring.go is a set of methods for monitoring

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewServiceMonitor(serviceMonitorName, namespace string) *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.ServiceMonitorsKind,
			APIVersion: monitoringv1.Group + "/" + monitoringv1.Version,
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

func NewPrometheusRule(ruleName, namespace string) *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PrometheusRuleKind,
			APIVersion: monitoringv1.Group + "/" + monitoringv1.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ruleName,
			Namespace: namespace,
		},
	}
}

func RemovePrometheusRule(namespace string, ruleName string) error {

	promRule := NewPrometheusRule(ruleName, namespace)

	err := sdk.Delete(promRule)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v prometheus rule: %v", promRule, err)
	}

	return nil
}
