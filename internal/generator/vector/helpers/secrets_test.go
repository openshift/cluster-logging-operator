package helpers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("helpers", func() {
	const outputName = "anOutput"
	var (
		legacyToken = &corev1.Secret{}
		aSecret     = &corev1.Secret{}
	)
	var _ = DescribeTable("#GetOutputSecret", func(o obs.OutputSpec, expSecret *corev1.Secret) {
		secrets := map[string]*corev1.Secret{
			constants.LogCollectorToken: legacyToken,
			outputName:                  aSecret,
		}
		Expect(GetOutputSecret(o, secrets)).To(Equal(expSecret))
	},
		Entry("should return the secret when found in all output secrets", obs.OutputSpec{
			Name: outputName,
			Type: obs.OutputTypeSplunk,
		}, aSecret),
		Entry("should return nil when not found in all output secrets", obs.OutputSpec{
			Name: "nameNotFoune",
			Type: obs.OutputTypeSplunk,
		}, nil),
	)

})
