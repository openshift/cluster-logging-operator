package outputs

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	"time"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

var _ = Describe("[Functional][Outputs][Unavailable] Output", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToHttpOutput()
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	Context("when the output is unavailable", func() {
		It("should not cause the collector/normalizer to restart", func() {
			skipAddingOutput := func(b *runtime.PodBuilder) error {
				return nil
			}
			Expect(framework.DeployWithVisitor(skipAddingOutput)).To(BeNil())
			//allow process to load config
			time.Sleep(8 * time.Second)
			Expect(oc.Literal().
				From(fmt.Sprintf("oc -n %s get pod %s -o jsonpath={.status.containerStatuses[0].restartCount}", framework.Namespace, framework.Name)).
				Run()).
				To(Equal("0"), "Exp. the pod to boot without restarting")
		})
	})
})
