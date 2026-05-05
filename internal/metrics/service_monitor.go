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
	prometheusCAFile          = "/etc/prometheus/configmaps/serving-certs-ca-bundle/service-ca.crt"
	prometheusBearerTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

func newServiceMonitor(namespace, name, serviceName string, owner metav1.OwnerReference, selector map[string]string, portName string, metricRelabelConfigs []*monitoringv1.RelabelConfig, profile string) *monitoringv1.ServiceMonitor {
	var endpoint = []monitoringv1.Endpoint{
		{
			Port:            portName,
			Path:            "/metrics",
			Scheme:          "https",
			BearerTokenFile: prometheusBearerTokenFile,
			TLSConfig: &monitoringv1.TLSConfig{
				CAFile: prometheusCAFile,
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					ServerName: fmt.Sprintf("%s.%s.svc", serviceName, namespace),
				},
			},
			MetricRelabelConfigs: metricRelabelConfigs,
		},
	}

	desired := runtime.NewServiceMonitor(namespace, name)
	if desired.Labels == nil {
		desired.Labels = map[string]string{}
	}
	desired.Labels[constants.LabelMetricsCollectionProfile] = profile
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

func ReconcileServiceMonitor(k8sClient client.Client, namespace, name, serviceName string, owner metav1.OwnerReference, selector map[string]string, portName string, metricRelabelConfigs []*monitoringv1.RelabelConfig, profile string) error {
	desired := newServiceMonitor(namespace, name, serviceName, owner, selector, portName, metricRelabelConfigs, profile)
	return reconcile.ServiceMonitor(k8sClient, desired)
}
