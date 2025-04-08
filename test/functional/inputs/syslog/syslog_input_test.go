package syslog

import (
	"context"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/syslog"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"k8s.io/apimachinery/pkg/util/wait"
	"strings"
	"time"
)

const (
	sysLogInputName = `syslog-source`
)

var _ = Describe("[Functional][Inputs][SysLog] Functional tests", func() {

	const (
		host  = "acme.com"
		app   = "Excel"
		pid   = 6868
		msg   = "Choose Your Destiny"
		msgId = "ID9"
	)

	var (
		framework *functional.CollectorFunctionalFramework
		RFC3164   = fmt.Sprintf("<30>Apr 26 15:14:30 %s %s[%d]: %s", host, app, pid, msg)
		RFC5425   = fmt.Sprintf("<39>1 2024-04-26T15:36:49.214+03:00 %s %s %d %s - %s", host, app, pid, msgId, msg)
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
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
				FromInputName(sysLogInputName,
					func(spec *obs.InputSpec) {
						spec.Type = obs.InputTypeReceiver
						spec.Receiver = &obs.ReceiverSpec{
							Port: 8443,
							Type: obs.ReceiverTypeSyslog,
						}
					}).ToHttpOutput()
			Expect(framework.DeployWithVisitors(append(
				framework.AddOutputContainersVisitors(),
				func(b *runtime.PodBuilder) error {
					return syslog.AddSenderContainer(b)
				}))).To(BeNil())

			for _, msg := range []string{RFC3164, RFC5425} {
				err := syslog.WriteToSyslogInputWithNetcat(framework, sysLogInputName, msg)
				Expect(err).To(BeNil(), "Expected no errors writing to SysLog input")
			}

			var lines []string
			Expect(wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true, func(context.Context) (done bool, err error) {
				if lines, err = framework.ReadInfrastructureLogsFrom(string(obs.OutputTypeHTTP)); err != nil {
					return true, err
				}
				return len(lines) >= 2, nil
			})).To(Succeed(), "Expected no errors reading the logs and there to be at least 2")

			for _, line := range lines {
				record := map[string]interface{}{}
				Expect(json.Unmarshal([]byte(line), &record)).To(BeNil())
				Expect(record).To(HaveKeyWithValue("message", msg))
				Expect(record).To(HaveKeyWithValue("hostname", host))
				Expect(record).To(HaveKeyWithValue("procid", float64(pid)))
				Expect(record).To(HaveKeyWithValue("log_type", string(obs.InputTypeInfrastructure)))
			}
		})

	})

})
