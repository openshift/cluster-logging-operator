package fluentd

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

func checkMountReadOnly(mount string) {
	touchFile := mount + "/1"
	result, err := oc.Exec().WithNamespace(helpers.OpenshiftLoggingNS).Pod("service/fluentd").
		WithCmd("bash", "-c", "touch "+touchFile).Run()
	Expect(result).To(HavePrefix("touch: cannot touch '" + touchFile + "': Read-only file system"))
	Expect(err).To(MatchError("exit status 1"))
}

var _ = Describe("Tests of fluentd container security stance", func() {
	e2e := helpers.NewE2ETestFramework()

	BeforeEach(func() {
		components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
		if err := e2e.SetupClusterLogging(components...); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
		}
		if err := e2e.WaitFor(helpers.ComponentTypeCollector); err != nil {
			Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", helpers.ComponentTypeCollector, err))
		}
	})

	AfterEach(func() {
		e2e.Cleanup()
		e2e.WaitForCleanupCompletion(helpers.OpenshiftLoggingNS, []string{"fluentd"})
	}, helpers.DefaultCleanUpTimeout)

	It("fluentd containers should have tight security settings", func() {
		By("having all Linux capabilities disabled")
		result, err := oc.Exec().WithNamespace(helpers.OpenshiftLoggingNS).Pod("service/fluentd").
			WithCmd("bash", "-c", "getpcaps 1 2>&1").Run()
		Expect(result).To(Equal("Capabilities for `1': ="))
		Expect(err).NotTo(HaveOccurred())

		By("having all sysctls disabled")
		result, err = oc.Exec().WithNamespace(helpers.OpenshiftLoggingNS).Pod("service/fluentd").
			WithCmd("/usr/sbin/sysctl", "net.ipv4.ip_local_port_range=0").Run()
		Expect(result).To(HavePrefix("sysctl: setting key \"net.ipv4.ip_local_port_range\": Read-only file system"))
		Expect(err).To(MatchError("exit status 255"))

		By("disabling privilege escalation")
		result, err = oc.Exec().WithNamespace(helpers.OpenshiftLoggingNS).Pod("service/fluentd").
			WithCmd("bash", "-c", "cat /proc/1/status | grep NoNewPrivs").Run()
		Expect(result).To(Equal("NoNewPrivs:\t1"))
		Expect(err).NotTo(HaveOccurred())

		By("mounting the root filesystem read-only")
		checkMountReadOnly("/")

		By("mounting the needed subdirectories of /var/log read-only")
		for _, d := range []string{"", "containers", "journal", "openshift-apiserver", "audit", "kube-apiserver", "pods", "oauth-apiserver"} {
			checkMountReadOnly("/var/log/" + d)
		}

		By("not running as a privileged container")
		result, err = oc.Get().WithNamespace(helpers.OpenshiftLoggingNS).Pod().Selector("component=fluentd").
			OutputJsonpath("{.items[0].spec.containers[0].securityContext.privileged}").Run()
		Expect(result).To(BeEmpty())
		Expect(err).NotTo(HaveOccurred())
	})
})
