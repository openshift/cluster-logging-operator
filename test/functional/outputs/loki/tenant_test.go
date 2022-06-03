package loki

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[Functional][Outputs][Loki] Tenancy", func() {
	var (
		f *functional.CollectorFunctionalFramework
		l *loki.Receiver
	)
	BeforeEach(func() {
		f = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)
		l = loki.NewReceiver(f.Namespace, "loki-server").EnableMultiTenant()
		ExpectOK(l.Create(f.Test.Client))
		// Set up the common template forwarder configuration.
		f.Forwarder.Spec.Outputs = []logging.OutputSpec{{
			Name: logging.OutputTypeLoki,
			Type: logging.OutputTypeLoki,
			URL:  l.InternalURL("").String(),
		}}
		f.Forwarder.Spec.Pipelines = []logging.PipelineSpec{{
			InputRefs:  []string{logging.InputNameApplication},
			OutputRefs: []string{logging.OutputTypeLoki},
			Labels:     map[string]string{"logging": "logging-value"},
		}}
	})

	AfterEach(func() {
		f.Cleanup()
	})

	DescribeTable("tenant settings",
		func(lokiSpec *logging.Loki, tenantID string) {
			f.Forwarder.Spec.Outputs[0].OutputTypeSpec.Loki = lokiSpec
			ExpectOK(f.Deploy())
			// Verify we can query for logs with the correct tenant ID
			msg := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), "hello")
			ExpectOK(f.WriteMessagesToApplicationLog(msg, 1))
			result, err := l.QueryUntil(`{log_type="application"}`, tenantID, 1)
			ExpectOK(err)
			Expect(result).NotTo(BeEmpty())
			records := result[0].Records()
			Expect(records).To(HaveLen(1))
			Expect(records[0]["message"]).To(Equal("hello"))
		},

		Entry("uses default tenantKey log_type", nil, "application"),
		Entry("uses fixed tenant", &logging.Loki{TenantID: "my-tenant"}, "my-tenant"),
		Entry("uses logging label as tenant", &logging.Loki{TenantKey: "openshift.labels.logging"}, "logging-value"),
	)

	It("gets no logs with invalid tenant", func() {
		ExpectOK(f.Deploy())
		ExpectOK(f.WritesApplicationLogs(2))

		// Query with the wrong tenant ID returns nothing.
		result, err := l.Query(`{log_type="application"}`, "bad-tenant", 1)
		ExpectOK(err)
		Expect(result).To(BeEmpty())

		// In multi-tenant mode, qury with an empty tenant-ID returns an error.
		result, err = l.Query(`{log_type="application"}`, "", 1)
		Expect(err).To(HaveOccurred())
		Expect(result).To(BeEmpty())
	})
})
