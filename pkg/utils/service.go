package utils

import (
	"fmt"

	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
func RemoveService(namespace string, serviceName string) error {

	service := NewService(
		serviceName,
		namespace,
		serviceName,
		[]core.ServicePort{},
	)

	err := sdk.Delete(service)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v service %v", serviceName, err)
	}

	return nil
}
