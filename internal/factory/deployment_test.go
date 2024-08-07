package factory

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

var _ = Describe("#NewDeployment", func() {

	var (
		deployment   *apps.Deployment
		expSelectors = runtime.Selectors("thename", "thecomponent", "thecomponent")
	)

	BeforeEach(func() {
		deployment = NewDeployment("thenamespace", "thename", "thecomponent", "thecomponent", 1, core.PodSpec{})
	})

	It("should leave the MinReadySeconds as the default", func() {
		Expect(deployment.Spec.MinReadySeconds).ToNot(Equal(0), "Exp. the MinReadySeconds to be the default")
	})

	It("should only include kubernetes common labels in the selector", func() {
		Expect(deployment.Spec.Selector.MatchLabels).To(Equal(expSelectors), "Exp. the selector to only include kubernetes common labels")
	})
})
