package factory

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

var _ = Describe("#NewDaemonSet", func() {

	var (
		daemonSet      *apps.DaemonSet
		expSelectors   = runtime.Selectors("theinstancename", "thecomponent", "thecomponent")
		op             = framework.Options{}
		maxUnavailable string
	)

	Context("with common properties and maxUnavailable empty", func() {
		BeforeEach(func() {
			daemonSet = NewDaemonSet(
				"thenamespace",
				"thenname",
				"theinstancename",
				"thecomponent",
				"thecomponent",
				maxUnavailable,
				core.PodSpec{},
			)
		})

		It("should leave the MinReadySeconds as the default", func() {
			Expect(daemonSet.Spec.MinReadySeconds).ToNot(Equal(0), "Exp. the MinReadySeconds to be the default")
		})

		It("should only include the kubernetes common labels in the selector", func() {
			Expect(daemonSet.Spec.Selector.MatchLabels).To(Equal(expSelectors), "Exp. the selector to contain kubernetes common labels")
		})
	})

	// This option should never be set as invalid since there is validation on the setter. This unit test
	// ensures the ds method handles it properly regardless
	DescribeTable("with maxUnavailable option", func(op framework.Options, exp string) {
		maxUnavailable = GetMaxUnavailableValue(op)
		daemonSet = NewDaemonSet(
			"thenamespace",
			"thenname",
			"theinstancename",
			"thecomponent",
			"thecomponent",
			maxUnavailable,
			core.PodSpec{},
		)
		Expect(daemonSet.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.String()).To(Equal(exp), "Exp. the maxUnavailable value to match")
	},
		Entry("missing", op, "100%"),
		Entry("set to empty string", framework.Options{framework.MaxUnavailableOption: ""}, "100%"),
		Entry("set to invalid value 'blue'", framework.Options{framework.MaxUnavailableOption: "blue"}, "100%"),
		Entry("set to invalid zero", framework.Options{framework.MaxUnavailableOption: "0"}, "100%"),
		Entry("set to invalid decimal", framework.Options{framework.MaxUnavailableOption: "2.5"}, "100%"),
		Entry("set to invalid percentage", framework.Options{framework.MaxUnavailableOption: "200%"}, "100%"),
		Entry("set to whole number", framework.Options{framework.MaxUnavailableOption: "50"}, "50"),
		Entry("set to percentage", framework.Options{framework.MaxUnavailableOption: "99%"}, "99%"),
		Entry("set to '100%'", framework.Options{framework.MaxUnavailableOption: "100%"}, "100%"),
	)
})
