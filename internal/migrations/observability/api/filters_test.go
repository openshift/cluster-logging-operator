package api

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	v1 "k8s.io/apiserver/pkg/apis/audit/v1"
)

var _ = Describe("#ConvertFilters", func() {
	It("should map logging.KubeApiAudit to observability.KubeApiAudit", func() {
		loggingAudit := &logging.KubeAPIAudit{
			Rules: []v1.PolicyRule{{
				Level: "RequestResponse",
				Resources: []v1.GroupResources{
					{
						Group:     "",
						Resources: []string{"pods"},
					},
				},
			}},
			OmitStages:        []v1.Stage{v1.StageRequestReceived},
			OmitResponseCodes: &[]int{404, 409},
		}
		expObsAudit := &obs.KubeAPIAudit{
			Rules: []v1.PolicyRule{{
				Level: "RequestResponse",
				Resources: []v1.GroupResources{
					{
						Group:     "",
						Resources: []string{"pods"},
					},
				},
			}},
			OmitStages:        []v1.Stage{v1.StageRequestReceived},
			OmitResponseCodes: &[]int{404, 409},
		}

		Expect(mapKubeApiAuditFilter(*loggingAudit)).To(Equal(expObsAudit))
	})

	It("should map logging.DropTest to observability.DropTest", func() {
		loggingDropTest := []logging.DropTest{
			{
				DropConditions: []logging.DropCondition{
					{
						Field:   ".foo",
						Matches: "bar",
					},
					{
						Field:      ".foo2",
						NotMatches: "baz",
					},
				},
			},
			{
				DropConditions: []logging.DropCondition{
					{
						Field:   ".test",
						Matches: "foo",
					},
					{
						Field:   ".laz",
						Matches: "paz",
					},
				},
			},
		}
		expObsDropTest := []obs.DropTest{
			{
				DropConditions: []obs.DropCondition{
					{
						Field:   ".foo",
						Matches: "bar",
					},
					{
						Field:      ".foo2",
						NotMatches: "baz",
					},
				},
			},
			{
				DropConditions: []obs.DropCondition{
					{
						Field:   ".test",
						Matches: "foo",
					},
					{
						Field:   ".laz",
						Matches: "paz",
					},
				},
			},
		}
		Expect(mapDropFilter(loggingDropTest)).To(Equal(expObsDropTest))
	})

	It("should map logging.PruneFilterSpec to observability.PruneFilterSpec", func() {
		loggingPruneSpec := &logging.PruneFilterSpec{
			In:    []string{"foo", "bar", "baz"},
			NotIn: []string{"test", "something", "other"},
		}

		expObsPruneSpec := &obs.PruneFilterSpec{
			In:    []string{"foo", "bar", "baz"},
			NotIn: []string{"test", "something", "other"},
		}

		Expect(mapPruneFilter(*loggingPruneSpec)).To(Equal(expObsPruneSpec))
	})

	It("should convert all filters appropriately", func() {
		var (
			loggingClfSpec = logging.ClusterLogForwarderSpec{}
		)
		loggingClfSpec.Filters = []logging.FilterSpec{
			{
				Name: "kube-api-audit",
				Type: logging.FilterKubeAPIAudit,
				FilterTypeSpec: logging.FilterTypeSpec{
					KubeAPIAudit: &logging.KubeAPIAudit{
						Rules: []v1.PolicyRule{{
							Level: "RequestResponse",
							Resources: []v1.GroupResources{
								{
									Group:     "",
									Resources: []string{"pods"},
								},
							},
						}},
						OmitStages: []v1.Stage{v1.StageRequestReceived},
					},
				},
			},
			{
				Name: "drop",
				Type: logging.FilterDrop,
				FilterTypeSpec: logging.FilterTypeSpec{
					DropTestsSpec: &[]logging.DropTest{
						{
							DropConditions: []logging.DropCondition{
								{
									Field:   ".foo",
									Matches: "bar",
								},
							},
						},
						{
							DropConditions: []logging.DropCondition{
								{
									Field:      ".baz",
									NotMatches: "test",
								},
							},
						},
					},
				},
			},
			{
				Name: "prune",
				Type: logging.FilterPrune,
				FilterTypeSpec: logging.FilterTypeSpec{
					PruneFilterSpec: &logging.PruneFilterSpec{
						In: []string{"foo", "bar"},
					},
				},
			},
		}

		expObsClfFilterSpec := []obs.FilterSpec{
			{
				Name: "kube-api-audit",
				Type: logging.FilterKubeAPIAudit,
				KubeAPIAudit: &obs.KubeAPIAudit{
					Rules: []v1.PolicyRule{{
						Level: "RequestResponse",
						Resources: []v1.GroupResources{
							{
								Group:     "",
								Resources: []string{"pods"},
							},
						},
					}},
					OmitStages: []v1.Stage{v1.StageRequestReceived},
				},
			},
			{
				Name: "drop",
				Type: logging.FilterDrop,
				DropTestsSpec: []obs.DropTest{
					{
						DropConditions: []obs.DropCondition{
							{
								Field:   ".foo",
								Matches: "bar",
							},
						},
					},
					{
						DropConditions: []obs.DropCondition{
							{
								Field:      ".baz",
								NotMatches: "test",
							},
						},
					},
				},
			},
			{
				Name: "prune",
				Type: logging.FilterPrune,
				PruneFilterSpec: &obs.PruneFilterSpec{
					In: []string{"foo", "bar"},
				},
			},
		}

		Expect(convertFilters(&loggingClfSpec)).To(Equal(expObsClfFilterSpec))
	})
})
