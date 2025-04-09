package scc_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	security "github.com/openshift/api/security/v1"
	"github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	. "github.com/openshift/cluster-logging-operator/internal/utils/comparators/scc"
)

var _ = Describe("scc#AreSame", func() {

	It("should succeed when they are the same", func() {
		left := auth.NewSCC()
		right := left.DeepCopy()
		same, reason := AreSame(*left, *right)
		Expect(same).To(BeTrue(), fmt.Sprintf("Exp. comparator to succeed when fields are the same, reason: %s are different", reason))
	})

	DescribeTable("should fail with different", func(modifications ...func(*security.SecurityContextConstraints)) {
		left := auth.NewSCC()
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
		Entry("Priority nil", func(right *security.SecurityContextConstraints) { right.Priority = nil }, func(left *security.SecurityContextConstraints) { left.Priority = utils.GetPtr[int32](12) }),
		Entry("Priority different value", func(right *security.SecurityContextConstraints) { right.Priority = utils.GetPtr[int32](12) }),
		Entry("AllowPrivilegedContainer", func(right *security.SecurityContextConstraints) { right.AllowPrivilegedContainer = true }),
		Entry("RequiredDropCapabilities", func(right *security.SecurityContextConstraints) {
			right.RequiredDropCapabilities = append(right.RequiredDropCapabilities, "foo")
		}),
		Entry("AllowHostDirVolumePlugin", func(right *security.SecurityContextConstraints) { right.AllowHostDirVolumePlugin = false }),
		Entry("Volumes", func(none *security.SecurityContextConstraints) {}, func(left *security.SecurityContextConstraints) { left.Volumes = left.Volumes[1:] }),
		Entry("DefaultAllowPrivilegeEscalation", func(right *security.SecurityContextConstraints) {
			right.DefaultAllowPrivilegeEscalation = utils.GetPtr(true)
		}),
		Entry("AllowPrivilegeEscalation", func(right *security.SecurityContextConstraints) { right.AllowPrivilegeEscalation = utils.GetPtr(true) }),
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
