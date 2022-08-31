package fluentd

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"runtime"
	"strconv"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

const (
	timeout         = `6m`
	pollingInterval = `10s`
)

var collector string

func runInFluentdContainer(command string, args ...string) (string, error) {
	return oc.Exec().WithNamespace(constants.OpenshiftNS).Pod(collector).Container(constants.CollectorName).WithCmd(command, args...).Run()
}

func checkMountReadOnly(mount string) {
	Eventually(func(g Gomega) {
		touchFile := mount + "/1"
		result, err := runInFluentdContainer("bash", "-c", "touch "+touchFile)
		g.Expect(result).To(HavePrefix("touch: cannot touch '" + touchFile + "': Read-only file system"))
		g.Expect(err).To(MatchError("exit status 1"))
	}, timeout, pollingInterval).Should(Succeed())
}

func verifyNamespaceLabels() {
	labels, err := oc.Get().Resource("namespaces", constants.OpenshiftNS).OutputJsonpath("{.metadata.labels}").Run()
	Expect(err).To(BeNil())
	// In ocp 4.12+ only enforce is required
	expMatch := fmt.Sprintf(`{.*"%s":"%s".*}`, constants.PodSecurityLabelEnforce, constants.PodSecurityLabelValue)
	Expect(labels).To(MatchRegexp(expMatch), "Expected label to be found")
}

var _ = Describe("Tests of collector container security stance", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)

	e2e := framework.NewE2ETestFramework()

	BeforeEach(func() {
		forwarder := &logging.ClusterLogForwarder{
			TypeMeta: metav1.TypeMeta{
				Kind:       logging.ClusterLogForwarderKind,
				APIVersion: logging.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.SingletonName,
				Namespace: constants.OpenshiftNS,
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

		nodes := 0
		out, err := oc.Literal().From("oc get nodes --skip-headers -o name").Run()
		if err != nil {
			Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", helpers.ComponentTypeCollector, err))
		}
		nodes = len(strings.Split(out, "\n"))
		log.V(3).Info("Waiting for a stable number collectors to be ready", "nodes", nodes)
		if err := e2e.WaitForResourceCondition(constants.OpenshiftNS, "daemonset", constants.CollectorName, "", "{.status.numberReady}", 10,
			func(out string) (bool, error) {
				ready, err := strconv.Atoi(strings.TrimSpace(out))
				if err != nil {
					return false, err
				}
				log.V(3).Info("Checking ready pods", "pods", ready, "need", nodes)
				if ready == nodes {
					return true, nil
				}
				return false, nil
			}); err != nil {
			Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", helpers.ComponentTypeCollector, err))
		}

		podList, err := oc.Get().WithNamespace(constants.OpenshiftNS).Pod().Selector("component=collector").OutputJsonpath("{.items[*].metadata.name}").Run()
		Expect(err).To(BeNil())
		for _, pod := range strings.Split(podList, " ") {
			if oc.Exec().WithNamespace(constants.OpenshiftNS).Pod(pod).Container(constants.CollectorName).WithCmd("hostname").Output() == nil {
				collector = pod
			}
		}
		Expect(collector).To(Not(BeNil()))
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

		// LOG-2620: containers violate PodSecurity for 4.12+
		By("making sure collector namespace has the pod security label")
		verifyNamespaceLabels()
	})
})
