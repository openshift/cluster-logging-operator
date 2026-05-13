package outputs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("[internal][validations] ClusterLogForwarder will validate maxWrite for AzureLogsIngestion output", func() {
	var (
		spec obs.OutputSpec
	)
	BeforeEach(func() {
		spec = obs.OutputSpec{
			Name:               "azureOutput",
			Type:               obs.OutputTypeAzureLogsIngestion,
			AzureLogsIngestion: &obs.AzureLogsIngestion{},
		}
	})

	Context("#validateAzureLogsIngestionMaxWrite", func() {

		It("should pass validation when no tuning is set", func() {
			Expect(validateAzureLogsIngestionMaxWrite(spec)).To(BeEmpty())
		})

		It("should pass validation when maxWrite is nil", func() {
			spec.AzureLogsIngestion.Tuning = &obs.AzureLogsIngestionTuningSpec{}
			Expect(validateAzureLogsIngestionMaxWrite(spec)).To(BeEmpty())
		})

		It("should pass validation when maxWrite is under 1MB", func() {
			spec.AzureLogsIngestion.Tuning = &obs.AzureLogsIngestionTuningSpec{
				BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
					MaxWrite: utils.GetPtr(resource.MustParse("500Ki")),
				},
			}
			Expect(validateAzureLogsIngestionMaxWrite(spec)).To(BeEmpty())
		})

		It("should pass validation when maxWrite is exactly 1MB", func() {
			spec.AzureLogsIngestion.Tuning = &obs.AzureLogsIngestionTuningSpec{
				BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
					MaxWrite: utils.GetPtr(resource.MustParse("1000000")),
				},
			}
			Expect(validateAzureLogsIngestionMaxWrite(spec)).To(BeEmpty())
		})

		It("should fail validation when maxWrite exceeds 1MB", func() {
			spec.AzureLogsIngestion.Tuning = &obs.AzureLogsIngestionTuningSpec{
				BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
					MaxWrite: utils.GetPtr(resource.MustParse("10M")),
				},
			}
			Expect(validateAzureLogsIngestionMaxWrite(spec)).ToNot(BeEmpty())
		})

		It("should pass validation when output is not AzureLogsIngestion", func() {
			spec = obs.OutputSpec{
				Name: "httpOutput",
				Type: obs.OutputTypeHTTP,
				HTTP: &obs.HTTP{},
			}
			Expect(validateAzureLogsIngestionMaxWrite(spec)).To(BeEmpty())
		})
	})
})
