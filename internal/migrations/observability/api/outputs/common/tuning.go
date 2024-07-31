package common

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

func MapBaseOutputTuning(outTuneSpec logging.OutputTuningSpec) *obs.BaseOutputTuningSpec {
	obsBaseTuningSpec := &obs.BaseOutputTuningSpec{
		MaxWrite:         outTuneSpec.MaxWrite,
		MinRetryDuration: outTuneSpec.MinRetryDuration,
		MaxRetryDuration: outTuneSpec.MaxRetryDuration,
	}

	switch outTuneSpec.Delivery {
	case logging.OutputDeliveryModeAtLeastOnce:
		obsBaseTuningSpec.Delivery = obs.DeliveryModeAtLeastOnce
	case logging.OutputDeliveryModeAtMostOnce:
		obsBaseTuningSpec.Delivery = obs.DeliveryModeAtMostOnce
	}

	return obsBaseTuningSpec
}
