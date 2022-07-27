package flowcontrol

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/client"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[E2E] FlowControl", func() {

	var (
		e2e          *framework.E2ETestFramework
		cl           *loggingv1.ClusterLogging
		clf          *loggingv1.ClusterLogForwarder
		l            *loki.Receiver
		logStressors []corev1.Pod
		lokiNS       string
		stressorNS   string
		promQuery    string
	)

	components := []helpers.LogComponentType{helpers.ComponentTypeCollectorVector, helpers.ComponentTypeStore}

	DeployLoggingComponents := func() {
		for idx := range logStressors {
			Expect(client.Get().Create(&logStressors[idx])).To(Succeed())
		}

		Expect(l.Create(client.Get())).To(Succeed())

		if err := e2e.CreateClusterLogging(cl); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
		}

		if err := e2e.CreateClusterLogForwarder(clf); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of clusterlogforwarder: %v", err))
		}

		for _, component := range components {
			if err := e2e.WaitFor(component); err != nil {
				Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
			}
		}
	}

	BeforeEach(func() {
		logStressors = make([]corev1.Pod, 0)
		e2e = framework.NewE2ETestFramework()

		cl = helpers.NewClusterLogging(components...)

		lokiNS = e2e.CreateTestNamespace()
		stressorNS = e2e.CreateTestNamespace()

		e2e.AddCleanup(func() error {
			opts := metav1.DeleteOptions{}
			return e2e.KubeClient.CoreV1().Namespaces().Delete(context.TODO(), lokiNS, opts)
		})

		e2e.AddCleanup(func() error {
			opts := metav1.DeleteOptions{}
			return e2e.KubeClient.CoreV1().Namespaces().Delete(context.TODO(), stressorNS, opts)
		})

		l = loki.NewReceiver(lokiNS, "loki-server")
		clf = runtime.NewClusterLogForwarder()

		for i := 0; i < 10; i += 1 {
			logStressors = append(logStressors,
				*runtime.NewLogGenerator(stressorNS, fmt.Sprintf("stressor-%d", i), -1, 10*time.Millisecond, "Test Message"),
			) // 100 lines per second
		}
	})

	AfterEach(func() {
		e2e.Cleanup()
	}, framework.DefaultCleanUpTimeout)

	It("applying policies in input/output of pipeline", func() {
		clf.Spec.Inputs = []loggingv1.InputSpec{
			{
				Name: "custom-app-0",
				Application: &loggingv1.Application{
					Namespaces: []string{stressorNS},
					GroupLimit: &loggingv1.LimitSpec{
						MaxRecordsPerSecond: 100,
					}, // 10 files and 100 group limit, so 10 lines per file,
				},
			},
			{
				Name: "custom-app-1",
				Application: &loggingv1.Application{
					Namespaces: []string{stressorNS},
					ContainerLimit: &loggingv1.LimitSpec{
						MaxRecordsPerSecond: 100,
					}, // 10 files and 100 group limit, so 10 lines per file,
				},
			},
		}
		clf.Spec.Outputs = []loggingv1.OutputSpec{
			{
				Name: "loki-0",
				Type: loggingv1.OutputTypeLoki,
				URL:  l.InternalURL("").String(),
				OutputTypeSpec: loggingv1.OutputTypeSpec{
					Loki: &loggingv1.Loki{},
				},
			},
			{
				Name: "loki-1",
				Type: loggingv1.OutputTypeLoki,
				URL:  l.InternalURL("").String(),
				OutputTypeSpec: loggingv1.OutputTypeSpec{
					Loki: &loggingv1.Loki{},
				},
				Limit: &loggingv1.LimitSpec{
					MaxRecordsPerSecond: 100,
				},
			},
		}
		clf.Spec.Pipelines = []loggingv1.PipelineSpec{
			{
				InputRefs:  []string{"custom-app-0"},
				OutputRefs: []string{"loki-0"},
				Name:       "group-policy-at-application",
			},
			{
				InputRefs:  []string{"custom-app-1"},
				OutputRefs: []string{"loki-0"},
				Name:       "container-policy-at-application",
			},
			{
				InputRefs:  []string{loggingv1.InputNameApplication},
				OutputRefs: []string{"loki-1"},
				Name:       "policy-at-loki",
			},
		}

		DeployLoggingComponents()

		if WaitForMetricsToShow() == false {
			Fail("Metrics not showing up in Prometheus")
		}
		// sleeping for 1 minute to ensure rate of metrics is stable
		time.Sleep(30 * time.Second)

		promQuery = fmt.Sprintf(VectorCompSentEvents, "group-policy-at-application")
		ExpectMetricsWithinRange(GetCollectorMetrics(promQuery), 0, 102)

		promQuery = fmt.Sprintf(SumMetric, VectorCompSentEvents)
		promQuery = fmt.Sprintf(promQuery, "container-policy-at-application") // Max number of logs allowed per second is 10 * 100 lines/sec
		ExpectMetricsWithinRange(GetCollectorMetrics(promQuery), 100, 1002)

		promQuery = fmt.Sprintf(VectorCompSentEvents, "loki-1")
		ExpectMetricsWithinRange(GetCollectorMetrics(promQuery), 0, 102)
	})

	It("applying policies in both input and output", func() {
		clf.Spec.Inputs = []loggingv1.InputSpec{
			{
				Name: "custom-app-2",
				Application: &loggingv1.Application{
					GroupLimit: &loggingv1.LimitSpec{
						MaxRecordsPerSecond: 200,
					},
				},
			},
		}
		clf.Spec.Outputs = []loggingv1.OutputSpec{
			{
				Name: "loki-2",
				Type: loggingv1.OutputTypeLoki,
				URL:  l.InternalURL("").String(),
				OutputTypeSpec: loggingv1.OutputTypeSpec{
					Loki: &loggingv1.Loki{},
				},
				Limit: &loggingv1.LimitSpec{
					MaxRecordsPerSecond: 100,
				},
			},
		}
		clf.Spec.Pipelines = []loggingv1.PipelineSpec{
			{
				InputRefs:  []string{"custom-app-2"},
				OutputRefs: []string{"loki-2"},
				Name:       "e2e-policy-input-output",
			},
		}

		DeployLoggingComponents()

		if WaitForMetricsToShow() == false {
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
