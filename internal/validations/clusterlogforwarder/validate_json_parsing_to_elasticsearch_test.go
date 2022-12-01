package clusterlogforwarder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

var _ = Describe("[internal][validations] ClusterLogForwarder", func() {
	var (
		clf      *v1.ClusterLogForwarder
		pipeline v1.PipelineSpec
		es       *v1.Elasticsearch
	)
	BeforeEach(func() {
		es = &v1.Elasticsearch{}
		pipeline = v1.PipelineSpec{
			InputRefs:  []string{string(v1.InputNameApplication)},
			OutputRefs: []string{"anOutput"},
			Parse:      "json",
		}
		clf = &v1.ClusterLogForwarder{
			Spec: v1.ClusterLogForwarderSpec{
				Outputs: []v1.OutputSpec{
					{
						Name: "anOutput",
						Type: v1.OutputTypeElasticsearch,
						OutputTypeSpec: v1.OutputTypeSpec{
							Elasticsearch: es,
						},
					},
				},
				Pipelines: []v1.PipelineSpec{
					pipeline,
				},
			},
		}
	})

	Context("#validateJsonParsingToElasticsearch", func() {

		It("should fail validation when the pipeline includes Elasticsearch and structuredTypeKey or structuredTypeName is missing", func() {
			Expect(validateJsonParsingToElasticsearch(*clf)).To(Not(Succeed()))
		})
		It("should pass validation when the pipeline includes Elasticsearch and structuredTypeName is spec'd", func() {
			es.StructuredTypeName = "foo"
			clf.Spec.Outputs[0].Elasticsearch = es
			Expect(validateJsonParsingToElasticsearch(*clf)).To(Succeed())
		})
		It("should pass validation when the pipeline includes Elasticsearch and structuredTypeKey is spec'd", func() {
			es.StructuredTypeKey = "foo"
			clf.Spec.Outputs[0].Elasticsearch = es
			Expect(validateJsonParsingToElasticsearch(*clf)).To(Succeed())
		})
		It("should pass validation when the pipeline includes Elasticsearch and OutputDefaults.StructuredTypeName is spec'd", func() {
			clf.Spec.OutputDefaults = &v1.OutputDefaults{Elasticsearch: &v1.ElasticsearchStructuredSpec{StructuredTypeName: "foo"}}
			Expect(validateJsonParsingToElasticsearch(*clf)).To(Succeed())
		})
		It("should pass validation when the pipeline includes Elasticsearch and and OutputDefaults.StructuredTypeKey is spec'd", func() {
			clf.Spec.OutputDefaults = &v1.OutputDefaults{Elasticsearch: &v1.ElasticsearchStructuredSpec{StructuredTypeKey: "foo"}}
			Expect(validateJsonParsingToElasticsearch(*clf)).To(Succeed())
		})
		It("should pass validation when the pipeline does not ref an Elasticsearch output type", func() {
			clf.Spec.Outputs[0].Type = v1.OutputTypeCloudwatch
			Expect(validateJsonParsingToElasticsearch(*clf)).To(Succeed())
		})

		It("should pass validation when the pipeline does not spec JSON parsing", func() {
			clf.Spec.Pipelines[0].Parse = ""
			Expect(validateJsonParsingToElasticsearch(*clf)).To(Succeed())
		})

	})

})
