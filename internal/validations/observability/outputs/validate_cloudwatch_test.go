package outputs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("[internal][validations] ClusterLogForwarder will validate maxWrite for Cloudwatch output", func() {
	var (
		spec obs.OutputSpec
	)
	BeforeEach(func() {
		spec = obs.OutputSpec{
			Name:       "cloudwatchOutput",
			Type:       obs.OutputTypeCloudwatch,
			Cloudwatch: &obs.Cloudwatch{},
		}
	})

	Context("#validateCloudwatchMaxWrite", func() {

		It("should pass validation when no tuning is set", func() {
			Expect(validateCloudwatchMaxWrite(spec)).To(BeEmpty())
		})

		It("should pass validation when maxWrite is nil", func() {
			spec.Cloudwatch.Tuning = &obs.CloudwatchTuningSpec{}
			Expect(validateCloudwatchMaxWrite(spec)).To(BeEmpty())
		})

		It("should pass validation when maxWrite is under 1MB", func() {
			spec.Cloudwatch.Tuning = &obs.CloudwatchTuningSpec{
				BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
					MaxWrite: utils.GetPtr(resource.MustParse("500Ki")),
				},
			}
			Expect(validateCloudwatchMaxWrite(spec)).To(BeEmpty())
		})

		It("should pass validation when maxWrite is exactly 1MB", func() {
			spec.Cloudwatch.Tuning = &obs.CloudwatchTuningSpec{
				BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
					MaxWrite: utils.GetPtr(resource.MustParse("1048576")),
				},
			}
			Expect(validateCloudwatchMaxWrite(spec)).To(BeEmpty())
		})

		It("should fail validation when maxWrite exceeds 1MB", func() {
			spec.Cloudwatch.Tuning = &obs.CloudwatchTuningSpec{
				BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
					MaxWrite: utils.GetPtr(resource.MustParse("10M")),
				},
			}
			Expect(validateCloudwatchMaxWrite(spec)).ToNot(BeEmpty())
		})

		It("should pass validation when output is not Cloudwatch", func() {
			spec = obs.OutputSpec{
				Name: "httpOutput",
				Type: obs.OutputTypeHTTP,
				HTTP: &obs.HTTP{},
			}
			Expect(validateCloudwatchMaxWrite(spec)).To(BeEmpty())
		})
	})
})
