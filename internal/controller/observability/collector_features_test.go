package observability_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/controller/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
)

var _ = Describe("#EvaluateAnnotationsForEnabledCapabilities", func() {

	It("should do nothing if the annotations are nil", func() {
		options := framework.Options{}
		observability.EvaluateAnnotationsForEnabledCapabilities(nil, options)
		Expect(options).To(BeEmpty(), "Exp no entries added to the options")
	})
	DescribeTable("when forwarder is not nil", func(enabledOption, value string, pairs ...string) {
		if len(pairs)%2 != 0 {
			Fail("Annotations must be passed as pairs to the test table")
		}
		options := framework.Options{}
		annotations := map[string]string{}
		for i := 0; i < len(pairs); i = i + 2 {
			key := pairs[i]
			value := pairs[i+1]
			annotations[key] = value
		}
		observability.EvaluateAnnotationsForEnabledCapabilities(annotations, options)
		if enabledOption == "" {
			Expect(options).To(BeEmpty(), "Exp. the option to be disabled")
		} else {
			Expect(options[enabledOption]).To(Equal(value), "Exp the option to equal the given value")
		}

	},
		Entry("enables debug for true", helpers.EnableDebugOutput, "true", AnnotationDebugOutput, "true"),
		Entry("enables debug for True", helpers.EnableDebugOutput, "true", AnnotationDebugOutput, "True"),
		Entry("disables debug for anything else", "", "", AnnotationDebugOutput, "abcdef"),
	)

})
