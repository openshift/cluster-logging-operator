package metrics

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	internalruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	"github.com/openshift/cluster-logging-operator/test/client"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	"github.com/openshift/cluster-logging-operator/test/helpers/prometheus"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	corev1 "k8s.io/api/core/v1"
)

const (
	allowlistedMetric    = `vector_open_files`
	nonAllowlistedMetric = `vector_started_total`
)

var _ = Describe("[e2e][collection][metrics] Metrics Collection Profiles", Ordered, func() {
	var (
		e2e           *framework.E2ETestFramework
		clf           *obs.ClusterLogForwarder
		l             *loki.Receiver
		forwarderName = "collector"
		lokiNS        string
		stressorNS    string

		sa  *corev1.ServiceAccount
		err error
	)

	BeforeAll(func() {
		e2e = framework.NewE2ETestFramework()
		lokiNS = e2e.CreateTestNamespace()
		stressorNS = e2e.CreateTestNamespace()

		sa, err = e2e.BuildAuthorizationFor(constants.OpenshiftNS, forwarderName).
			AllowClusterRole(framework.ClusterRoleCollectApplicationLogs).
			Create()
		Expect(err).To(BeNil())

		l = loki.NewReceiver(lokiNS, "loki-server")
		clf = obsruntime.NewClusterLogForwarder(constants.OpenshiftNS, forwarderName, internalruntime.Initialize, func(clf *obs.ClusterLogForwarder) {
			clf.Spec.ServiceAccount.Name = sa.Name
		})

		clf.Spec.Outputs = []obs.OutputSpec{
			{
				Name: "loki-out",
				Type: obs.OutputTypeLoki,
				Loki: &obs.Loki{
					URLSpec: obs.URLSpec{
						URL: l.InternalURL("").String(),
					},
				},
			},
		}
		clf.Spec.Pipelines = []obs.PipelineSpec{
			{
				InputRefs:  []string{string(obs.InputTypeApplication)},
				OutputRefs: []string{"loki-out"},
				Name:       "app-to-loki",
			},
		}

		stressor := runtime.NewLogGenerator(stressorNS, "stressor", 0, 100*time.Millisecond, "profile test message")
		Expect(client.Get().Create(stressor)).To(Succeed())
		Expect(l.Create(client.Get())).To(Succeed())

		if err := e2e.CreateObservabilityClusterLogForwarder(clf); err != nil {
			Fail(fmt.Sprintf("Unable to create CLF: %v", err))
		}
		if err := e2e.WaitForDaemonSet(clf.Namespace, clf.Name); err != nil {
			Fail(fmt.Sprintf("Failed waiting for collector: %v", err))
		}
	})

	AfterAll(func() {
		e2e.Cleanup()
	})

	Context("ServiceMonitor labels", func() {
		DescribeTable("should have the correct collection-profile label",
			func(name, expectedProfile string) {
				output, err := oc.Get().
					WithNamespace(constants.OpenshiftNS).
					Resource("servicemonitor", name).
					OutputJsonpath(`{.metadata.labels['monitoring\.openshift\.io/collection-profile']}`).
					Run()
				Expect(err).NotTo(HaveOccurred(), "Failed to get %s-profile ServiceMonitor", expectedProfile)
				Expect(output).To(Equal(expectedProfile))
			},
			Entry("full profile", forwarderName, constants.MetricsCollectionProfileFull),
			Entry("minimal profile", constants.MetricsCollectionProfileMinimal+"-"+forwarderName, constants.MetricsCollectionProfileMinimal),
		)
	})

	Context("Full profile scraping", func() {
		It("should have all collector metrics available in Prometheus under the full profile", func() {
			By("waiting for collector metrics to appear in Thanos")
			Eventually(func(g Gomega) {
				response, err := prometheus.Query(fmt.Sprintf(`%s{namespace="%s"}`, allowlistedMetric, constants.OpenshiftNS))
				g.Expect(err).NotTo(HaveOccurred(), "Failed to query allowlisted metric")
				g.Expect(prometheus.HasResults(response)).To(BeTrue())
			}, 5*time.Minute, 30*time.Second).Should(Succeed(), "Allowlisted metric should be present under full profile")

			By("verifying a non-allowlisted metric is also present under full profile")
			Eventually(func(g Gomega) {
				response, err := prometheus.Query(fmt.Sprintf(`%s{namespace="%s"}`, nonAllowlistedMetric, constants.OpenshiftNS))
				g.Expect(err).NotTo(HaveOccurred(), "Failed to query non-allowlisted metric")
				g.Expect(prometheus.HasResults(response)).To(BeTrue())
			}, 5*time.Minute, 15*time.Second).Should(Succeed(), "Non-allowlisted metric should also be present under full profile")
		})
	})
})
