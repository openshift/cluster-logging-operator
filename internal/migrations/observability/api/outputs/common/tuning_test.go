package common

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

var _ = Describe("#MapTuning", func() {
	It("should map logging base output tuning to observability base tuning", func() {
		loggingOutputBaseTune := logging.OutputTuningSpec{
			Delivery:         logging.OutputDeliveryModeAtLeastOnce,
			MaxWrite:         utils.GetPtr(resource.MustParse("10M")),
			MinRetryDuration: utils.GetPtr(time.Duration(1)),
			MaxRetryDuration: utils.GetPtr(time.Duration(5)),
		}
		obsOutputBaseTune := &obs.BaseOutputTuningSpec{
			Delivery:         obs.DeliveryModeAtLeastOnce,
			MaxWrite:         utils.GetPtr(resource.MustParse("10M")),
			MinRetryDuration: utils.GetPtr(time.Duration(1)),
			MaxRetryDuration: utils.GetPtr(time.Duration(5)),
		}
		actualObsBaseTune := MapBaseOutputTuning(loggingOutputBaseTune)

		Expect(actualObsBaseTune).To(Equal(obsOutputBaseTune))
	})
})
