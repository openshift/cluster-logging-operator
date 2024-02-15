package clusterlogging

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/api/logging/v1"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Migrating ClusterLogging instance", func() {

	Context("when migrating visualization", func() {
		It("should return clusterlogging as-is when visualization and LogStore are not defined", func() {
			spec := ClusterLoggingSpec{}
			Expect(MigrateVisualizationSpec(spec)).To(Equal(ClusterLoggingSpec{}))
		})

		It("should return clusterlogging as-is when visualization defined with empty value and LogStore is not defined", func() {
			spec := ClusterLoggingSpec{}
			spec.Visualization = &VisualizationSpec{}
			Expect(MigrateVisualizationSpec(spec)).To(Equal(ClusterLoggingSpec{Visualization: &VisualizationSpec{}}))
		})

		It("should return clusterlogging with visualization as ocp-console when LogStore defined with lokistack and visualisation with nil", func() {
			spec := ClusterLoggingSpec{}
			spec.LogStore = &LogStoreSpec{
				Type:      LogStoreTypeLokiStack,
				LokiStack: LokiStackStoreSpec{},
			}
			Expect(MigrateVisualizationSpec(spec)).To(Equal(ClusterLoggingSpec{
				LogStore: &LogStoreSpec{
					Type:      LogStoreTypeLokiStack,
					LokiStack: LokiStackStoreSpec{},
				},
				Visualization: &VisualizationSpec{
					Type:       VisualizationTypeOCPConsole,
					OCPConsole: &OCPConsoleSpec{},
				}}))
		})

		It("should return clusterlogging as is when LogStore defined with Elasticsearch", func() {
			spec := ClusterLoggingSpec{}
			spec.LogStore = &LogStoreSpec{
				Type:          LogStoreTypeElasticsearch,
				Elasticsearch: &ElasticsearchSpec{},
			}
			Expect(MigrateVisualizationSpec(spec)).To(Equal(ClusterLoggingSpec{
				LogStore: &LogStoreSpec{
					Type:          LogStoreTypeElasticsearch,
					Elasticsearch: &ElasticsearchSpec{},
				},
			}))
		})

		It("should return clusterlogging as is when visualization is defined with not empty value", func() {
			spec := ClusterLoggingSpec{}
			spec.LogStore = &LogStoreSpec{
				Type:      LogStoreTypeLokiStack,
				LokiStack: LokiStackStoreSpec{},
			}
			spec.Visualization = &VisualizationSpec{
				Type:   VisualizationTypeKibana,
				Kibana: &KibanaSpec{},
			}
			Expect(MigrateVisualizationSpec(spec)).To(Equal(ClusterLoggingSpec{
				LogStore: &LogStoreSpec{
					Type:      LogStoreTypeLokiStack,
					LokiStack: LokiStackStoreSpec{},
				},
				Visualization: &VisualizationSpec{
					Type:   VisualizationTypeKibana,
					Kibana: &KibanaSpec{},
				},
			}))
		})

		Context("when migrating kibana specs", func() {
			It("should migrate kibana nodeselectors and tolerations to top level spec.visualization", func() {
				spec := ClusterLoggingSpec{
					Visualization: &VisualizationSpec{
						Type: VisualizationTypeKibana,
						Kibana: &KibanaSpec{
							NodeSelector: map[string]string{
								"test": "test",
							},
							Tolerations: []v1.Toleration{
								{
									Key:   "test",
									Value: "test",
								},
							},
						},
					},
				}
				Expect(MigrateVisualizationSpec(spec)).To(Equal(ClusterLoggingSpec{
					Visualization: &VisualizationSpec{
						Type: VisualizationTypeKibana,
						NodeSelector: map[string]string{
							"test": "test",
						},
						Tolerations: []v1.Toleration{
							{
								Key:   "test",
								Value: "test",
							},
						},
						Kibana: &KibanaSpec{
							NodeSelector: map[string]string{
								"test": "test",
							},
							Tolerations: []v1.Toleration{
								{
									Key:   "test",
									Value: "test",
								},
							},
						},
					}}))
			})
		})
	})

})
