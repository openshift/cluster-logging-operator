package network

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func withServiceTypeLabel(serviceType string) func(o runtime.Object) {
	return func(o runtime.Object) {
		labels := map[string]string{
			constants.LabelLoggingServiceType: serviceType,
		}
		utils.AddLabels(runtime.Meta(o), labels)
	}
}

// ReconcileService reconciles the service that exposes metrics
func ReconcileService(k8sClient client.Client, namespace, name, instanceName, component, portName, certSecretName string, portNum int32, owner metav1.OwnerReference, visitors func(o runtime.Object)) error {
	desired := factory.NewService(
		name,
		namespace,
		component,
		instanceName,
		[]v1.ServicePort{
			{
				Port:       portNum,
				TargetPort: intstr.FromString(portName),
				Name:       portName,
			},
		},
		withServiceTypeLabel(constants.ServiceTypeMetrics),
		visitors,
	)

	desired.Annotations = map[string]string{
		constants.AnnotationServingCertSecretName: certSecretName,
	}
	utils.AddOwnerRefToObject(desired, owner)
	return reconcile.Service(k8sClient, desired)
}

func ReconcileInputService(k8sClient client.Client, namespace, name, instance, certSecretName string, port, targetPort int32, receiverType obs.ReceiverType, owner metav1.OwnerReference, visitors func(o runtime.Object)) error {
	desired := factory.NewService(
		name,
		namespace,
		constants.CollectorName,
		instance,
		[]v1.ServicePort{
			{
				Port: port,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: targetPort,
				},
				Protocol: v1.ProtocolTCP,
			},
		},
		withServiceTypeLabel(constants.ServiceTypeInput),
		visitors,
	)
	desired.Labels[constants.LabelLoggingInputServiceType] = string(receiverType)
	selectors := runtime.Selectors(instance, constants.CollectorName, desired.Labels[constants.LabelK8sName])
	desired.Spec.Selector = selectors

	desired.Annotations = map[string]string{
		constants.AnnotationServingCertSecretName: certSecretName,
	}

	utils.AddOwnerRefToObject(desired, owner)
	return reconcile.Service(k8sClient, desired)
}
