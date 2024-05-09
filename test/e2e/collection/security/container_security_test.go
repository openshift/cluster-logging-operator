package security

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"runtime"
	"strconv"
	"strings"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

var _ = Describe("Tests of collector container security stance", func() {

	var (
		collector               string
		runInCollectorContainer = func(command string, args ...string) (string, error) {
			return oc.Exec().WithNamespace(constants.OpenshiftNS).Pod(collector).Container(constants.CollectorName).WithCmd(command, args...).RunFor(time.Second * 15)
		}
		checkMountReadOnly = func(mount string) {
			touchFile := mount + "/1"
			result, err := runInCollectorContainer("touch", touchFile)
			Expect(result).To(MatchRegexp("touch:.cannot.*touch.*" + touchFile + ".*Read-only file system"))
			Expect(err).To(MatchError("exit status 1"))
		}

		verifyNamespaceLabels = func() {
			labels, err := oc.Get().Resource("namespaces", constants.OpenshiftNS).OutputJsonpath("{.metadata.labels}").Run()
			Expect(err).To(BeNil())
			// In ocp 4.12+ only enforce is required
			expMatch := fmt.Sprintf(`{.*"%s":"%s".*}`, constants.PodSecurityLabelEnforce, constants.PodSecurityLabelValue)
			Expect(labels).To(MatchRegexp(expMatch), "Expected label to be found")
		}
		e2e *framework.E2ETestFramework
	)

	AfterEach(func() {
		e2e.Cleanup()
		e2e.WaitForCleanupCompletion(constants.OpenshiftNS, []string{constants.CollectorName})
	}, framework.DefaultCleanUpTimeout)

	var _ = DescribeTable("collector containers should have tight security settings", func(collectorType logging.LogCollectionType) {
		_, filename, _, _ := runtime.Caller(0)
		log.Info("Running ", "filename", filename)
		e2e = framework.NewE2ETestFramework()

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
				Outputs: []logging.OutputSpec{
					{
						Name: "es",
						Type: logging.OutputTypeElasticsearch,
						URL:  "http://foo.bar.svc:24224",
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						Name:       "infra-pipe",
						OutputRefs: []string{"es"},
						InputRefs:  []string{logging.InputNameInfrastructure, logging.InputNameApplication, logging.InputNameAudit},
					},
				},
			},
		}
		if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of clusterlogforwarder: %v", err))
		}
		if err := e2e.DeployComponents(helpers.ComponentTypeCollectorVector); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
		}
		nodes := 0
		out, err := oc.Literal().From("oc get nodes -o name").Run()
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

		pod, err := oc.Get().WithNamespace(constants.OpenshiftNS).Pod().Selector("component=collector").OutputJsonpath("{.items[0].metadata.name}").Run()
		Expect(err).To(BeNil())
		collector = pod
		Expect(collector).To(Not(BeNil()))

		By("having all Linux capabilities disabled")
		result, err := runInCollectorContainer("getpcaps", "1")
		Expect(result).To(MatchRegexp(`^(Capabilities\sfor\s.)?1'?:\s=$`))
		Expect(err).NotTo(HaveOccurred())

		By("having all sysctls disabled")
		result, _ = runInCollectorContainer("/usr/sbin/sysctl", "net.ipv4.ip_local_port_range=0")
		Expect(result).To(ContainSubstring("sysctl: no such file"))

		By("disabling privilege escalation")
		result, err = runInCollectorContainer("cat", "/proc/1/status")
		Expect(result).To(MatchRegexp("NoNewPrivs:\t1"))
		Expect(err).NotTo(HaveOccurred())

		By("mounting the root filesystem read-only")
		checkMountReadOnly("/")

		By("mounting the needed subdirectories of /var/log read-only")
		for _, d := range []string{"", "journal", "openshift-apiserver", "audit", "kube-apiserver", "pods", "oauth-apiserver"} {
			checkMountReadOnly("/var/log/" + d)
		}

		By("not running as a privileged container")
		result, err = oc.Get().WithNamespace(constants.OpenshiftNS).Pod().Selector("component=collector").
			OutputJsonpath("{.items[0].spec.containers[0].securityContext.privileged}").Run()
		Expect(result).To(BeEmpty())
		Expect(err).NotTo(HaveOccurred())

		//TODO: fix me for vector buffered
		//if collectorType == logging.LogCollectionTypeFluentd {
		//	By("making sure on disk footprint is readable only to the collector")
		//	Eventually(func() (string, error) {
		//		return runInCollectorContainer("bash", "-ceuo", "pipefail", `stat --format=%a /var/lib/fluentd/es/* | sort -u`)
		//	}).
		//		WithTimeout(2*time.Minute).
		//		Should(Equal("600"), "Exp the data directory to have alternate permissions")
		//}

		// LOG-2620: containers violate PodSecurity for 4.12+
		By("making sure collector namespace has the pod security label")
		verifyNamespaceLabels()

	},
		Entry("for vector impl", logging.LogCollectionTypeVector),
	)
})
