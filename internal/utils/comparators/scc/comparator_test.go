package scc_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	security "github.com/openshift/api/security/v1"
	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	. "github.com/openshift/cluster-logging-operator/internal/utils/comparators/scc"
)

var _ = Describe("scc#AreSame", func() {

	It("should succeed when they are the same", func() {
		left := collector.NewSCC()
		right := left.DeepCopy()
		same, reason := AreSame(*left, *right)
		Expect(same).To(BeTrue(), fmt.Sprintf("Exp. comparator to succeed when fields are the same, reason: %s are different", reason))
	})

	DescribeTable("should fail with different", func(modifications ...func(*security.SecurityContextConstraints)) {
		left := collector.NewSCC()
		right := left.DeepCopy()
		if len(modifications) > 0 {
			modifications[0](right)
		}
		if len(modifications) > 1 {
			modifications[1](left)
		}
		same, _ := AreSame(*left, *right)
		Expect(same).To(BeFalse(), "Exp. comparator to fail for dissimilar property")
	},
		Entry("Priority nil", func(right *security.SecurityContextConstraints) { right.Priority = nil }, func(left *security.SecurityContextConstraints) { left.Priority = utils.GetInt32(12) }),
		Entry("Priority different value", func(right *security.SecurityContextConstraints) { right.Priority = utils.GetInt32(12) }),
		Entry("AllowPrivilegedContainer", func(right *security.SecurityContextConstraints) { right.AllowPrivilegedContainer = true }),
		Entry("RequiredDropCapabilities", func(right *security.SecurityContextConstraints) {
			right.RequiredDropCapabilities = append(right.RequiredDropCapabilities, "foo")
		}),
		Entry("AllowHostDirVolumePlugin", func(right *security.SecurityContextConstraints) { right.AllowHostDirVolumePlugin = false }),
		Entry("Volumes", func(right *security.SecurityContextConstraints) { right.Volumes = right.Volumes[1:] }),
		Entry("DefaultAllowPrivilegeEscalation", func(right *security.SecurityContextConstraints) {
			right.DefaultAllowPrivilegeEscalation = utils.GetBool(true)
		}),
		Entry("AllowPrivilegeEscalation", func(right *security.SecurityContextConstraints) { right.AllowPrivilegeEscalation = utils.GetBool(true) }),
		Entry("RunAsUser", func(right *security.SecurityContextConstraints) {
			right.RunAsUser = security.RunAsUserStrategyOptions{}
		}),
		Entry("SELinuxContext", func(right *security.SecurityContextConstraints) {
			right.SELinuxContext = security.SELinuxContextStrategyOptions{}
		}),
		Entry("ReadOnlyRootFilesystem", func(right *security.SecurityContextConstraints) { right.ReadOnlyRootFilesystem = false }),
		Entry("ForbiddenSysctls", func(right *security.SecurityContextConstraints) { right.ForbiddenSysctls = []string{"abc"} }),
		Entry("SeccompProfiles", func(right *security.SecurityContextConstraints) { right.SeccompProfiles = []string{"abc"} }),
	)
})
