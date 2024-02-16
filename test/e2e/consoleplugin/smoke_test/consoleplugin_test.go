package consoleplugin_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"

	logger "github.com/ViaQ/logerr/v2/log"
	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/visualization/console"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/client"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

func init() {
	test.Must(configv1.AddToScheme(scheme.Scheme))
	if s := os.Getenv("LOG_LEVEL"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			log.SetLogger(logger.NewLogger("functional-framework", logger.WithVerbosity(n)))
		}
	}
}

var ctx = context.TODO()

var _ = Describe("[ConsolePlugin]", func() {
	var (
		c *client.Test
		r *console.Reconciler
	)

	cleanup := func() {
		_ = r.Delete(ctx)
		_ = c.Remove(testruntime.NewClusterLogForwarder())
		_ = c.Remove(testruntime.NewClusterLogging())
	}

	BeforeEach(func() {
		c = client.NewTest()
		r = console.NewReconciler(
			c.ControllerRuntimeClient(),
			console.NewConfig(testruntime.NewClusterLogging(), "lokiService", "korrel8r", "korrel8r", []string{}), nil)
		cleanup() // Clear out objects left behind by previous tests.
	})

	AfterEach(func() {
		cleanup()
		c.Close()
	})

	It("activates logging view if log type is lokistack", func() {
		cl := testruntime.NewClusterLogging()
		cl.Spec = loggingv1.ClusterLoggingSpec{
			ManagementState: loggingv1.ManagementStateManaged,
			LogStore: &loggingv1.LogStoreSpec{
				Type: loggingv1.LogStoreTypeLokiStack,
				LokiStack: loggingv1.LokiStackStoreSpec{
					Name: "testing-stack",
				},
			},
			Visualization: &loggingv1.VisualizationSpec{
				Type:   loggingv1.VisualizationTypeOCPConsole,
				Kibana: &loggingv1.KibanaSpec{},
			},
			Collection: &loggingv1.CollectionSpec{
				Type: loggingv1.LogCollectionTypeFluentd,
			},
		}
		ExpectOK(c.Recreate(cl))

		// cl should deploy the console plugin
		cp := &consolev1alpha1.ConsolePlugin{}
		runtime.Initialize(cp, cl.Namespace, "logging-view-plugin")
		Eventually(func() error { return c.Get(cp) }, time.Minute, time.Second).Should(Succeed())
		By("Got console plugin: " + test.JSONLine(cp))

		Eventually(func() error {
			pods := &corev1.PodList{}
			selector := client.MatchingLabels{constants.LabelK8sName: console.Name}
			if err := c.List(pods, selector, client.InNamespace(cl.GetNamespace())); err != nil {
				return err
			}
			for _, pod := range pods.Items {
				for _, cond := range pod.Status.Conditions {
					if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
						return nil
					}
				}
			}
			return fmt.Errorf("No pods found in ready condition")
		},
			time.Minute, time.Second).Should(Succeed())
	})
})
