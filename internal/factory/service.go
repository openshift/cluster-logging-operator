package factory

import (
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	core "k8s.io/api/core/v1"
)

// NewService stubs an instance of a Service
func NewService(serviceName string, namespace string, selectorComponent string, servicePorts []core.ServicePort, visitors ...func(o runtime.Object)) *core.Service {
	service := runtime.NewService(namespace, serviceName, visitors...)
	selector := map[string]string{
		"component": selectorComponent,
		"provider":  "openshift",
	}
	runtime.NewServiceBuilder(service).WithSelector(selector).WithServicePort(servicePorts).AddLabel("logging-infra", "support")
	return service
}
