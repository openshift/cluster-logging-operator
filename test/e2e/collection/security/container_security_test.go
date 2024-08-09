package security

import (
	"context"
	"fmt"
	internalruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"runtime"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

var _ = Describe("Tests of collector container security stance", func() {

	const (
		namespace = constants.OpenshiftNS
		name      = constants.CollectorName
	)

	var (
		collector               string
		runInCollectorContainer = func(command string, args ...string) (string, error) {
			return oc.Exec().WithNamespace(namespace).Pod(collector).Container(name).WithCmd(command, args...).RunFor(time.Second * 15)
		}
		checkMountReadOnly = func(mount string) {
			touchFile := mount + "/1"
			result, err := runInCollectorContainer("touch", touchFile)
			Expect(result).To(MatchRegexp("touch:.cannot.*touch.*" + touchFile + ".*Read-only file system"))
			Expect(err).To(MatchError("exit status 1"))
		}

		verifyNamespaceLabels = func() {
			labels, err := oc.Get().Resource("namespaces", namespace).OutputJsonpath("{.metadata.labels}").Run()
			Expect(err).To(BeNil())
			// In ocp 4.12+ only enforce is required
			expMatch := fmt.Sprintf(`{.*"%s":"%s".*}`, constants.PodSecurityLabelEnforce, constants.PodSecurityLabelValue)
			Expect(labels).To(MatchRegexp(expMatch), "Expected label to be found")
		}
		e2e            *framework.E2ETestFramework
		serviceAccount *corev1.ServiceAccount
		err            error
	)

	AfterEach(func() {
		e2e.Cleanup()
		e2e.WaitForCleanupCompletion(namespace, []string{name})
	}, framework.DefaultCleanUpTimeout)

	It("should have tight security settings", func() {
		_, filename, _, _ := runtime.Caller(0)
		log.Info("Running ", "filename", filename)
		e2e = framework.NewE2ETestFramework()

		if serviceAccount, err = e2e.BuildAuthorizationFor(namespace, name).
			AllowClusterRole(framework.ClusterRoleCollectInfrastructureLogs).
			AllowClusterRole(framework.ClusterRoleCollectApplicationLogs).
			AllowClusterRole(framework.ClusterRoleCollectAuditLogs).Create(); err != nil {
			Fail(err.Error())
		}

		forwarder := obsruntime.NewClusterLogForwarder(namespace, name, internalruntime.Initialize,
			func(clf *obs.ClusterLogForwarder) {
				clf.Spec = obs.ClusterLogForwarderSpec{
					Outputs: []obs.OutputSpec{
						{
							Name: "es",
							Type: obs.OutputTypeElasticsearch,
							Elasticsearch: &obs.Elasticsearch{
								URLSpec: obs.URLSpec{
									URL: "http://foo.bar.svc:24224",
								},
							},
						},
					},
					Pipelines: []obs.PipelineSpec{
						{
							Name:       "infra-pipe",
							OutputRefs: []string{"es"},
							InputRefs:  []string{string(obs.InputTypeInfrastructure), string(obs.InputTypeApplication), string(obs.InputTypeAudit)},
						},
					},
					ServiceAccount: obs.ServiceAccount{
						Name: serviceAccount.Name,
					},
				}
			})

		if err := e2e.CreateObservabilityClusterLogForwarder(forwarder); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of clusterlogforwarder: %v", err))
		}

		log.V(3).Info("Waiting for a stable number collectors to be ready")
		if err := e2e.WaitForDaemonSet(forwarder.Namespace, forwarder.Name); err != nil {
			Fail(fmt.Sprintf("Failed waiting for collector %s/%s to be ready: %v", forwarder.Namespace, forwarder.Name, err))
		}

		labelSelector := fmt.Sprintf("%s=%s", constants.LabelK8sComponent, constants.CollectorName)
		pods, err := e2e.KubeClient.CoreV1().Pods(forwarder.Namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})

		Expect(err).ToNot(HaveOccurred(), "Exp. to retrieve pods associated with the deployment")
		Expect(pods.Items).ToNot(BeEmpty(), fmt.Sprintf("Exp. to find deployed collector pods in %q with label %q", forwarder.Namespace, labelSelector))
		collector = pods.Items[0].Name

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
		result, err = oc.Get().WithNamespace(forwarder.Namespace).Pod().Selector(labelSelector).
			OutputJsonpath("{.items[0].spec.containers[0].securityContext.privileged}").Run()
		Expect(result).To(BeEmpty())
		Expect(err).NotTo(HaveOccurred())

		//TODO: fix me for vector buffered
		//	By("making sure on disk footprint is readable only to the collector")
		//	Eventually(func() (string, error) {
		//		return runInCollectorContainer("bash", "-ceuo", "pipefail", `stat --format=%a /var/lib/fluentd/es/* | sort -u`)
		//	}).
		//		WithTimeout(2*time.Minute).
		//		Should(Equal("600"), "Exp the data directory to have alternate permissions")

		// LOG-2620: containers violate PodSecurity for 4.12+
		By("making sure collector namespace has the pod security label")
		verifyNamespaceLabels()

	})
})
