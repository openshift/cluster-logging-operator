package misc

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
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

	It("should set VECTOR_RAISE_FD_LIMIT=true in the collector container", func() {
		Expect(framework.DeployWithVisitor(func(b *runtime.PodBuilder) error {
			b.GetContainer(constants.CollectorName).AddEnvVar("VECTOR_RAISE_FD_LIMIT", "true")
			return nil
		})).To(BeNil())

		out, err := framework.RunCommand(constants.CollectorName, "sh", "-c", "echo $VECTOR_RAISE_FD_LIMIT")
		Expect(err).To(BeNil())
		Expect(out).To(ContainSubstring("true"))
	})

	It("should start successfully when VECTOR_RAISE_FD_LIMIT is false", func() {
		Expect(framework.DeployWithVisitor(func(b *runtime.PodBuilder) error {
			b.GetContainer(constants.CollectorName).AddEnvVar("VECTOR_RAISE_FD_LIMIT", "false")
			return nil
		})).To(BeNil())

		out, err := framework.RunCommand(constants.CollectorName, "sh", "-c", "echo $VECTOR_RAISE_FD_LIMIT")
		Expect(err).To(BeNil())
		Expect(out).To(ContainSubstring("false"))
	})
})
