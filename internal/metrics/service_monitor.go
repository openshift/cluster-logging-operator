package metrics

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
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

func ReconcileServiceMonitor(er record.EventRecorder, k8sClient client.Client, namespace, name, component, portName string, owner metav1.OwnerReference) error {
	desired := NewServiceMonitor(namespace, name, component, portName, owner)
	return reconcile.ServiceMonitor(er, k8sClient, desired)
}

func RemoveServiceMonitor(er record.EventRecorder, k8sClient client.Client, namespace, name string) {
	serviceMonitor := runtime.NewServiceMonitor(namespace, name)
	if err := k8sClient.Delete(context.TODO(), serviceMonitor); err != nil && !errors.IsNotFound(err) {
		log.Error(err, "logfilemetricexporter.RemoveServiceMonitor")
		er.Eventf(serviceMonitor, v1.EventTypeWarning, constants.EventReasonRemoveObject, "Unable to remove ServiceMonitor %s/%s: %v", namespace, name, err)
	} else {
		er.Eventf(serviceMonitor, v1.EventTypeNormal, constants.EventReasonRemoveObject, "Removed ServiceMonitor %s/%s", namespace, name)
	}
}
