package pipelines

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("Pipeline validation #verifyHostNameNotFilteredForGCL", func() {
	var (
		gclOutput = obs.OutputSpec{
			Name: "gcl-out",
			Type: obs.OutputTypeGoogleCloudLogging,
			GoogleCloudLogging: &obs.GoogleCloudLogging{
				ID: obs.GoogleCloudLoggingId{
					Type:  obs.GoogleCloudLoggingIdTypeBillingAccount,
					Value: "billingAccountID",
				},
			},
		}
		pruneHost = obs.FilterSpec{
			Name: "prune",
			Type: obs.FilterTypePrune,
		}
		pipelineSpec = obs.PipelineSpec{
			Name:       "gclPruneHost",
			OutputRefs: []string{gclOutput.Name},
			InputRefs:  []string{string(obs.InputTypeApplication)},
			FilterRefs: []string{pruneHost.Name},
		}
		outputs = map[string]obs.OutputSpec{
			gclOutput.Name: gclOutput,
		}
		filters = map[string]*obs.FilterSpec{}
	)

	DescribeTable("should return empty", func(pruneSpec obs.PruneFilterSpec) {
		gclOutput := obs.OutputSpec{
			Name: "gcl-out",
			Type: obs.OutputTypeGoogleCloudLogging,
			GoogleCloudLogging: &obs.GoogleCloudLogging{
				ID: obs.GoogleCloudLoggingId{
					Type:  obs.GoogleCloudLoggingIdTypeBillingAccount,
					Value: "billingAccountID",
				},
			},
		}

		pruneHost = obs.FilterSpec{
			Name:            "prune",
			Type:            obs.FilterTypePrune,
			PruneFilterSpec: &pruneSpec,
		}

		filters = map[string]*obs.FilterSpec{
			pruneHost.Name: &pruneHost,
		}
		outputs := map[string]obs.OutputSpec{
			gclOutput.Name: gclOutput,
		}

		cond := verifyHostNameNotFilteredForGCL(pipelineSpec, outputs, filters)
		Expect(cond).To(BeEmpty())
	},
		Entry("when `in` does not include .hostname", obs.PruneFilterSpec{In: []obs.FieldPath{".foo"}}),
		Entry("when `notIn` includes .hostname", obs.PruneFilterSpec{NotIn: []obs.FieldPath{".hostname"}}),
		Entry("when `in` does not include and `notIn` includes .hostname", obs.PruneFilterSpec{In: []obs.FieldPath{".foo"}, NotIn: []obs.FieldPath{".hostname"}}))

	It("should not return empty when prune filters `.hostname` for pipeline without GCL output", func() {
		pruneHost := obs.FilterSpec{
			Name: "prune",
			Type: obs.FilterTypePrune,
			PruneFilterSpec: &obs.PruneFilterSpec{
				In:    []obs.FieldPath{".foo, .hostname"},
				NotIn: []obs.FieldPath{".foo"},
			},
		}
		filters[pruneHost.Name] = &pruneHost

		cond := verifyHostNameNotFilteredForGCL(pipelineSpec, outputs, filters)
		Expect(cond).To(Not(BeEmpty()))
	})

	It("should return empty when prune filters `.hostname` for pipeline without GCL output", func() {
		esOutput := obs.OutputSpec{

			Name:          "myOutput",
			Type:          "elasticsearch",
			Elasticsearch: &obs.Elasticsearch{},
		}

		pruneHost := obs.FilterSpec{
			Name: "prune",
			Type: obs.FilterTypePrune,
			PruneFilterSpec: &obs.PruneFilterSpec{
				In:    []obs.FieldPath{".foo, .hostname"},
				NotIn: []obs.FieldPath{".foo"},
			},
		}
		outputs = map[string]obs.OutputSpec{
			esOutput.Name: esOutput,
		}
		filters[pruneHost.Name] = &pruneHost

		cond := verifyHostNameNotFilteredForGCL(pipelineSpec, outputs, filters)
		Expect(cond).To(BeEmpty())
	})
})
