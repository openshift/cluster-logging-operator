package output

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"k8s.io/apimachinery/pkg/api/resource"
	"time"
)

type fakeSink struct {
	Compression string
}

func (s *fakeSink) SetCompression(algo string) {
	s.Compression = algo
}

var _ = Describe("ConfigStrategy for tuning Outputs", func() {

	const (
		ID = "id"
	)

	Context("Compression", func() {
		It("should not set the compression when none", func() {
			output := NewOutput(logging.OutputSpec{
				Tuning: &logging.OutputTuningSpec{
					Compression: "none",
				},
			}, nil, framework.NoOptions)
			sink := &fakeSink{}
			output.VisitSink(sink)
			Expect(sink.Compression).To(BeEmpty())
		})
		It("should not set the compression when empty", func() {
			output := NewOutput(logging.OutputSpec{
				Tuning: &logging.OutputTuningSpec{
					Compression: "",
				},
			}, nil, framework.NoOptions)
			sink := &fakeSink{}
			output.VisitSink(sink)
			Expect(sink.Compression).To(BeEmpty())
		})
		It("should set the compression when not empty or none", func() {
			output := NewOutput(logging.OutputSpec{
				Tuning: &logging.OutputTuningSpec{
					Compression: "gzip",
				},
			}, nil, framework.NoOptions)
			sink := &fakeSink{}
			output.VisitSink(sink)
			Expect(sink.Compression).To(Equal("gzip"))
		})
	})
	Context("MaxRetryDuration", func() {

		It("should rely upon the defaults and generate nothing when zero", func() {
			output := NewOutput(logging.OutputSpec{
				Tuning: &logging.OutputTuningSpec{}}, nil, nil)
			Expect(``).To(EqualConfigFrom(common.NewRequest(ID, output)))
		})

		It("should set request.retry_max_duration_secs for values greater then zero", func() {
			output := NewOutput(logging.OutputSpec{
				Tuning: &logging.OutputTuningSpec{
					MaxRetryDuration: utils.GetPtr(35 * 1000 * time.Millisecond),
				},
			}, nil, nil)

			Expect(`
[sinks.id.request]
retry_max_duration_secs = 35
`).To(EqualConfigFrom(common.NewRequest(ID, output)))

		})
	})
	Context("MinRetryDuration", func() {

		It("should rely upon the defaults and generate nothing when zero", func() {
			output := NewOutput(logging.OutputSpec{
				Tuning: &logging.OutputTuningSpec{}}, nil, nil)
			Expect(``).To(EqualConfigFrom(common.NewRequest(ID, output)))
		})

		It("should set request.retry_initial_backoff_secs for values greater then zero", func() {
			output := NewOutput(logging.OutputSpec{
				Tuning: &logging.OutputTuningSpec{
					MinRetryDuration: utils.GetPtr(25 * 1000 * time.Millisecond),
				},
			}, nil, nil)

			Expect(`
[sinks.id.request]
retry_initial_backoff_secs = 25
`).To(EqualConfigFrom(common.NewRequest(ID, output)))

		})
	})
	Context("MaxWrite", func() {

		It("should rely upon the defaults and generate nothing when zero", func() {
			output := NewOutput(logging.OutputSpec{
				Tuning: &logging.OutputTuningSpec{}}, nil, nil)
			Expect(``).To(EqualConfigFrom(common.NewBatch(ID, output)))
		})

		It("should set batch.max_bytes for values greater then zero", func() {
			output := NewOutput(logging.OutputSpec{
				Tuning: &logging.OutputTuningSpec{
					MaxWrite: utils.GetPtr(resource.MustParse("1Ki")),
				},
			}, nil, nil)

			Expect(`
[sinks.id.batch]
max_bytes = 1024
`).To(EqualConfigFrom(common.NewBatch(ID, output)))

		})
	})

	Context("when delivery is spec'd", func() {

		Context("AtLeastOnce", func() {
			var output = NewOutput(logging.OutputSpec{
				Tuning: &logging.OutputTuningSpec{
					Delivery: logging.OutputDeliveryModeAtLeastOnce,
				},
			}, nil, nil)
			It("should enable acknowledgments", func() {
				Expect(`
[sinks.id.acknowledgements]
enabled = true
`).To(EqualConfigFrom(common.NewAcknowledgments(ID, output)))

			})
			It("should block when the buffer becomes full", func() {
				Expect(`
[sinks.id.buffer]
when_full = "block"
`).To(EqualConfigFrom(common.NewBuffer(ID, output)))
			})
		})

		Context("AtMostOnce", func() {

			var output = NewOutput(logging.OutputSpec{
				Tuning: &logging.OutputTuningSpec{
					Delivery: logging.OutputDeliveryModeAtMostOnce,
				},
			}, nil, nil)

			It("should not enable acknowledgements and not be present", func() {
				Expect("").To(EqualConfigFrom(common.NewAcknowledgments(ID, output)))
				Expect("").To(EqualConfigFrom(common.NewAcknowledgments(ID, nil)), "exp it to handle a nil config strategy")
			})
			It("should drop_newest when the buffer becomes full", func() {
				Expect(`
[sinks.id.buffer]
when_full = "drop_newest"
`).To(EqualConfigFrom(common.NewBuffer(ID, output)))
			})
		})
	})
})
