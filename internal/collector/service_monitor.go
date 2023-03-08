package collector

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

func NewServiceMonitor(namespace, name string, owner metav1.OwnerReference) *monitoringv1.ServiceMonitor {
	var endpoints []monitoringv1.Endpoint
	for _, portName := range []string{MetricsPortName, ExporterPortName} {
		endpoints = append(endpoints,
			monitoringv1.Endpoint{
				Port:   portName,
				Path:   "/metrics",
				Scheme: "https",
				TLSConfig: &monitoringv1.TLSConfig{
					CAFile: prometheusCAFile,
					SafeTLSConfig: monitoringv1.SafeTLSConfig{
						ServerName: fmt.Sprintf("%s.%s.svc", name, namespace),
					},
				},
			},
		)
	}
	desired := runtime.NewServiceMonitor(namespace, name)
	desired.Spec = monitoringv1.ServiceMonitorSpec{
		JobLabel:  "monitor-collector",
		Endpoints: endpoints,
		Selector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"logging-infra": "support",
			},
		},
		NamespaceSelector: monitoringv1.NamespaceSelector{
			MatchNames: []string{namespace},
		},
	}

	utils.AddOwnerRefToObject(desired, owner)

	return desired
}

// ReconcileServiceMonitor reconciles the service monitor specifically for exposing collector metrics
func ReconcileServiceMonitor(er record.EventRecorder, k8sClient client.Client, namespace, name string, owner metav1.OwnerReference) error {
	desired := NewServiceMonitor(namespace, name, owner)
	return reconcile.ServiceMonitor(er, k8sClient, desired)
}

func RemoveServiceMonitor(er record.EventRecorder, k8sClient client.Client, namespace, name string) {
	serviceMonitor := runtime.NewServiceMonitor(namespace, name)
	if err := k8sClient.Delete(context.TODO(), serviceMonitor); err != nil && !errors.IsNotFound(err) {
		log.Error(err, "collector.RemoveServiceMonitor")
		er.Eventf(serviceMonitor, v1.EventTypeWarning, constants.EventReasonRemoveObject, "Unable to remove ServiceMonitor %s/%s: %v", namespace, name, err)
	} else {
		er.Eventf(serviceMonitor, v1.EventTypeNormal, constants.EventReasonRemoveObject, "Removed ServiceMonitor %s/%s", namespace, name)
	}
}
