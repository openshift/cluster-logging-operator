package outputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"time"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	"github.com/openshift/cluster-logging-operator/test/runtime"
)

var _ = Describe("[Functional][Outputs] FluentdForward Output", func() {

	var (
		framework *functional.FluentdFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewFluentdFunctionalFramework()
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToFluentForwardOutput()
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
			//allow fluent process to load config
			time.Sleep(8 * time.Second)
			Expect(oc.Literal().
				From(fmt.Sprintf("oc -n %s get pod %s -o jsonpath={.status.containerStatuses[0].restartCount}", framework.Namespace, framework.Name)).
				Run()).
				To(Equal("0"), "Exp. the pod to boot without restarting")
		})
	})
})
