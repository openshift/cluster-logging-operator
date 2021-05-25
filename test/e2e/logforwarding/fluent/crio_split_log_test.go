package fluent_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/e2e/logforwarding/fluent"
	"github.com/openshift/cluster-logging-operator/test/helpers/fluentd"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	"strings"
)

//E2E test for checking correct reassembly after splitting big logs by CRI-O.
//LogGenerator container will put into stdout prepared huge Java stacktrace in JSON format.
//CRI-O will spilt it in to the parts, like:
// 	2021-04-06T00:17:09.669794202Z stdout P First part of JSON log entry
//	2021-04-06T00:17:09.669794202Z stdout P Second part of JSON log entry
//	2021-04-06T00:17:10.113242941Z stderr F Last part of JSON log entry
//After that in should be assembled by fluent-plugin-concat and stored in original state.

var _ = Describe("[ClusterLogForwarder]", func() {

	var (
		c *client.Test
		f *Fixture
	)

	BeforeEach(func() { c = client.NewTest() })
	AfterEach(func() { c.Close() })

	Context("When a Java container logs a multi-line stack trace", func() {
		It("should be forwarded as a single message", func() {
			containerName := "log-generator-" + strings.ToLower(string(utils.GetRandomWord(7)))
			f = NewFixture(c.NS.Name, "")
			f.LogGenerator = runtime.NewOneLineLogGenerator(c.NS.Name, containerName,
				strings.ReplaceAll(fluent.JsonJavaStackTrace, "\"", "\\\"")) // todo
			clf := f.ClusterLogForwarder
			f.Receiver.AddSource(&fluentd.Source{Name: "application", Type: "forward", Port: 24224})
			addPipeline(clf, f.Receiver.Sources["application"])
			f.Create(c.Client)
			r := f.Receiver.Sources["application"].TailReader()
			line, err := r.ReadLine()
			matchers.ExpectOK(err)
			for !strings.Contains(line, fmt.Sprintf("\"container_name\":\"%s\"", containerName)) {
				line, err = r.ReadLine()
				matchers.ExpectOK(err)
			}
			// Hack for easy parsing output a proper json array
			out := "[" + strings.TrimRight(strings.Replace(line, "\n", ",", -1), ",") + "]"
			logs, err := types.ParseLogs(out)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			Expect(logs[0].Message).Should(Equal(fluent.JsonJavaStackTrace))
		})
	})
})
