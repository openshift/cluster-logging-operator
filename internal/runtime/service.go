package runtime

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ServiceBuilder struct {
	Service *corev1.Service
}

// NewService returns a corev1.Service with namespace and name.
func NewService(namespace, name string) *corev1.Service {
	svc := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{},
		},
	}
	Initialize(svc, namespace, name)
	return svc
}

func NewServiceBuilder(namespace, name string) *ServiceBuilder {
	return &ServiceBuilder{
		Service: NewService(namespace, name),
	}
}

func NewServiceBuilderFor(svc *corev1.Service) *ServiceBuilder {
	return &ServiceBuilder{
		Service: svc,
	}
}

func (builder *ServiceBuilder) WithAnnotations(annotations map[string]string) *ServiceBuilder {
	builder.Service.Annotations = annotations
	return builder
}

func (builder *ServiceBuilder) WithLabels(labels map[string]string) *ServiceBuilder {
	builder.Service.Labels = labels
	return builder
}

func (builder *ServiceBuilder) WithSelector(selector map[string]string) *ServiceBuilder {
	builder.Service.Spec.Selector = selector
	return builder
}

func (builder *ServiceBuilder) AddServicePort(port int32, targetPort int) *ServiceBuilder {
	builder.Service.Spec.Ports = append(builder.Service.Spec.Ports, corev1.ServicePort{
		Port:       port,
		TargetPort: intstr.FromInt(targetPort),
	})
	return builder
}
func (builder *ServiceBuilder) WithServicePorts(servicePorts []corev1.ServicePort) *ServiceBuilder {
	builder.Service.Spec.Ports = append(builder.Service.Spec.Ports, servicePorts...)
	return builder
}

// SvcClusterLocal returns the svc.cluster.local hostname for name and namespace.
func SvcClusterLocal(namespace, name string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace)
}
