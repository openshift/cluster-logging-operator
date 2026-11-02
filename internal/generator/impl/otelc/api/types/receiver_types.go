package types

type ReceiverType string

const (
	ReceiverTypeFileLog ReceiverType = "filelog"
)

// Receiver is an OpenTelemetry Collector receiver for signals coming into the collector
type Receiver interface {
	ID() string
	ReceiverType() ReceiverType
}
