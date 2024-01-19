package helpers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/logstore/lokistack"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("helpers", func() {
	const outputName = "anOutput"
	var (
		legacyToken = &corev1.Secret{}
		aSecret     = &corev1.Secret{}
	)
	var _ = DescribeTable("#GetOutputSecret", func(o logging.OutputSpec, expSecret *corev1.Secret) {
		secrets := map[string]*corev1.Secret{
			constants.LogCollectorToken: legacyToken,
			outputName:                  aSecret,
		}
		Expect(GetOutputSecret(o, secrets)).To(Equal(expSecret))
	},
		Entry("should return the secret when found in all output secrets", logging.OutputSpec{
			Name: outputName,
			Type: logging.OutputTypeSplunk,
		}, aSecret),
		Entry("should return nil when not found in all output secrets", logging.OutputSpec{
			Name: "nameNotFoune",
			Type: logging.OutputTypeSplunk,
		}, nil),
		Entry("should return the legacy token secret when output is default lokiStack", logging.OutputSpec{
			Name: lokistack.FormatOutputNameFromInput(logging.InputNameApplication),
			Type: logging.OutputTypeLoki,
		}, legacyToken),
	)

})
