package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

var _ = Describe("OutputSpec", func() {
	It("recognizes valid type names", func() {
		for _, s := range []string{
			OutputTypeElasticsearch,
			OutputTypeFluentdForward,
			OutputTypeSyslog,
			OutputTypeCloudwatch,
			OutputTypeLoki,
			OutputTypeGoogleCloudLogging,
		} {
			Expect(IsOutputTypeName(s)).To(BeTrue(), "expect recognize %s", s)
		}
	})
	It("rejects unknown type", func() {
		Expect(IsOutputTypeName("bad")).To(BeFalse())
	})
})
