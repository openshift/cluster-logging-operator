package network

import (
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
func ReconcileService(er record.EventRecorder, k8sClient client.Client, namespace, name, selectorComponent, portName, certSecretName string, portNum int32, owner metav1.OwnerReference, visitors func(o runtime.Object)) error {
	desired := factory.NewService(
		name,
		namespace,
		selectorComponent,
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
