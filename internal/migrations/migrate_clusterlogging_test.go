package migrations

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("Migrating ClusterLogging instance", func() {

	Context("Migrate Collection Spec", func() {
		var (
			cl        loggingv1.ClusterLoggingSpec
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
			fluentTuning = &loggingv1.FluentdForwarderSpec{
				InFile: &loggingv1.FluentdInFileSpec{},
				Buffer: &loggingv1.FluentdBufferSpec{},
			}
		)
		BeforeEach(func() {

			cl = loggingv1.ClusterLoggingSpec{
				Collection: &loggingv1.CollectionSpec{
					Logs: &loggingv1.LogCollectionSpec{
						Type: loggingv1.LogCollectionTypeFluentd,
						CollectorSpec: loggingv1.CollectorSpec{
							Resources:    resources,
							NodeSelector: nodeSelector,
							Tolerations:  tolerations,
						},
					},
				},
				Forwarder: &loggingv1.ForwarderSpec{Fluentd: fluentTuning},
			}
		})

		Context("when migrating forwarder and collection.logs to collection", func() {
			It("should return clusterlogging as-is when collection is not defined", func() {
				spec := loggingv1.ClusterLoggingSpec{}
				Expect(MigrateCollectionSpec(spec)).To(Equal(loggingv1.ClusterLoggingSpec{}))
			})

			It("should return clusterlogging as-is when collection defined with empty value", func() {
				spec := loggingv1.ClusterLoggingSpec{}
				spec.Collection = &loggingv1.CollectionSpec{}
				Expect(MigrateCollectionSpec(spec)).To(Equal(loggingv1.ClusterLoggingSpec{Collection: &loggingv1.CollectionSpec{}}))
			})

			Context("when new collection fields are not set", func() {
				It("should move deprecated fields", func() {
					Expect(MigrateCollectionSpec(cl)).To(Equal(loggingv1.ClusterLoggingSpec{
						Collection: &loggingv1.CollectionSpec{
							Type: loggingv1.LogCollectionTypeFluentd,
							CollectorSpec: loggingv1.CollectorSpec{
								Resources:    resources,
								NodeSelector: nodeSelector,
								Tolerations:  tolerations,
							},
							Fluentd: fluentTuning,
						},
					}))
				})
			})

			Context("when new collection fields are set", func() {
				It("should ignore deprecated fields", func() {

					cl.Collection.Type = loggingv1.LogCollectionTypeFluentd

					Expect(MigrateCollectionSpec(cl)).To(Equal(loggingv1.ClusterLoggingSpec{
						Collection: &loggingv1.CollectionSpec{
							Type:    loggingv1.LogCollectionTypeFluentd,
							Fluentd: fluentTuning,
						},
					}))
				})

			})
		})
	})

	Context("Migrate Visualization Spec", func() {
		var visSpec loggingv1.VisualizationSpec

		BeforeEach(func() {
			visSpec = loggingv1.VisualizationSpec{
				Type:   loggingv1.VisualizationTypeKibana,
				Kibana: &loggingv1.KibanaSpec{},
			}
		})

		It("Should not migrate if vis type is not Kibana ", func() {
			visSpec.Type = loggingv1.VisualizationTypeOCPConsole
			visSpec.Kibana = nil

			newVisSpec := MigrateVisualizationSpec(visSpec)

			Expect(newVisSpec.NodeSelector).To(BeNil())
			Expect(newVisSpec.Tolerations).To(BeNil())
			Expect(newVisSpec).To(Equal(visSpec))

		})

		It("Should migrate kibana nodeSelectors only", func() {
			visSpec.Kibana.NodeSelector = map[string]string{"test": "test"}

			Expect(visSpec.NodeSelector).To(BeNil(), "Exp. no defined nodeSelector at top level")

			newSpec := MigrateVisualizationSpec(visSpec)

			Expect(newSpec.NodeSelector).ToNot(BeNil(), "Exp. nodeSelector at top level")
			Expect(newSpec.Tolerations).To(BeNil(), "Exp. no defined Tolerations at top level")
			Expect(newSpec.NodeSelector).To(Equal(visSpec.Kibana.NodeSelector))
		})

		It("Should migrate kibana tolerations", func() {
			visSpec.Kibana.Tolerations = []corev1.Toleration{
				{
					Key:   "test",
					Value: "test",
				},
			}

			Expect(visSpec.Tolerations).To(BeNil(), "Exp. no defined Tolerations at top level")
			newSpec := MigrateVisualizationSpec(visSpec)
			Expect(newSpec.NodeSelector).To(BeNil(), "Exp. no defined nodeSelector at top level")
			Expect(newSpec.Tolerations).ToNot(BeNil(), "Exp. Tolerations at top level")
			Expect(newSpec.Tolerations).To(Equal(visSpec.Kibana.Tolerations))
		})

		It("Should migrate kibana nodeSelectors and tolerations", func() {
			visSpec.Kibana.NodeSelector = map[string]string{"node1": "val1"}
			visSpec.Kibana.Tolerations = []corev1.Toleration{
				{
					Key:   "tol1",
					Value: "val1",
				},
			}

			Expect(visSpec.NodeSelector).To(BeNil(), "Exp. no defined nodeSelector at top level")
			Expect(visSpec.Tolerations).To(BeNil(), "Exp. no defined Tolerations at top level")

			newVisSpec := MigrateVisualizationSpec(visSpec)

			Expect(newVisSpec.NodeSelector).ToNot(BeNil(), "Exp. nodeSelector at top level")
			Expect(newVisSpec.Tolerations).ToNot(BeNil(), "Exp. Tolerations at top level")

			Expect(newVisSpec.NodeSelector).To(Equal(visSpec.Kibana.NodeSelector))
			Expect(newVisSpec.Tolerations).To(Equal(visSpec.Kibana.Tolerations))
		})

	})

	Context("Migrate ClusterLogging Spec", func() {
		var (
			cl        loggingv1.ClusterLoggingSpec
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
			nodeSelector = map[string]string{"clnode1": "clnode2"}
			tolerations  = []corev1.Toleration{
				{Key: "cltol1", Operator: corev1.TolerationOpExists, Value: "cltol2", Effect: corev1.TaintEffectNoExecute},
			}
			fluentTuning = &loggingv1.FluentdForwarderSpec{
				InFile: &loggingv1.FluentdInFileSpec{},
				Buffer: &loggingv1.FluentdBufferSpec{},
			}
		)
		BeforeEach(func() {

			cl = loggingv1.ClusterLoggingSpec{
				Collection: &loggingv1.CollectionSpec{
					Logs: &loggingv1.LogCollectionSpec{
						Type: loggingv1.LogCollectionTypeFluentd,
						CollectorSpec: loggingv1.CollectorSpec{
							Resources:    resources,
							NodeSelector: nodeSelector,
							Tolerations:  tolerations,
						},
					},
				},
				Forwarder: &loggingv1.ForwarderSpec{Fluentd: fluentTuning},
				Visualization: &loggingv1.VisualizationSpec{
					Type: loggingv1.VisualizationTypeKibana,
					Kibana: &loggingv1.KibanaSpec{
						Tolerations: []corev1.Toleration{
							{
								Key:   "key1",
								Value: "val1",
							},
						},
						NodeSelector: map[string]string{"node1": "val1"},
					},
				},
			}
		})

		It("Should migrate ClusterLogging spec's collection and visualization", func() {
			cl = MigrateClusterLogging(cl)
			Expect(cl.Collection).To(Equal(&loggingv1.CollectionSpec{
				Type: loggingv1.LogCollectionTypeFluentd,
				CollectorSpec: loggingv1.CollectorSpec{
					Resources:    resources,
					NodeSelector: nodeSelector,
					Tolerations:  tolerations,
				},
				Fluentd: fluentTuning,
			}))

			Expect(cl.Visualization).To(Equal(&loggingv1.VisualizationSpec{
				Type:         loggingv1.VisualizationTypeKibana,
				NodeSelector: map[string]string{"node1": "val1"},
				Tolerations: []corev1.Toleration{
					{
						Key:   "key1",
						Value: "val1",
					},
				},
				Kibana: &loggingv1.KibanaSpec{
					Tolerations: []corev1.Toleration{
						{
							Key:   "key1",
							Value: "val1",
						},
					},
					NodeSelector: map[string]string{"node1": "val1"},
				},
			}))
		})

		It("Should not migrate visualization if not defined", func() {
			cl.Visualization = nil
			cl = MigrateClusterLogging(cl)
			Expect(cl.Collection).To(Equal(&loggingv1.CollectionSpec{
				Type: loggingv1.LogCollectionTypeFluentd,
				CollectorSpec: loggingv1.CollectorSpec{
					Resources:    resources,
					NodeSelector: nodeSelector,
					Tolerations:  tolerations,
				},
				Fluentd: fluentTuning,
			}))

			Expect(cl.Visualization).To(BeNil(), "Visualization spec should still be nil")
		})
	})

})
