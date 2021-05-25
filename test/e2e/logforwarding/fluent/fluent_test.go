package fluent_test

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/helpers/fluentd"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

const message = "My life is my message"

var _ = Describe("[ClusterLogForwarder]", func() {
	const basePort = 24224
	var (
		c          *client.Test
		f          *Fixture
		portOffset int
		logTypes   = loggingv1.ReservedInputNames.UnsortedList()
	)
	BeforeEach(func() { c = client.NewTest(); f = NewFixture(c.NS.Name, message) })
	AfterEach(func() { c.Close() })

	Context("with app/infra/audit receiver", func() {
		BeforeEach(func() {
			for _, logType := range logTypes {
				f.Receiver.AddSource(&fluentd.Source{Name: logType, Type: "forward", Port: basePort + portOffset})
				portOffset++
			}
		})

		It("forwards application logs only", func() {
			clf := f.ClusterLogForwarder
			addPipeline(clf, f.Receiver.Sources["application"])
			f.Create(c.Client)
			r := f.Receiver.Sources["application"].TailReader()
			for i := 0; i < 10; {
				l, err := r.ReadLine()
				ExpectOK(err)
				Expect(l).To(ContainSubstring(`"viaq_index_name":"app`)) // Only app logs
				if strings.Contains(l, message) {
					i++ // Count our own app messages, ignore others.
				}
			}
			for _, name := range []string{"infrastructure", "audit"} {
				Expect(f.Receiver.Sources[name].HasOutput()).To(BeFalse())
			}
		})

		It("forwards infrastructure logs only", func() {
			clf := f.ClusterLogForwarder
			addPipeline(clf, f.Receiver.Sources["infrastructure"])
			f.Create(c.Client)
			r := f.Receiver.Sources["infrastructure"].TailReader()
			l, err := r.ReadLine()
			ExpectOK(err)
			Expect(l).To(ContainSubstring(`"viaq_index_name":"inf`)) // Only infra logs
		})

		It("forwards different types to different outputs with labels", func() {
			clf := f.ClusterLogForwarder
			for _, name := range logTypes {
				s := f.Receiver.Sources[name]
				clf.Spec.Outputs = append(clf.Spec.Outputs, loggingv1.OutputSpec{
					Name: s.Name,
					Type: "fluentdForward",
					URL:  fmt.Sprintf("tcp://%v:%v", s.Host(), s.Port),
				})
				clf.Spec.Pipelines = append(clf.Spec.Pipelines, loggingv1.PipelineSpec{
					InputRefs:  []string{s.Name},
					OutputRefs: []string{s.Name},
					Labels:     map[string]string{"log-type": s.Name},
				})
			}
			f.Create(c.Client)
			time.Sleep(30 * time.Second)
			for _, name := range logTypes {
				name := name // Don't bind to range variable
				r := f.Receiver.Sources[name].TailReader()
				Expect(r.ReadLine()).To(SatisfyAny(Equal(""), ContainSubstring(fmt.Sprintf(`"log-type":%q`, name))))
			}
		})
	})
})

func addPipeline(clf *loggingv1.ClusterLogForwarder, s *fluentd.Source) {
	clf.Spec.Outputs = append(clf.Spec.Outputs, loggingv1.OutputSpec{
		Name: s.Name,
		Type: "fluentdForward",
		URL:  fmt.Sprintf("tcp://%v:%v", s.Host(), s.Port),
	})
	clf.Spec.Pipelines = append(clf.Spec.Pipelines,
		loggingv1.PipelineSpec{
			InputRefs:  []string{s.Name},
			OutputRefs: []string{s.Name},
		})
}
