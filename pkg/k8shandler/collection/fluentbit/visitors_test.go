package fluentbit

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test/matchers"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ = Describe("Collection fluentbit visitors", func() {
	BeforeEach(func() {
		os.Setenv("IMAGE_FLUENTBIT", "ATESTVALE")
	})
	Context("#VisitPodSpec", func() {

		It("Should add a collector container", func() {
			podSpec := &v1.PodSpec{}
			VisitPodSpec(podSpec)
			Expect(len(podSpec.Containers)).To(Equal(1), "Exp more containers to be added to the pod")
			container := podSpec.Containers[0]
			Expect(*(container.SecurityContext.Privileged)).To(BeTrue())

			Expect(container.VolumeMounts).To(IncludeVolumeMount(v1.VolumeMount{Name: "varlog", MountPath: "/var/log", ReadOnly: true}))
			Expect(container.VolumeMounts).To(IncludeVolumeMount(v1.VolumeMount{Name: "config", MountPath: "/etc/fluent-bit", ReadOnly: true}))

			Expect(container.Ports).To(IncludeContainerPort(v1.ContainerPort{Name: "cmetrics", ContainerPort: 2020, Protocol: v1.ProtocolTCP}))
			Expect(len(container.Ports[0].Name) <= 15).To(BeTrue(), "Container port names must be less then 15 charachters")

			Expect(container.Env).To(IncludeEnvVar(v1.EnvVar{Name: "POD_IP", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIP"}}}))
		})
	})

	Context("#VisitService", func() {
		It("should expose the metrics port of the collector container", func() {
			service := &v1.Service{}
			VisitService(service)
			Expect(service.Spec.Ports).To(IncludeServicePort(v1.ServicePort{Name: "cmetrics", Port: 2020, TargetPort: intstr.FromString(constants.CollectorMetricsPortName)}))
		})
	})
	Context("VisitServiceMonitor", func() {
		It("should add monitoring of the collector container", func() {
			sm := &monitoringv1.ServiceMonitor{}
			VisitServiceMonitor(sm)
			Expect(sm.Spec.Endpoints).To(IncludeMetricsEndpoint(monitoringv1.Endpoint{
				Port:   "cmetrics",
				Path:   "/api/v1/metrics/prometheus",
				Scheme: "http",
			}))
		})
	})
})
