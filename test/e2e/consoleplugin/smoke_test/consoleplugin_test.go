package consoleplugin_test

import (
	"context"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	logger "github.com/ViaQ/logerr/v2/log"
	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/consoleplugin"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/client"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
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
		r *consoleplugin.Reconciler
	)

	BeforeEach(func() {
		c = client.NewTest()
		r = consoleplugin.NewReconciler(
			c.ControllerRuntimeClient(),
			consoleplugin.NewConfig(testruntime.NewClusterLogging(), "lokiService"))
		// Clear out any previosly-existing objects
		_ = c.Remove(testruntime.NewClusterLogging())
		_ = c.Remove(testruntime.NewClusterLogForwarder())
		_ = r.Delete(ctx)
	})

	AfterEach(func() { c.Close() })

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
		}
		ExpectOK(c.Recreate(cl))

		// cl should deploy the console plugin
		cp := &consolev1alpha1.ConsolePlugin{}
		runtime.Initialize(cp, cl.Namespace, "logging-view-plugin")
		Eventually(func() error { return c.Get(cp) }, time.Minute, time.Second).Should(Succeed())
		By("Got console plugin: " + test.JSONLine(cp))

		// Get the console web page, check our plugin is listed.
		con := &configv1.Console{ObjectMeta: metav1.ObjectMeta{Name: "cluster"}}
		ExpectOK(c.Get(con))
		cfg := config.GetConfigOrDie()
		http, err := rest.HTTPClientFor(cfg)
		ExpectOK(err)

		u := con.Status.ConsoleURL + "/monitoring/logs"
		By("Checking console URL: " + u)
		Eventually(func() string {
			resp, err := http.Get(u)
			ExpectOK(err)
			b, err := ioutil.ReadAll(resp.Body)
			ExpectOK(err)
			By("Got console response: " + string(b))
			return string(b)
		}, time.Minute, time.Second).Should(MatchRegexp(`"consolePlugins":\[.*"logging-view-plugin"`))
	})
})
