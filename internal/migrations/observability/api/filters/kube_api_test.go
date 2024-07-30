package filters

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	v1 "k8s.io/apiserver/pkg/apis/audit/v1"
)

var _ = Describe("#MapKubeAPIAudit", func() {
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

		Expect(MapKubeApiAuditFilter(*loggingAudit)).To(Equal(expObsAudit))
	})
})
