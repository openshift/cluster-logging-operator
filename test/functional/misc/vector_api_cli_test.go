package misc

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
)

var _ = Describe("[Functional][Misc][API_CLI] Functional test", func() {

	var framework *functional.CollectorFunctionalFramework

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).FromInput(obs.InputNameInfrastructure).ToHttpOutput()
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("invoking vector CLI commands that talk to the vector API", func() {
		It("should work", func() {
			Expect(framework.Deploy()).To(BeNil())
			out, _ := framework.RunCommand(constants.CollectorName, `curl`, `-sv`, `-m`, `5`, `--connect-timeout`, `3`, `http://127.0.0.1:8686/health`)
			Expect(out).To(ContainSubstring(`{"ok":true}`))
		})
	})
})
