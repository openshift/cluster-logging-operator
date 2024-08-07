package logfilesmetricexporter

import (
	"context"
	_ "embed"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	"k8s.io/apimachinery/pkg/util/wait"
	"regexp"
	"strings"
	"time"
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

	It("should be deployed by the operator and producing metrics", func() {
		e2e.AddCleanup(func() error {
			return oc.Literal().From("oc -n openshift-logging delete --ignore-not-found logfilemetricexporter instance").Output()
		})
		err = createLFME(validCR)
		Expect(err).ToNot(HaveOccurred())
		Expect(e2e.WaitForDaemonSet(constants.OpenshiftNS, constants.LogfilesmetricexporterName)).To(Succeed())

		args := []string{"-k", "-s", fmt.Sprintf("https://%s.%s.svc:2112/metrics", constants.LogfilesmetricexporterName, constants.OpenshiftNS)}
		err = wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Second*30, true, func(context.Context) (done bool, err error) {
			out, err := oc.Exec().WithNamespace(constants.OpenshiftNS).Pod(fmt.Sprintf("ds/%s", constants.LogfilesmetricexporterName)).WithCmd("curl", args...).Run()
			Expect(err).ToNot(HaveOccurred(), out)
			log.V(5).Info("Polling metrics", "result", out)
			if !strings.Contains(out, "log_logged_bytes_total") {
				return false, nil
			}
			return regexp.MatchString(`log_logged_bytes_total{.*} [1-9][0-9]*`, out)
		})
		Expect(err).ToNot(HaveOccurred(), "Exp. to find log_logged_bytes_total being calculated")
	})
})
