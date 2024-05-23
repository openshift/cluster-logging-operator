package network

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileService reconciles the service that exposes metrics
func ReconcileService(er record.EventRecorder, k8sClient client.Client, namespace, name, component, portName, certSecretName string, portNum int32, owner metav1.OwnerReference, visitors func(o runtime.Object)) error {
	desired := factory.NewService(
		name,
		namespace,
		component,
		name,
		[]v1.ServicePort{
			{
				Port:       portNum,
				TargetPort: intstr.FromString(portName),
				Name:       portName,
			},
		},
		visitors,
	)

	desired.Annotations = map[string]string{
		constants.AnnotationServingCertSecretName: certSecretName,
	}
	utils.AddOwnerRefToObject(desired, owner)
	return reconcile.Service(er, k8sClient, desired)
}

func ReconcileInputService(er record.EventRecorder, k8sClient client.Client, namespace, name, instance, certSecretName string, port, targetPort int32, receiverType obs.ReceiverType, isDaemonset bool, owner metav1.OwnerReference, visitors func(o runtime.Object)) error {
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
		visitors,
	)

	if !isDaemonset {
		desired.Spec.Selector[constants.CollectorDeploymentKind] = constants.DeploymentType
	}

	desired.Annotations = map[string]string{
		constants.AnnotationServingCertSecretName: certSecretName,
	}

	desired.Labels[constants.LabelComponent] = fmt.Sprintf("%s-input-service", receiverType)

	utils.AddOwnerRefToObject(desired, owner)
	return reconcile.Service(er, k8sClient, desired)
}
