package migrations

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

var _ = Describe("MigrateDefaultOutput", func() {

	var (
		pipelines []logging.PipelineSpec
		outputs   []logging.OutputSpec
		spec      logging.ClusterLogForwarderSpec
		esSpec    *logging.Elasticsearch
	)

	BeforeEach(func() {
		esSpec = &logging.Elasticsearch{
			ElasticsearchStructuredSpec: logging.ElasticsearchStructuredSpec{
				StructuredTypeKey: "foo.bar",
			},
		}
		pipelines = []logging.PipelineSpec{
			{
				Name:       "test",
				OutputRefs: []string{"first", "second"},
			},
		}
		outputs = []logging.OutputSpec{
			{
				Name: "first",
				Type: logging.OutputTypeElasticsearch,
				OutputTypeSpec: logging.OutputTypeSpec{
					Elasticsearch: esSpec,
				},
			},
		}
		spec = logging.ClusterLogForwarderSpec{
			Outputs:   outputs,
			Pipelines: pipelines,
		}
	})

	It("should not add the default OutputSpec when it is not referenced by a pipeline", func() {
		Expect(MigrateDefaultOutput(spec)).To(Equal(outputs))
	})

	Context("when a pipeline references 'default'", func() {

		var exp []logging.OutputSpec
		BeforeEach(func() {
			pipelines[0].OutputRefs = append(spec.Pipelines[0].OutputRefs, logging.OutputNameDefault)
			spec = logging.ClusterLogForwarderSpec{
				Outputs:   outputs,
				Pipelines: pipelines,
			}
		})

		Context("and outputs does not explicitly spec 'default'", func() {
			BeforeEach(func() {
				exp = append(outputs, NewDefaultOutput(nil))
			})

			It("should add the default OutputSpec", func() {
				Expect(MigrateDefaultOutput(spec)).To(Equal(exp), fmt.Sprintf("Exp. default output because of pipeline %v", pipelines))
			})
			It("should add the default OutputSpec and OutputDefaults when OutputDefaults are spec'd", func() {
				spec.OutputDefaults = &logging.OutputDefaults{
					Elasticsearch: &logging.ElasticsearchStructuredSpec{
						StructuredTypeKey: "foo.bar",
					},
				}
				exp[1].Elasticsearch = &logging.Elasticsearch{
					ElasticsearchStructuredSpec: *spec.OutputDefaults.Elasticsearch,
				}

				Expect(MigrateDefaultOutput(spec)).To(Equal(exp), fmt.Sprintf("Exp. default output because of pipeline %v and OutputDefault %v", pipelines, spec.OutputDefaults))
			})
		})

		Context("and outputs includes an OutputSpec named 'default'", func() {
			var tobereplaced logging.OutputSpec
			BeforeEach(func() {
				tobereplaced = logging.OutputSpec{
					Name:   logging.OutputNameDefault,
					Type:   logging.OutputTypeElasticsearch,
					URL:    "thiswillgetreplaced",
					Secret: &logging.OutputSecretSpec{Name: "replacem"},
				}

			})

			It("should replace the OutputSpec with the default OutputSpec", func() {
				exp = append(outputs, NewDefaultOutput(nil))
				spec = logging.ClusterLogForwarderSpec{
					Outputs:   append(outputs, tobereplaced),
					Pipelines: pipelines,
				}
				Expect(MigrateDefaultOutput(spec)).To(Equal(exp), fmt.Sprintf("Exp. default output because of pipeline %v", pipelines))
			})

			It("should replace the OutputSpec with the default OutputSpec and use the config (e.g. structureTypeKey) defined in the original OutputSpec", func() {
				tobereplaced.Elasticsearch = esSpec
				exp = append(outputs, NewDefaultOutput(&logging.OutputDefaults{Elasticsearch: &esSpec.ElasticsearchStructuredSpec}))
				spec = logging.ClusterLogForwarderSpec{
					Outputs:        append(outputs, tobereplaced),
					Pipelines:      pipelines,
					OutputDefaults: &logging.OutputDefaults{Elasticsearch: &logging.ElasticsearchStructuredSpec{StructuredTypeKey: "abc"}},
				}
				Expect(MigrateDefaultOutput(spec)).To(Equal(exp), fmt.Sprintf("Exp. default output because of pipeline %v and ElasticsearchSpec %v", pipelines, esSpec))
			})
		})

	})

})
