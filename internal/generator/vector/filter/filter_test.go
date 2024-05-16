package filter

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("filter validation", func() {
	It("rejects filter with no type", func() {
		_, err := RemapVRL(&obs.FilterSpec{Name: "foo"})
		Expect(err).To(MatchError("filter foo: missing filter type"))
	})
	It("rejects filter with bad type", func() {

		_, err := RemapVRL(&obs.FilterSpec{Name: "foo", Type: "notatype"})
		Expect(err).To(MatchError("filter foo: unknown filter type: notatype"))
	})
	It("accepts filter with nil spec (default filter)", func() {
		_, err := RemapVRL(&obs.FilterSpec{Name: "foo", Type: obs.FilterTypeKubeAPIAudit})
		Expect(err).To(Succeed())
	})
})
