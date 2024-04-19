// go:build !fluentd
package syslog

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/syslog"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	"strings"
	"time"
)

const (
	sysLogInputName = `syslog-source`
)

var _ = Describe("[Functional][Inputs][SysLog] Functional tests", func() {

	if testfw.LogCollectionType != logging.LogCollectionTypeVector {
		defer GinkgoRecover()
		Skip("skip for non-vector")
	}

	host := "acme.com"
	app := "Excel"
	pid := 6868
	msg := "Choose Your Destiny"
	msgId := "ID9"
	var (
		framework *functional.CollectorFunctionalFramework
		RFC3164   = fmt.Sprintf("<30>Apr 26 15:14:30 %s %s[%d]: %s", host, app, pid, msg)
		RFC5425   = fmt.Sprintf("<39>1 2024-04-26T15:36:49.214+03:00 %s %s %d %s - %s", host, app, pid, msgId, msg)
	)

	BeforeEach(func() {
		Expect(testfw.LogCollectionType).To(Equal(logging.LogCollectionTypeVector))
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(logging.LogCollectionTypeVector)
		framework.VisitConfig = func(conf string) string {
			return strings.Replace(conf, "enabled = true", "enabled = false", 2) // turn off TLS for testing
		}
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("", func() {
		It("should sending log record to the SysLog input", func() {
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInputWithVisitor(sysLogInputName,
					func(spec *logging.InputSpec) {
						spec.Receiver = &logging.ReceiverSpec{
							Type: logging.ReceiverTypeSyslog,
							ReceiverTypeSpec: &logging.ReceiverTypeSpec{
								Syslog: &logging.SyslogReceiver{},
							},
						}
					}).ToHttpOutput()

			Expect(framework.DeployWithVisitor(
				func(b *runtime.PodBuilder) error {
					err := syslog.AddSenderContainer(b)
					Expect(err).To(BeNil(), "Expected no errors deploying syslog sender")
					return framework.AddVectorHttpOutputWithConfig(b, framework.Forwarder.Spec.Outputs[0], "", nil, functional.InfrastructureLogFile)
				}),
			).To(BeNil())

			err := syslog.WriteToSyslogInputWithNetcat(framework, sysLogInputName, RFC3164)
			Expect(err).To(BeNil(), "Expected no errors writing to SysLog input")
			time.Sleep(5 * time.Second)
			err = syslog.WriteToSyslogInputWithNetcat(framework, sysLogInputName, RFC5425)
			Expect(err).To(BeNil(), "Expected no errors writing to SysLog input")

			raw, err := framework.ReadFileFromWithRetryInterval("http", functional.InfrastructureLogFile, time.Second)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).ToNot(BeEmpty())
			lines := strings.Split(raw, "\n")
			Expect(len(lines)).To(BeEquivalentTo(2))
			for _, line := range lines {
				record := map[string]interface{}{}
				Expect(json.Unmarshal([]byte(line), &record)).To(BeNil())
				message, ok := record["message"].(string)
				Expect(ok).To(BeTrue())
				Expect(message).To(BeEquivalentTo(msg))
				hostname, ok := record["hostname"].(string)
				Expect(ok).To(BeTrue())
				Expect(hostname).To(BeEquivalentTo(host))
				procid, ok := record["procid"].(float64)
				Expect(ok).To(BeTrue())
				Expect(procid).To(BeEquivalentTo(pid))
				logType, ok := record["log_type"].(string)
				Expect(ok).To(BeTrue())
				Expect(logType).To(BeEquivalentTo("infrastructure"))
			}
		})

	})

})
