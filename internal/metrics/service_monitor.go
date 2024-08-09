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

func NewServiceMonitor(namespace, name, component, portName string, owner metav1.OwnerReference) *monitoringv1.ServiceMonitor {
	var endpoint = []monitoringv1.Endpoint{
		{
			Port:   portName,
			Path:   "/metrics",
			Scheme: "https",
			TLSConfig: &monitoringv1.TLSConfig{
				CAFile: prometheusCAFile,
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					ServerName: fmt.Sprintf("%s.%s.svc", name, namespace),
				},
			},
			// Replaces labels that have `-` with `_`
			// Example:
			// app_kubernetes_io_part-of -> app_kubernetes_io_part_of
			MetricRelabelConfigs: []*monitoringv1.RelabelConfig{
				{
					SourceLabels: []monitoringv1.LabelName{
						"__name__",
					},
					TargetLabel: "__name__",
					Regex:       "(.*)-(.*)",
					Replacement: "${1}_${2}",
				},
			},
		},
	}

	desired := runtime.NewServiceMonitor(namespace, name)
	desired.Spec = monitoringv1.ServiceMonitorSpec{
		JobLabel:  fmt.Sprintf("monitor-%s", name),
		Endpoints: endpoint,
		Selector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"logging-infra":             "support",
				constants.LabelK8sInstance:  name,
				constants.LabelK8sComponent: component,
			},
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

func ReconcileServiceMonitor(k8sClient client.Client, namespace, name, component, portName string, owner metav1.OwnerReference) error {
	desired := NewServiceMonitor(namespace, name, component, portName, owner)
	return reconcile.ServiceMonitor(k8sClient, desired)
}
