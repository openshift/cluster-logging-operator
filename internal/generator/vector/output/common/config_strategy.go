package common

// ConfigStrategy abstracts the generator specific sections from the domain
type ConfigStrategy interface {
	VisitAcknowledgements(a Acknowledgments) Acknowledgments
	VisitBatch(b Batch) Batch
	VisitRequest(r Request) Request
	VisitBuffer(b Buffer) Buffer

	// VisitSink allows setting top-level sink parameters
	VisitSink(s SinkConfig)
}

// SinkConfig is an abstraction to set common root level configuration parameters of a sink (e.g. compression)
// Not all sinks may support all parameters
type SinkConfig interface {

	// SetCompression of the sink
	SetCompression(algo string)
}
