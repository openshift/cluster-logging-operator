package metrics

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	prometheusCAFile = "/etc/prometheus/configmaps/serving-certs-ca-bundle/service-ca.crt"
)

func newServiceMonitor(namespace, name string, owner metav1.OwnerReference, selector map[string]string, portName string) *monitoringv1.ServiceMonitor {
	replacement := "${1}_${2}"
	serverName := fmt.Sprintf("%s.%s.svc", name, namespace)
	var endpoint = []monitoringv1.Endpoint{
		{
			Port:   portName,
			Path:   "/metrics",
			Scheme: "https",
			TLSConfig: &monitoringv1.TLSConfig{
				CAFile: prometheusCAFile,
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					ServerName: &serverName,
				},
			},
			// Replaces labels that have `-` with `_`
			// Example:
			// app_kubernetes_io_part-of -> app_kubernetes_io_part_of
			MetricRelabelConfigs: []monitoringv1.RelabelConfig{
				{
					SourceLabels: []monitoringv1.LabelName{
						"__name__",
					},
					TargetLabel: "__name__",
					Regex:       "(.*)-(.*)",
					Replacement: &replacement,
				},
			},
		},
	}

	desired := runtime.NewServiceMonitor(namespace, name)
	desired.Spec = monitoringv1.ServiceMonitorSpec{
		JobLabel:  fmt.Sprintf("monitor-%s", name),
		Endpoints: endpoint,
		Selector: metav1.LabelSelector{
			MatchLabels: selector,
		},
		NamespaceSelector: monitoringv1.NamespaceSelector{
			MatchNames: []string{namespace},
		},
		PodTargetLabels: []string{
			constants.LabelK8sName,
			constants.LabelK8sComponent,
			constants.LabelK8sPartOf,
			constants.LabelK8sInstance,
		},
	}

	utils.AddOwnerRefToObject(desired, owner)

	return desired
}

func BuildSelector(component, instance string) map[string]string {
	return map[string]string{
		constants.LabelLoggingServiceType: constants.ServiceTypeMetrics,
		constants.LabelK8sComponent:       component,
		constants.LabelK8sInstance:        instance,
	}
}

func ReconcileServiceMonitor(k8sClient client.Client, namespace, name string, owner metav1.OwnerReference, selector map[string]string, portName string) error {
	desired := newServiceMonitor(namespace, name, owner, selector, portName)
	return reconcile.ServiceMonitor(k8sClient, desired)
}
