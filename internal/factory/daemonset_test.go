package factory

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

var _ = Describe("#NewDaemonSet", func() {

	var (
		daemonSet    *apps.DaemonSet
		expSelectors = runtime.Selectors("theinstancename", "thecomponent", "thecomponent")
	)

	BeforeEach(func() {
		daemonSet = NewDaemonSet("thenamespace", "thenname", "theinstancename", "thecomponent", "thecomponent", core.PodSpec{})
	})

	It("should leave the MinReadySeconds as the default", func() {
		Expect(daemonSet.Spec.MinReadySeconds).ToNot(Equal(0), "Exp. the MinReadySeconds to be the default")
	})

	It("should only include the kubernetes common labels in the selector", func() {
		Expect(daemonSet.Spec.Selector.MatchLabels).To(Equal(expSelectors), "Exp. the selector to contain kubernetes common labels")
	})

})
