package otlp

import (
	"encoding/json"
)

type Logs struct {
	ResourceLogs []ResourceLog `json:"resourceLogs,omitempty"`
}

type ResourceLog struct {
	Resource  Resource   `json:"resource,omitempty"`
	ScopeLogs []ScopeLog `json:"scopeLogs,omitempty"`
}

type Resource struct {
	Attributes []Attribute `json:"attributes,omitempty"`
}

type Attribute struct {
	Key   string         `json:"key,omitempty"`
	Value AttributeValue `json:"value,omitempty"`
}

type Scope struct {
	Name       string      `json:"name,omitempty"`
	Version    string      `json:"version,omitempty"`
	Attributes []Attribute `json:"attributes,omitempty"`
}

type ScopeLog struct {
	Scope      Scope       `json:"scope,omitempty"`
	LogRecords []LogRecord `json:"logRecords,omitempty"`
}

type LogRecord struct {
	TimeUnixNano         string      `json:"timeUnixNano,omitempty"`
	ObservedTimeUnixNano string      `json:"observedTimeUnixNano,omitempty"`
	SeverityNumber       int         `json:"severityNumber,omitempty"`
	TraceID              string      `json:"traceId,omitempty"`
	SpanID               string      `json:"spanId,omitempty"`
	Body                 StringValue `json:"body,omitempty"`
	Attributes           []Attribute `json:"attributes,omitempty"`
}

type AttributeValue struct {
	StringValue string      `json:"stringValue,omitempty"`
	Bool        bool        `json:"boolValue,omitempty"`
	Int         int         `json:"intValue,omitempty"`
	Float       float64     `json:"doubleValue,omitempty"`
	Array       ArrayValue  `json:"arrayValue,omitempty"`
	Map         KVListValue `json:"kvlistValue,omitempty"`
}

type StringValue struct {
	StringValue string `json:"stringValue,omitempty"`
}

type ArrayValue struct {
	Values []StringValue `json:"values,omitempty"`
}

type KVListValue struct {
	Values []AttributeValue `json:"values,omitempty"`
}

func ParseLogs(in string) (Logs, error) {
	logs := Logs{}
	if in == "" {
		return logs, nil
	}
	err := json.Unmarshal([]byte(in), &logs)
	if err != nil {
		return logs, err
	}

	return logs, nil
}
