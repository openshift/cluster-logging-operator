package filter

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

var _ = Describe("filter validation", func() {
	It("rejects filter with no type", func() {
		_, err := RemapVRL(&loggingv1.FilterSpec{Name: "foo"})
		Expect(err).To(MatchError("filter foo: missing filter type"))
	})
	It("rejects filter with bad type", func() {

		_, err := RemapVRL(&loggingv1.FilterSpec{Name: "foo", Type: "notatype"})
		Expect(err).To(MatchError("filter foo: unknown filter type: notatype"))
	})
	It("accepts filter with nil spec (default filter)", func() {
		_, err := RemapVRL(&loggingv1.FilterSpec{Name: "foo", Type: loggingv1.FilterKubeAPIAudit})
		Expect(err).To(Succeed())
	})
})
