package outputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("validating Splunk", func() {
	Context("#ValidateSplunkMetadata", func() {

		It("should fail validation if meet indexed field not in payload", func() {
			spec := obs.OutputSpec{
				Name: "output",
				Type: obs.OutputTypeSplunk,
				Splunk: &obs.Splunk{
					PayloadKey:    ".openshift",
					IndexedFields: []obs.FieldPath{`.log_type`, `.openshift.sequence`, `.kubernetes.annotations."openshift.io/scc"`},
				},
			}
			res := ValidateSplunk(spec)
			Expect(res).ToNot(BeEmpty())
			Expect(len(res)).To(Equal(2))
			Expect(res[0]).To(Equal(`Indexed field: .log_type not part of payload: .openshift`))
			Expect(res[1]).To(Equal(`Indexed field: .kubernetes.annotations."openshift.io/scc" not part of payload: .openshift`))
		})

		It("should pass validation indexed key from payload", func() {
			spec := obs.OutputSpec{
				Name: "output",
				Type: obs.OutputTypeSplunk,
				Splunk: &obs.Splunk{
					PayloadKey:    ".openshift",
					IndexedFields: []obs.FieldPath{`.openshift.sequence`, `.openshift.cluster_id`},
				},
			}
			res := ValidateSplunk(spec)
			Expect(res).To(BeEmpty())
		})

		It("should pass validation payload not set", func() {
			spec := obs.OutputSpec{
				Name: "output",
				Type: obs.OutputTypeSplunk,
				Splunk: &obs.Splunk{
					IndexedFields: []obs.FieldPath{`.log_type`, `.openshift.sequence`, `.kubernetes.annotations."openshift.io/scc"`},
				},
			}
			res := ValidateSplunk(spec)
			Expect(res).To(BeEmpty())
		})

		It("should pass validation if indexed field not set", func() {
			spec := obs.OutputSpec{
				Name: "output",
				Type: obs.OutputTypeSplunk,
				Splunk: &obs.Splunk{
					PayloadKey: ".openshift",
				},
			}
			res := ValidateSplunk(spec)
			Expect(res).To(BeEmpty())
		})
	})
})
