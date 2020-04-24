package kibana

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

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
				"logging-infra": "support",
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

//RemoveService with given name and namespace
func (clusterRequest *KibanaRequest) RemoveService(serviceName string) error {

	service := NewService(
		serviceName,
		clusterRequest.cluster.Namespace,
		serviceName,
		[]core.ServicePort{},
	)

	err := clusterRequest.Delete(service)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v service %v", serviceName, err)
	}

	return nil
}
