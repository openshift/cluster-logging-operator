package logfilesmetricexporter

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	"k8s.io/apimachinery/pkg/util/wait"
)

//go:embed valid.yaml
var validCR string

//go:embed invalid.yaml
var inValidCR string

var _ = Describe("[e2e][logfilemetricexporter] LogFileMetricsExporter", func() {

	defer GinkgoRecover()

	var (
		err        error
		e2e        = framework.NewE2ETestFramework()
		createLFME = func(cr string) error {
			lfme := &v1alpha1.LogFileMetricExporter{}
			test.MustUnmarshal(cr, lfme)
			return e2e.Create(lfme)
		}
	)
	AfterEach(func() {
		e2e.Cleanup()
	})

	It("should reject any CR not named openshift-logging/instance", func() {
		err = createLFME(inValidCR)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(MatchRegexp("is invalid.*supported values.*instance"), "exp. the CR to be rejected because it is not THE singleton")
	})

	It("should serve metrics to authorized clients providing a valid bearer token", func() {
		e2e.AddCleanup(func() error {
			return oc.Literal().From("oc -n openshift-logging delete --ignore-not-found logfilemetricexporter instance").Output()
		})
		metricsReaderRoleName := fmt.Sprintf("%s-metrics-reader", constants.ClusterLoggingOperator)
		metricsReaderBindingName := fmt.Sprintf("%s-metrics-reader", constants.LogfilesmetricexporterName)
		metricsAuthRoleName := fmt.Sprintf("%s-metrics-auth", constants.LogfilesmetricexporterName)
		e2e.AddCleanup(func() error {
			return oc.Literal().From("oc delete --ignore-not-found clusterrole %s", metricsReaderRoleName).Output()
		})
		e2e.AddCleanup(func() error {
			return oc.Literal().From("oc delete --ignore-not-found clusterrolebinding %s", metricsReaderBindingName).Output()
		})
		// Delete the metrics auth ClusterRoleBinding
		// The LFME reconciles the ClusterRoleBinding and ClusterRole for metrics auth
		e2e.AddCleanup(func() error {
			return oc.Literal().From("oc delete --ignore-not-found clusterrolebinding %s", metricsAuthRoleName).Output()
		})

		err = createLFME(validCR)
		Expect(err).ToNot(HaveOccurred())
		Expect(e2e.WaitForDaemonSet(constants.OpenshiftNS, constants.LogfilesmetricexporterName)).To(Succeed())

		By("creating the metrics-reader ClusterRole")
		roleFilePath, err := filepath.Abs(filepath.Join("..", "..", "..", "config", "rbac", "metrics_reader_role.yaml"))
		Expect(err).ToNot(HaveOccurred(), "Failed to construct role file path")
		_, err = oc.Literal().From("oc apply -f %s", roleFilePath).Run()
		Expect(err).ToNot(HaveOccurred(), "Failed to create metrics-reader ClusterRole")

		By("creating a ClusterRoleBinding for the service account to allow access to metrics")
		_, err = oc.Literal().From("oc create clusterrolebinding %s --clusterrole=%s --serviceaccount=%s:%s",
			metricsReaderBindingName, metricsReaderRoleName, constants.OpenshiftNS, constants.LogfilesmetricexporterName).Run()
		Expect(err).ToNot(HaveOccurred(), "Failed to create ClusterRoleBinding")

		metricsURL := fmt.Sprintf("https://%s.%s.svc:2112/metrics", constants.LogfilesmetricexporterName, constants.OpenshiftNS)
		curlCmd := fmt.Sprintf(`curl -s -k -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" %s`, metricsURL)

		err = wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Second*30, true, func(context.Context) (done bool, err error) {
			out, err := oc.Exec().WithNamespace(constants.OpenshiftNS).
				Pod(fmt.Sprintf("ds/%s", constants.LogfilesmetricexporterName)).
				WithCmd("sh", "-c", curlCmd).Run()
			Expect(err).ToNot(HaveOccurred(), out)
			log.V(5).Info("Polling secure metrics", "result", out)
			if !strings.Contains(out, "log_logged_bytes_total") {
				return false, nil
			}
			return regexp.MatchString(`log_logged_bytes_total{.*} [1-9][0-9]*`, out)
		})
		Expect(err).ToNot(HaveOccurred(), "Exp. to scrape metrics with a bearer token")
	})
})
