package factory

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

var _ = Describe("#NewDaemonSet", func() {

	var (
		daemonSet   *apps.DaemonSet
		expSelector = map[string]string{
			"provider":      "openshift",
			"component":     "thecomponent",
			"logging-infra": "thecomponent",
		}
		expLabels = map[string]string{
			"provider":       "openshift",
			"component":      "thecomponent",
			"logging-infra":  "thecomponent",
			"implementation": "collectorImpl",
		}
	)

	BeforeEach(func() {
		daemonSet = NewDaemonSet("thenname", "thenamespace", "thecomponent", "thecomponent", "collectorImpl", core.PodSpec{})
	})

	It("should leave the MinReadySeconds as the default", func() {
		Expect(daemonSet.Spec.MinReadySeconds).ToNot(Equal(0), "Exp. the MinReadySeconds to be the default")
	})

	It("should only include the provider, component, logging-infra in the selector", func() {
		Expect(daemonSet.Spec.Selector.MatchLabels).To(Equal(expSelector), "Exp. the selector to only include: provider, component, logging-infra")
	})

	It("should include the collector implementation in the labels only", func() {
		Expect(daemonSet.Labels).To(Equal(expLabels))
		Expect(daemonSet.Spec.Template.Labels).To(Equal(expLabels))
	})
	It("should include the critical pod annotation", func() {
		Expect(daemonSet.Spec.Template.ObjectMeta.Annotations).To(HaveKey("scheduler.alpha.kubernetes.io/critical-pod"))
	})

})
