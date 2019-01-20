package k8shandler

import (
	"fmt"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	v1alpha1 "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	prometheusCAFile = "/etc/prometheus/configmaps/serving-certs-ca-bundle/service-ca.crt"
)

// CreateOrUpdateServiceMonitors ensures the existence of ServiceMonitors for Elasticsearch cluster
func CreateOrUpdateServiceMonitors(dpl *v1alpha1.Elasticsearch) error {
	serviceMonitorName := fmt.Sprintf("monitor-%s-%s", dpl.Name, "cluster")
	owner := asOwner(dpl)

	labelsWithDefault := appendDefaultLabel(dpl.Name, dpl.Labels)

	elasticsearchScMonitor := createServiceMonitor(serviceMonitorName, dpl.Namespace, labelsWithDefault)
	addOwnerRefToObject(elasticsearchScMonitor, owner)
	err := sdk.Create(elasticsearchScMonitor)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Elasticsearch ServiceMonitor: %v", err)
	}

	// TODO: handle update - use retry.RetryOnConflict

	return nil
}

func createServiceMonitor(serviceMonitorName, namespace string, labels map[string]string) *monitoringv1.ServiceMonitor {
	svcMonitor := serviceMonitor(serviceMonitorName, namespace, labels)
	labelSelector := metav1.LabelSelector{
		MatchLabels: labels,
	}
	tlsConfig := monitoringv1.TLSConfig{
		CAFile: prometheusCAFile,
	}
	endpoint := monitoringv1.Endpoint{
		Port:            "restapi",
		Path:            "/_prometheus/metrics",
		Scheme:          "https",
		BearerTokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
		TLSConfig:       &tlsConfig,
	}
	svcMonitor.Spec = monitoringv1.ServiceMonitorSpec{
		JobLabel:  "monitor-elasticsearch",
		Endpoints: []monitoringv1.Endpoint{endpoint},
		Selector:  labelSelector,
	}
	return svcMonitor
}

func serviceMonitor(serviceMonitorName string, namespace string, labels map[string]string) *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.ServiceMonitorsKind,
			APIVersion: monitoringv1.Group + "/" + monitoringv1.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceMonitorName,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}
