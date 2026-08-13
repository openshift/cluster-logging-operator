package misc

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
)

var _ = Describe("[Functional][Misc][RaiseFdLimit] Vector raise-fd-limit", func() {

	var framework *functional.CollectorFunctionalFramework

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeInfrastructure).
			ToHttpOutput()
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	It("should start with VECTOR_RAISE_FD_LIMIT=true", func() {
		Expect(framework.Deploy()).To(BeNil())

		out, err := framework.RunCommand(constants.CollectorName, "sh", "-c", "echo $VECTOR_RAISE_FD_LIMIT")
		Expect(err).To(BeNil())
		Expect(out).To(ContainSubstring("true"))
	})
})
