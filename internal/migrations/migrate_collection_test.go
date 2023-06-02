package migrations

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("Migrating ClusterLogging instance", func() {
	var (
		cl        ClusterLoggingSpec
		resources = &corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("2Gi"),
				corev1.ResourceCPU:    resource.MustParse("2"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("1Gi"),
				corev1.ResourceCPU:    resource.MustParse("1"),
			},
		}
		nodeSelector = map[string]string{"foo": "bar"}
		tolerations  = []corev1.Toleration{
			{Key: "foo", Operator: corev1.TolerationOpExists, Value: "bar", Effect: corev1.TaintEffectNoExecute},
		}
		fluentTuning = &FluentdForwarderSpec{
			InFile: &FluentdInFileSpec{},
			Buffer: &FluentdBufferSpec{},
		}
	)
	BeforeEach(func() {

		cl = ClusterLoggingSpec{
			Collection: &CollectionSpec{
				Logs: &LogCollectionSpec{
					Type: LogCollectionTypeFluentd,
					CollectorSpec: CollectorSpec{
						Resources:    resources,
						NodeSelector: nodeSelector,
						Tolerations:  tolerations,
					},
				},
			},
			Forwarder: &ForwarderSpec{Fluentd: fluentTuning},
		}
	})

	Context("when migrating forwarder and collection.logs to collection", func() {
		It("should return clusterlogging as-is when collection is not defined", func() {
			spec := ClusterLoggingSpec{}
			Expect(MigrateCollectionSpec(spec)).To(Equal(ClusterLoggingSpec{}))
		})

		It("should return clusterlogging as-is when collection defined with empty value", func() {
			spec := ClusterLoggingSpec{}
			spec.Collection = &CollectionSpec{}
			Expect(MigrateCollectionSpec(spec)).To(Equal(ClusterLoggingSpec{Collection: &CollectionSpec{}}))
		})

		Context("when new collection fields are not set", func() {
			It("should move deprecated fields", func() {
				Expect(MigrateCollectionSpec(cl)).To(Equal(ClusterLoggingSpec{
					Collection: &CollectionSpec{
						Type: LogCollectionTypeFluentd,
						CollectorSpec: CollectorSpec{
							Resources:    resources,
							NodeSelector: nodeSelector,
							Tolerations:  tolerations,
						},
						Fluentd: fluentTuning,
					},
				}))
			})
		})

		Context("when new collection and deprecated fields are set", func() {
			It("should prefer deprecated fields", func() {

				cl.Collection.Type = LogCollectionTypeFluentd
				cl.Collection.Resources = &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("1000Gi"),
						corev1.ResourceCPU:    resource.MustParse("4"),
					},
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("15Gi"),
						corev1.ResourceCPU:    resource.MustParse("18"),
					},
				}
				cl.Collection.Tolerations = []corev1.Toleration{
					{Key: "abc", Operator: corev1.TolerationOpExists, Value: "bar", Effect: corev1.TaintEffectNoExecute},
				}
				cl.Collection.NodeSelector = map[string]string{"ignorme": "yes"}

				Expect(MigrateCollectionSpec(cl)).To(Equal(ClusterLoggingSpec{
					Collection: &CollectionSpec{
						Type: LogCollectionTypeFluentd,
						CollectorSpec: CollectorSpec{
							Resources:    resources,
							NodeSelector: nodeSelector,
							Tolerations:  tolerations,
						},
						Fluentd: fluentTuning,
					},
				}))
			})

		})

	})

})
