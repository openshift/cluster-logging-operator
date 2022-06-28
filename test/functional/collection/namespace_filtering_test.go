package collection

import (
	"runtime"
	"time"

	"github.com/openshift/cluster-logging-operator/test/framework/functional"

	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

var _ = Describe("[Functional][Collection] Namespace filtering", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)
	var (
		instance      = functional.NewCollectorFunctionalFramework()
		appNamespace1 = "application-ns1"
		appNamespace2 = "application-ns2"
	)

	BeforeEach(func() {
		functional.NewClusterLogForwarderBuilder(instance.Forwarder).
			FromInputWithVisitor("application-logs", func(spec *logging.InputSpec) {
				spec.Application = &logging.Application{
					Namespaces: []string{appNamespace1},
				}
			}).Named("test-app").
			ToFluentForwardOutput()
		Expect(instance.Deploy()).To(BeNil())
	})
	It("should send logs from one namespace only", func() {

		msg := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), "my message")
		Expect(instance.WriteMessagesToNamespace(msg, appNamespace1, 1)).To(Succeed())
		Expect(instance.WriteMessagesToNamespace(msg, appNamespace2, 1)).To(Succeed())

		logs, err := instance.ReadApplicationLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Error fetching logs from %s: %v", logging.OutputTypeFluentdForward, err)
		Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", logging.OutputTypeFluentdForward)

		// verify only appNamespace1 logs appear in Application logs
		for _, log := range logs {
			Expect(log.Kubernetes.NamespaceName).To(Equal(appNamespace1))
		}
	})

})
