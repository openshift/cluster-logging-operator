package fluentd

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

const (
	timeout         = `1m`
	pollingInterval = `10s`
)

func runInFluentdContainer(command string, args ...string) (string, error) {
	return oc.Exec().WithNamespace(helpers.OpenshiftLoggingNS).WithPodGetter(oc.Get().WithNamespace(helpers.OpenshiftLoggingNS).Pod().Selector("component=fluentd").OutputJsonpath("{.items[0].metadata.name}")).Container("fluentd").WithCmd(command, args...).Run()
}

func checkMountReadOnly(mount string) {
	Eventually(func(g Gomega) {
		touchFile := mount + "/1"
		result, err := runInFluentdContainer("bash", "-c", "touch "+touchFile)
		g.Expect(result).To(HavePrefix("touch: cannot touch '" + touchFile + "': Read-only file system"))
		g.Expect(err).To(MatchError("exit status 1"))
	}, timeout, pollingInterval).Should(Succeed())
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
		Eventually(func(g Gomega) {
			result, err := runInFluentdContainer("bash", "-c", "getpcaps 1 2>&1")
			g.Expect(result).To(Equal("Capabilities for `1': ="))
			g.Expect(err).NotTo(HaveOccurred())
		}, timeout, pollingInterval).Should(Succeed())

		By("having all sysctls disabled")
		Eventually(func(g Gomega) {
			result, err := runInFluentdContainer("/usr/sbin/sysctl", "net.ipv4.ip_local_port_range=0")
			g.Expect(result).To(HavePrefix("sysctl: setting key \"net.ipv4.ip_local_port_range\": Read-only file system"))
			g.Expect(err).To(MatchError("exit status 255"))
		}, timeout, pollingInterval).Should(Succeed())

		By("disabling privilege escalation")
		Eventually(func(g Gomega) {
			result, err := runInFluentdContainer("bash", "-c", "cat /proc/1/status | grep NoNewPrivs")
			g.Expect(result).To(Equal("NoNewPrivs:\t1"))
			g.Expect(err).NotTo(HaveOccurred())
		}, timeout, pollingInterval).Should(Succeed())

		By("mounting the root filesystem read-only")
		checkMountReadOnly("/")

		By("mounting the needed subdirectories of /var/log read-only")
		for _, d := range []string{"", "containers", "journal", "openshift-apiserver", "audit", "kube-apiserver", "pods", "oauth-apiserver"} {
			checkMountReadOnly("/var/log/" + d)
		}

		By("not running as a privileged container")
		Eventually(func(g Gomega) {
			result, err := oc.Get().WithNamespace(helpers.OpenshiftLoggingNS).Pod().Selector("component=fluentd").
				OutputJsonpath("{.items[0].spec.containers[0].securityContext.privileged}").Run()
			g.Expect(result).To(BeEmpty())
			g.Expect(err).NotTo(HaveOccurred())
		}, timeout, pollingInterval).Should(Succeed())
	})
})
