package api

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

var _ = Describe("#ConvertPipelines", func() {
	Context("references default function", func() {
		It("should be true if a pipeline references default output", func() {
			outputs := []string{"es-out", "default", "cw"}
			Expect(referencesDefaultOutput(outputs)).To(BeTrue())
		})
		It("should be false if a pipeline does not reference default output", func() {
			outputs := []string{"es-out", "cw"}
			Expect(referencesDefaultOutput(outputs)).To(BeFalse())
		})
	})

	DescribeTable("generate pipeline filter function", func(loggingPipelineSpec logging.PipelineSpec, expObsFilters []obs.FilterSpec, expObsFilterRefs []string) {
		actObsFilterSpec, actFilterRefs := generatePipelineFilters(loggingPipelineSpec, &sets.String{})
		Expect(actObsFilterSpec).To(Equal(expObsFilters))
		Expect(actFilterRefs).To(Equal(expObsFilterRefs))
	},
		Entry("should generate detect multiline error as a filter and filterRef", logging.PipelineSpec{
			Name:                  "detectPipeline",
			DetectMultilineErrors: true,
		}, []obs.FilterSpec{
			{
				Name: detectMultilineErrorFilterName,
				Type: obs.FilterTypeDetectMultiline,
			},
		},
			[]string{detectMultilineErrorFilterName}),
		Entry("should generate parse json as a filter and filterRef", logging.PipelineSpec{
			Name:  "parsePipeline",
			Parse: "json",
		}, []obs.FilterSpec{
			{
				Name: parseFilterName,
				Type: obs.FilterTypeParse,
			},
		},
			[]string{parseFilterName}),
		Entry("should generate labels as a filter and filterRef", logging.PipelineSpec{
			Name:   "labelPipeline",
			Labels: map[string]string{"foo": "bar"},
		}, []obs.FilterSpec{
			{
				Name:            "filter-labelPipeline-openShiftLabels",
				Type:            obs.FilterTypeOpenshiftLabels,
				OpenShiftLabels: map[string]string{"foo": "bar"},
			},
		},
			[]string{"filter-labelPipeline-openShiftLabels"}),
		Entry("should generate all pipeline filters and filterRefs", logging.PipelineSpec{
			Name:                  "filterPipeline",
			Labels:                map[string]string{"foo": "bar"},
			Parse:                 "json",
			DetectMultilineErrors: true,
		}, []obs.FilterSpec{
			{
				Name: detectMultilineErrorFilterName,
				Type: obs.FilterTypeDetectMultiline,
			},
			{
				Name: parseFilterName,
				Type: obs.FilterTypeParse,
			},
			{
				Name:            "filter-filterPipeline-openShiftLabels",
				Type:            obs.FilterTypeOpenshiftLabels,
				OpenShiftLabels: map[string]string{"foo": "bar"},
			},
		},
			[]string{detectMultilineErrorFilterName, parseFilterName, "filter-filterPipeline-openShiftLabels"}),
	)

	Context("pipeline with default reference", func() {
		It("should convert pipeline with logstore es and add default-es to outputRefs", func() {
			logStoreSpec := &logging.LogStoreSpec{
				Type: logging.LogStoreTypeElasticsearch,
			}

			loggingClfSpec := logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						Name:       "my-app-audit",
						InputRefs:  []string{logging.InputNameApplication, logging.InputNameAudit},
						OutputRefs: []string{"es-out", "default"},
					},
				},
			}

			expObsPipelineSpec := []obs.PipelineSpec{
				{
					Name:       "my-app-audit",
					InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeAudit)},
					OutputRefs: []string{"es-out", "default-elasticsearch"},
				},
			}

			actPipelineSpec, _, needDefault := convertPipelines(logStoreSpec, &loggingClfSpec)
			Expect(actPipelineSpec).To(Equal(expObsPipelineSpec))
			Expect(needDefault).To(BeTrue())
		})
		It("should convert pipeline with logstore lokistack and add default-lokistack to outputRef", func() {
			logStoreSpec := &logging.LogStoreSpec{
				Type: logging.LogStoreTypeLokiStack,
				LokiStack: logging.LokiStackStoreSpec{
					Name: "my-lokistack",
				},
			}

			loggingClfSpec := logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						Name:       "my-app-audit",
						InputRefs:  []string{logging.InputNameApplication, logging.InputNameAudit},
						OutputRefs: []string{"es-out", "default"},
					},
				},
			}
			expPipelineSpec := []obs.PipelineSpec{
				{
					Name:       "my-app-audit",
					InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeAudit)},
					OutputRefs: []string{"es-out", "default-lokistack"},
				},
			}

			actPipelineSpec, _, needDefault := convertPipelines(logStoreSpec, &loggingClfSpec)
			Expect(actPipelineSpec).To(Equal(expPipelineSpec))
			Expect(needDefault).To(BeTrue())
		})
		It("should convert pipelines that do not reference default", func() {
			loggingClfSpec := logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						Name:       "my-app",
						InputRefs:  []string{logging.InputNameApplication},
						OutputRefs: []string{"es-out", "foo", "bar"},
					},
					{
						Name:       "my-infra-audit",
						InputRefs:  []string{logging.InputNameInfrastructure, logging.InputNameAudit},
						OutputRefs: []string{"es-out", "foo", "baz"},
					},
				},
			}
			expPipelineSpec := []obs.PipelineSpec{
				{
					Name:       "my-app",
					InputRefs:  []string{string(obs.InputTypeApplication)},
					OutputRefs: []string{"es-out", "foo", "bar"},
				},
				{
					Name:       "my-infra-audit",
					InputRefs:  []string{string(obs.InputTypeInfrastructure), string(obs.InputTypeAudit)},
					OutputRefs: []string{"es-out", "foo", "baz"},
				},
			}

			actPipelineSpec, _, needDefault := convertPipelines(nil, &loggingClfSpec)
			Expect(actPipelineSpec).To(Equal(expPipelineSpec))
			Expect(needDefault).To(BeFalse())
		})
	})

	It("should convert all logging pipelines to observability pipelines", func() {
		logStoreSpec := &logging.LogStoreSpec{
			Type: logging.LogStoreTypeLokiStack,
			LokiStack: logging.LokiStackStoreSpec{
				Name: "my-lokistack",
			},
		}
		loggingClfSpec := logging.ClusterLogForwarderSpec{
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "my-app",
					InputRefs:  []string{logging.InputNameApplication},
					OutputRefs: []string{"es-out", "foo", "bar"},
					Parse:      "json",
				},
				{
					Name:       "my-infra-audit",
					InputRefs:  []string{logging.InputNameInfrastructure, logging.InputNameAudit},
					OutputRefs: []string{"es-out", "foo", "baz"},
					Parse:      "json",
				},
				{
					Name:       "my-app-default",
					InputRefs:  []string{logging.InputNameApplication},
					OutputRefs: []string{"default"},
					Labels:     map[string]string{"foo": "bar"},
				},
			},
		}
		expPipelineFilterSpecs := []obs.FilterSpec{
			{
				Name: parseFilterName,
				Type: obs.FilterTypeParse,
			},
			{
				Name:            "filter-my-app-default-openShiftLabels",
				Type:            obs.FilterTypeOpenshiftLabels,
				OpenShiftLabels: map[string]string{"foo": "bar"},
			},
		}
		expPipelineSpec := []obs.PipelineSpec{
			{
				Name:       "my-app",
				InputRefs:  []string{string(obs.InputTypeApplication)},
				OutputRefs: []string{"es-out", "foo", "bar"},
				FilterRefs: []string{parseFilterName},
			},
			{
				Name:       "my-infra-audit",
				InputRefs:  []string{string(obs.InputTypeInfrastructure), string(obs.InputTypeAudit)},
				OutputRefs: []string{"es-out", "foo", "baz"},
				FilterRefs: []string{parseFilterName},
			},
			{
				Name:       "my-app-default",
				InputRefs:  []string{logging.InputNameApplication},
				OutputRefs: []string{"default-lokistack"},
				FilterRefs: []string{"filter-my-app-default-openShiftLabels"},
			},
		}

		actPipelineSpec, filterSpecs, needDefault := convertPipelines(logStoreSpec, &loggingClfSpec)
		Expect(actPipelineSpec).To(Equal(expPipelineSpec))
		Expect(filterSpecs).To(Equal(expPipelineFilterSpecs))
		Expect(needDefault).To(BeTrue())
	})

})
