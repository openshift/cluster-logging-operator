package factory

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//NewService stubs an instance of a Service
func NewService(serviceName string, namespace string, selectorComponent string, servicePorts []core.ServicePort) *core.Service {
	return &core.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: core.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels: map[string]string{
				"logging-infra":                "support",
				"app.kubernetes.io/name":       serviceName,
				"app.kubernetes.io/component":  constants.CollectorName,
				"app.kubernetes.io/created-by": constants.ClusterLoggingOperator,
				"app.kubernetes.io/managed-by": constants.ClusterLoggingOperator,
			},
		},
		Spec: core.ServiceSpec{
			Selector: map[string]string{
				"component": selectorComponent,
				"provider":  "openshift",
			},
			Ports: servicePorts,
		},
	}
}
