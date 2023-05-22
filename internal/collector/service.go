package collector

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileService reconciles the service specifically for the collector that exposes the collector metrics
func (f *Factory) ReconcileService(er record.EventRecorder, k8sClient client.Client, namespace, name string, owner metav1.OwnerReference) error {
	desired := factory.NewService(
		constants.CollectorName,
		namespace,
		constants.CollectorName,
		[]v1.ServicePort{
			{
				Port:       MetricsPort,
				TargetPort: intstr.FromString(MetricsPortName),
				Name:       MetricsPortName,
			},
			{
				Port:       ExporterPort,
				TargetPort: intstr.FromString(ExporterPortName),
				Name:       ExporterPortName,
			},
		},
		f.CommonLabelInitializer,
	)

	desired.Annotations = map[string]string{
		constants.AnnotationServingCertSecretName: constants.CollectorMetricSecretName,
	}

	utils.AddOwnerRefToObject(desired, owner)
	return reconcile.Service(er, k8sClient, desired)
}
