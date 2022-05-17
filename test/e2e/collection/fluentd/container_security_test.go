package fluentd

import (
	"fmt"
	"runtime"

	"github.com/ViaQ/logerr/v2/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	timeout         = `6m`
	pollingInterval = `10s`
)

func runInFluentdContainer(command string, args ...string) (string, error) {
	return oc.Exec().WithNamespace(constants.OpenshiftNS).WithPodGetter(oc.Get().WithNamespace(constants.OpenshiftNS).Pod().Selector("component=collector").OutputJsonpath("{.items[0].metadata.name}")).Container(constants.CollectorName).WithCmd(command, args...).Run()
}

func checkMountReadOnly(mount string) {
	Eventually(func(g Gomega) {
		touchFile := mount + "/1"
		result, err := runInFluentdContainer("bash", "-c", "touch "+touchFile)
		g.Expect(result).To(HavePrefix("touch: cannot touch '" + touchFile + "': Read-only file system"))
		g.Expect(err).To(MatchError("exit status 1"))
	}, timeout, pollingInterval).Should(Succeed())
}

var _ = Describe("Tests of collector container security stance", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.NewLogger("e2e-fluentd").Info("Running ", "filename", filename)
	e2e := framework.NewE2ETestFramework()

	BeforeEach(func() {
		forwarder := &logging.ClusterLogForwarder{
			TypeMeta: metav1.TypeMeta{
				Kind:       logging.ClusterLogForwarderKind,
				APIVersion: logging.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "instance",
			},
			Spec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name:           "infra-input",
						Infrastructure: &logging.Infrastructure{},
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Name: "_infrastructure",
						Type: logging.OutputTypeFluentdForward,
						URL:  "tcp://foo.bar.svc:24224",
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						Name:       "infra-pipe",
						OutputRefs: []string{"_infrastructure"},
						InputRefs:  []string{"infra-input"},
					},
				},
			},
		}
		if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of clusterlogforwarder: %v", err))
		}
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
		e2e.WaitForCleanupCompletion(constants.OpenshiftNS, []string{constants.CollectorName})
	}, framework.DefaultCleanUpTimeout)

	It("collector containers should have tight security settings", func() {
		By("having all Linux capabilities disabled")
		Eventually(func(g Gomega) {
			result, err := runInFluentdContainer("bash", "-c", "getpcaps 1 2>&1")
			g.Expect(result).To(MatchRegexp(`^(Capabilities\sfor\s.)?1'?:\s=$`))
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
			result, err := oc.Get().WithNamespace(constants.OpenshiftNS).Pod().Selector("component=collector").
				OutputJsonpath("{.items[0].spec.containers[0].securityContext.privileged}").Run()
			g.Expect(result).To(BeEmpty())
			g.Expect(err).NotTo(HaveOccurred())
		}, timeout, pollingInterval).Should(Succeed())

		By("making sure on disk footprint is readable only to the collector")
		Eventually(func(g Gomega) {
			result, err := runInFluentdContainer("bash", "-c", "stat --format=%a /var/lib/fluentd/_infrastructure/* | sort -u")
			g.Expect(result).To(Equal("600"))
			g.Expect(err).NotTo(HaveOccurred())
		}, timeout, pollingInterval).Should(Succeed())
	})
})
