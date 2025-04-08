package flowcontrol

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	"github.com/openshift/cluster-logging-operator/test/client"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("[E2E] FlowControl", func() {

	var (
		e2e           *framework.E2ETestFramework
		clf           *obs.ClusterLogForwarder
		l             *loki.Receiver
		logStressors  []corev1.Pod
		lokiNS        string
		stressorNS    string
		promQuery     string
		forwarderName = "collector"

		sa  *corev1.ServiceAccount
		err error
	)

	DeployLoggingComponents := func() {
		for idx := range logStressors {
			Expect(client.Get().Create(&logStressors[idx])).To(Succeed())
		}

		Expect(l.Create(client.Get())).To(Succeed())

		if err := e2e.CreateObservabilityClusterLogForwarder(clf); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of clusterlogforwarder: %v", err))
		}

		if err := e2e.WaitForDaemonSet(clf.Namespace, clf.Name); err != nil {
			Fail(fmt.Sprintf("Failed waiting for collector %s/%s to be ready: %v", clf.Namespace, clf.Name, err))
		}
	}

	BeforeEach(func() {
		logStressors = make([]corev1.Pod, 0)
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

		for i := 0; i < 10; i += 1 {
			logStressors = append(logStressors,
				*runtime.NewLogGenerator(stressorNS, fmt.Sprintf("stressor-%d", i), -1, 10*time.Millisecond, "Test Message"),
			) // 100 lines per second
		}
	})

	AfterEach(func() {
		e2e.Cleanup()
	}, framework.DefaultCleanUpTimeout)

	It("should apply policies in both input and output", func() {
		clf.Spec.Inputs = []obs.InputSpec{
			{
				Name: "custom-app-2",
				Type: obs.InputTypeApplication,
				Application: &obs.Application{
					Tuning: &obs.ContainerInputTuningSpec{
						RateLimitPerContainer: &obs.LimitSpec{
							MaxRecordsPerSecond: 200,
						},
					},
				},
			},
		}
		clf.Spec.Outputs = []obs.OutputSpec{
			{
				Name: "loki-2",
				Type: obs.OutputTypeLoki,
				Loki: &obs.Loki{
					URLSpec: obs.URLSpec{
						URL: l.InternalURL("").String(),
					},
				},
				Limit: &obs.LimitSpec{
					MaxRecordsPerSecond: 100,
				},
			},
		}
		clf.Spec.Pipelines = []obs.PipelineSpec{
			{
				InputRefs:  []string{"custom-app-2"},
				OutputRefs: []string{"loki-2"},
				Name:       "e2e-policy-input-output",
			},
		}

		DeployLoggingComponents()

		if !WaitForMetricsToShow() {
			Fail("Metrics not showing up in Prometheus")
		}
		// sleeping for 1 minute to ensure rate of metrics is stable
		time.Sleep(30 * time.Second)

		promQuery = fmt.Sprintf(VectorCompSentEvents, "loki-2")
		ExpectMetricsWithinRange(GetCollectorMetrics(promQuery), 0, 102)

		promQuery = fmt.Sprintf(VectorCompSentEvents, "e2e-policy-input-output")
		ExpectMetricsWithinRange(GetCollectorMetrics(promQuery), 0, 202)
	})
})
