package otlp

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

const (
	MaxEventsGroupByContainer = uint64(1)
	MaxEventsGroupBySource    = uint64(1)
	MaxEventsGroupByHost      = uint64(1)
	expireAfterMs             = 15000
)

func GroupByContainer(inputs ...string) types.Transform {
	return transforms.NewReduce(func(r *transforms.Reduce) {
		r.ExpireAfterMs = expireAfterMs
		r.MaxEvents = MaxEventsGroupByContainer
		r.GroupBy = []string{".openshift.cluster_id",
			".kubernetes.namespace_name", ".kubernetes.pod_name", ".kubernetes.container_name"}
		r.MergeStrategies = &transforms.MergeStrategies{
			LogRecords: transforms.MergeStrategiesLogRecordsArray,
			Resource:   transforms.MergeStrategiesResourceRetain,
		}
	}, inputs...)
}

func GroupBySource(inputs ...string) types.Transform {
	return transforms.NewReduce(func(r *transforms.Reduce) {
		r.ExpireAfterMs = expireAfterMs
		r.MaxEvents = MaxEventsGroupBySource
		r.GroupBy = []string{".openshift.cluster_id", ".openshift.log_type", ".openshift.log_source"}
		r.MergeStrategies = &transforms.MergeStrategies{
			LogRecords: transforms.MergeStrategiesLogRecordsArray,
			Resource:   transforms.MergeStrategiesResourceRetain,
		}
	}, inputs...)
}

func GroupByHost(inputs ...string) types.Transform {
	return transforms.NewReduce(func(r *transforms.Reduce) {
		r.ExpireAfterMs = expireAfterMs
		r.MaxEvents = MaxEventsGroupByHost
		r.GroupBy = []string{".openshift.cluster_id", ".openshift.hostname", ".openshift.log_type", ".openshift.log_source"}
		r.MergeStrategies = &transforms.MergeStrategies{
			LogRecords: transforms.MergeStrategiesLogRecordsArray,
			Resource:   transforms.MergeStrategiesResourceRetain,
		}
	}, inputs...)
}
