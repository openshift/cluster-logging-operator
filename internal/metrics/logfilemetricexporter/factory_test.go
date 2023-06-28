package logfilemetricexporter

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	loggingv1a1 "github.com/openshift/cluster-logging-operator/apis/logging/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/tls"
)

var _ = Describe("LogFileMetricExporter functions", func() {
	var (
		cluster = &loggingv1.ClusterLogging{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.SingletonName,
				Namespace: constants.OpenshiftNS,
			},
			Spec: loggingv1.ClusterLoggingSpec{
				ManagementState: loggingv1.ManagementStateManaged,
				LogStore: &loggingv1.LogStoreSpec{
					Type: loggingv1.LogStoreTypeElasticsearch,
				},
				Collection: &loggingv1.CollectionSpec{
					Type:          loggingv1.LogCollectionTypeFluentd,
					CollectorSpec: loggingv1.CollectorSpec{},
				},
			},
		}

		// Adding ns and label to account for addSecurityLabelsToNamespace() added in LOG-2620
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"test": "true"},
				Name:   cluster.Namespace,
			},
		}

		reqClient = fake.NewFakeClient( //nolint
			cluster,
			namespace,
		)

		logFileMetricExporter = &loggingv1a1.LogFileMetricExporter{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.SingletonName,
				Namespace: constants.WatchNamespace,
			},
		}

		expectedPort = corev1.ContainerPort{
			Name:          "exporter-port",
			ContainerPort: int32(2112),
			Protocol:      "TCP",
		}

		tlsProfile, _  = tls.FetchAPIServerTlsProfile(reqClient)
		tlsProfileSpec = tls.GetClusterTLSProfileSpec(tlsProfile)
	)

	Context("new exporter container", func() {

		It("should spec a new container when no specs specified in custom resource", func() {
			exporterContainer := newLogMetricsExporterContainer(*logFileMetricExporter, tlsProfileSpec)

			Expect(exporterContainer.Name).To(Equal(constants.LogfilesmetricexporterName))

			// Container's port
			Expect(exporterContainer.Ports).ToNot(BeEmpty(), "Exp at least 1 containerPort in list")
			Expect(exporterContainer.Ports).To(HaveLen(1), "Exp. only 1 containerPort speced")
			Expect(exporterContainer.Ports).To(ContainElement(expectedPort))

			// Volume mounts
			Expect(exporterContainer.VolumeMounts).ToNot(BeEmpty(), "Exp to contain volume mounts")
			Expect(exporterContainer.VolumeMounts).To(HaveLen(3), "Expect exactly 3 volume mounts")

			// No resources
			Expect(exporterContainer.Resources.Limits).To(BeEmpty(), "Exp no limits")
			Expect(exporterContainer.Resources.Requests).To(BeEmpty(), "Exp no requests")
		})

		It("should spec a new container with specified resource requests/limits", func() {
			resMemory := resource.MustParse("100Gi")
			resCPU := resource.MustParse("500m")

			logFileMetricExporter.Spec = loggingv1a1.LogFileMetricExporterSpec{
				Resources: &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resCPU,
						corev1.ResourceMemory: resMemory,
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resCPU,
						corev1.ResourceMemory: resMemory,
					},
				},
			}

			exporterContainer := newLogMetricsExporterContainer(*logFileMetricExporter, tlsProfileSpec)

			Expect(exporterContainer.Name).To(Equal(constants.LogfilesmetricexporterName))

			// Resources.Requests
			Expect(exporterContainer.Resources.Requests).ToNot(BeEmpty(), "Exp requests")
			Expect(exporterContainer.Resources.Requests.Cpu().Cmp(resCPU)).To(Equal(0))
			Expect(exporterContainer.Resources.Requests.Memory().Cmp(resMemory)).To(Equal(0))
			// Resources.Limits
			Expect(exporterContainer.Resources.Limits).ToNot(BeEmpty(), "Exp requests")
			Expect(exporterContainer.Resources.Limits.Cpu().Cmp(resCPU)).To(Equal(0))
			Expect(exporterContainer.Resources.Limits.Memory().Cmp(resMemory)).To(Equal(0))

		})
	})

	Context("new exporter podspec", func() {
		It("should spec a new default pod with default tolerations and NodeSelector", func() {
			podSpec := NewPodSpec(*logFileMetricExporter, tlsProfileSpec)
			Expect(podSpec.NodeSelector).ToNot(BeEmpty())
			Expect(podSpec.NodeSelector).To(HaveLen(1))
			Expect(podSpec.NodeSelector).To(HaveKeyWithValue("kubernetes.io/os", "linux"))

			// Tolerations
			Expect(podSpec.Tolerations).To(Equal(constants.DefaultTolerations()))
		})

		It("should spec a new pod with defined tolerations and NodeSelector", func() {
			testTol1 := corev1.Toleration{
				Key:      "test/key",
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectNoSchedule,
			}

			logFileMetricExporter.Spec.Tolerations = []corev1.Toleration{
				testTol1,
			}

			expectedTolerations := []corev1.Toleration{
				testTol1,
			}
			expectedTolerations = append(expectedTolerations, constants.DefaultTolerations()...)

			testNodeSelect := map[string]string{
				"testNode": "testval",
				"test2":    "someOtherVal",
			}

			logFileMetricExporter.Spec.NodeSelector = testNodeSelect

			podSpec := NewPodSpec(*logFileMetricExporter, tlsProfileSpec)
			Expect(podSpec.Containers).ToNot(BeEmpty())

			// NodeSelector
			Expect(podSpec.NodeSelector).ToNot(BeEmpty())
			Expect(podSpec.NodeSelector).To(HaveLen(3))
			Expect(podSpec.NodeSelector).To(HaveKeyWithValue("kubernetes.io/os", "linux"))
			Expect(podSpec.NodeSelector).To(HaveKeyWithValue("test2", "someOtherVal"))

			// Tolerations
			Expect(podSpec.Tolerations).To(ContainElement(testTol1))
			Expect(podSpec.Tolerations).To(Equal(expectedTolerations))
		})

		It("should not have duplicate tolerations in the pod spec", func() {
			testTol1 := corev1.Toleration{
				Key:      "node-role.kubernetes.io/master",
				Operator: corev1.TolerationOpEqual,
				Effect:   corev1.TaintEffectNoSchedule,
			}

			logFileMetricExporter.Spec.Tolerations = []corev1.Toleration{
				testTol1,
			}

			expectedTolerations := []corev1.Toleration{
				testTol1,
				{
					Key:      "node.kubernetes.io/disk-pressure",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			}

			podSpec := NewPodSpec(*logFileMetricExporter, tlsProfileSpec)
			Expect(podSpec.Containers).ToNot(BeEmpty())

			// Tolerations
			Expect(podSpec.Tolerations).To(ContainElement(testTol1))
			Expect(podSpec.Tolerations).To(Equal(expectedTolerations))
		})
	})
})
