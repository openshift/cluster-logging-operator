package types

type TransformType string

const (
	TransformTypeDetectExceptions TransformType = "detect_exceptions"
	TransformTypeFilter           TransformType = "filter"
	TransformTypeLogToMetric      TransformType = "log_to_metric"
	TransformTypeReduce           TransformType = "reduce"
	TransformTypeRemap            TransformType = "remap"
	TransformTypeRoute            TransformType = "route"
	TransformTypeThrottle         TransformType = "throttle"
)

type Transform interface {
	TransformType() TransformType
}
