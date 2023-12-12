package factory

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

var _ = Describe("#NewDeployment", func() {

	var (
		deployment  *apps.Deployment
		expSelector = map[string]string{
			"provider":                        "openshift",
			"component":                       "thecomponent",
			"logging-infra":                   "thecomponent",
			constants.CollectorDeploymentKind: constants.DeploymentType,
		}
		expLabels = map[string]string{
			"provider":                           "openshift",
			"component":                          "thecomponent",
			"logging-infra":                      "thecomponent",
			"pod-security.kubernetes.io/enforce": "privileged",
			"security.openshift.io/scc.podSecurityLabelSync": "false",
			"implementation":                  "collectorImpl",
			constants.CollectorDeploymentKind: constants.DeploymentType,
		}
	)

	BeforeEach(func() {
		deployment = NewDeployment("thenamespace", "thename", "thecomponent", "thecomponent", "collectorImpl", core.PodSpec{})
	})

	It("should leave the MinReadySeconds as the default", func() {
		Expect(deployment.Spec.MinReadySeconds).ToNot(Equal(0), "Exp. the MinReadySeconds to be the default")
	})

	It("should only include the provider, component, logging-infra, and collector-deployment-kind in the selector", func() {
		Expect(deployment.Spec.Selector.MatchLabels).To(Equal(expSelector), "Exp. the selector to only include: provider, component, logging-infra, and collector-deployment-kind")
	})

	It("should include the collector implementation in the labels only", func() {
		Expect(deployment.Labels).To(Equal(expLabels))
		Expect(deployment.Spec.Template.Labels).To(Equal(expLabels))
	})
})
