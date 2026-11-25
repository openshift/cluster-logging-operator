package receivers

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/receivers/operators"
	types "github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"
)

// FileLog represents the OpenTelemetry Collector filelog receiver configuration
// See: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver
//
// Example:
//
//	fileLog := receivers.NewFileLog("", "/var/log/app/*.log", "/var/log/service/*.log")
//	fileLog.StartAt = "beginning"
//	fileLog.Encoding = "utf-8"
//	fileLog.MaxConcurrentFiles = 256
//	fileLog.Attributes = map[string]interface{}{
//		"log.type":    "application",
//		"environment": "production",
//	}
//	fileLog.Multiline = &receivers.Multiline{
//		LineStartPattern: "^\\d{4}-\\d{2}-\\d{2}",
//	}
//	fileLog.Operators = []operators.Operator{
//		{
//			Type: operators.OperatorTypeJSONParser,
//			ID:   "json_parse",
//		},
//	}
//	receiversMap := api.Receivers{
//		"file_log/app": fileLog,
//	}
//	data, err := yaml.Marshal(receiversMap)
//
// produces:
//
//	file_log/app:
//	    include:
//	        - /var/log/app/*.log
//	        - /var/log/service/*.log
//	    start_at: beginning
//	    encoding: utf-8
//	    max_concurrent_files: 256
//	    multiline:
//	        line_start_pattern: ^\d{4}-\d{2}-\d{2}
//	    attributes:
//	        environment: production
//	        log.type: application
//	    operators:
//	        - type: json_parser
//	          id: json_parse
//
// A retry configuration can be added via RetryOnFailure:
//
//	fileLog.RetryOnFailure = &receivers.RetryOnFailure{
//		Enabled:         true,
//		InitialInterval: "1s",
//		MaxInterval:     "30s",
//		MaxElapsedTime:  "5m",
//	}
//
// which serializes as:
//
//	retry_on_failure:
//	    enabled: true
//	    initial_interval: 1s
//	    max_interval: 30s
//	    max_elapsed_time: 5m
type FileLog struct {
	// id is the receiver identifier in the format "type" or "type/name"
	// It is not serialized to YAML
	id string

	// File selection
	Include          []string `yaml:"include"`                      // Required: list of file glob patterns
	Exclude          []string `yaml:"exclude,omitempty"`            // List of file glob patterns to exclude
	ExcludeOlderThan string   `yaml:"exclude_older_than,omitempty"` // Duration (e.g., "24h")

	// Reading behavior
	StartAt          string `yaml:"start_at,omitempty"`           // "beginning" or "end" (default: "end")
	PollInterval     string `yaml:"poll_interval,omitempty"`      // Duration (default: "200ms")
	ForceFlushPeriod string `yaml:"force_flush_period,omitempty"` // Duration (default: "500ms")
	Encoding         string `yaml:"encoding,omitempty"`           // Encoding type (default: "utf-8")
	Compression      string `yaml:"compression,omitempty"`        // "", "gzip", or "auto"
	OnTruncate       string `yaml:"on_truncate,omitempty"`        // "ignore", "read_whole_file", or "read_new"

	// File identification
	FingerprintSize string `yaml:"fingerprint_size,omitempty"` // Size (default: "1kb")

	// Buffer and size limits
	InitialBufferSize  string `yaml:"initial_buffer_size,omitempty"`   // Size (default: "16KiB")
	MaxLogSize         int64  `yaml:"max_log_size,omitempty"`          // Size (default: "1MiB")
	MaxLogSizeBehavior string `yaml:"max_log_size_behavior,omitempty"` // "split" or "truncate" (default: "split")

	// Concurrency
	MaxConcurrentFiles int `yaml:"max_concurrent_files,omitempty"` // Number (default: 1024)
	MaxBatches         int `yaml:"max_batches,omitempty"`          // Number (0 = no limit)

	// File operations
	DeleteAfterRead bool `yaml:"delete_after_read,omitempty"` // Requires feature gate
	AcquireFsLock   bool `yaml:"acquire_fs_lock,omitempty"`   // Unix only

	// Whitespace handling
	PreserveLeadingWhitespaces  bool `yaml:"preserve_leading_whitespaces,omitempty"`
	PreserveTrailingWhitespaces bool `yaml:"preserve_trailing_whitespaces,omitempty"`

	// File attributes
	IncludeFileName           bool `yaml:"include_file_name,omitempty"`             // Default: true
	IncludeFilePath           bool `yaml:"include_file_path,omitempty"`             // Default: false
	IncludeFileNameResolved   bool `yaml:"include_file_name_resolved,omitempty"`    // Default: false
	IncludeFilePathResolved   bool `yaml:"include_file_path_resolved,omitempty"`    // Default: false
	IncludeFileOwnerName      bool `yaml:"include_file_owner_name,omitempty"`       // Not Windows
	IncludeFileOwnerGroupName bool `yaml:"include_file_owner_group_name,omitempty"` // Not Windows
	IncludeFilePermissions    bool `yaml:"include_file_permissions,omitempty"`      // Not Windows
	IncludeFileRecordNumber   bool `yaml:"include_file_record_number,omitempty"`    // Default: false
	IncludeFileRecordOffset   bool `yaml:"include_file_record_offset,omitempty"`    // Default: false

	// Multiline configuration
	Multiline *Multiline `yaml:"multiline,omitempty"`

	// Header metadata parsing (requires feature gate)
	Header *Header `yaml:"header,omitempty"`

	// Ordering criteria
	OrderingCriteria *OrderingCriteria `yaml:"ordering_criteria,omitempty"`

	// Storage
	Storage        string `yaml:"storage,omitempty"`          // Storage extension ID
	PollsToArchive int    `yaml:"polls_to_archive,omitempty"` // Experimental

	// Retry on failure
	RetryOnFailure *RetryOnFailure `yaml:"retry_on_failure,omitempty"`

	// Metadata
	Attributes map[string]interface{} `yaml:"attributes,omitempty"` // Key-value pairs added to entry attributes
	Resource   map[string]interface{} `yaml:"resource,omitempty"`   // Key-value pairs added to entry resource

	// Operators (array of operator configurations)
	Operators []operators.Operator `yaml:"operators,omitempty"`
}

// Multiline configuration for multi-line log entries
type Multiline struct {
	LineStartPattern string `yaml:"line_start_pattern,omitempty"` // Regex (mutually exclusive with line_end_pattern)
	LineEndPattern   string `yaml:"line_end_pattern,omitempty"`   // Regex (mutually exclusive with line_start_pattern)
	OmitPattern      bool   `yaml:"omit_pattern,omitempty"`       // Whether to omit the pattern from the output
}

// Header metadata parsing configuration
type Header struct {
	Pattern           string               `yaml:"pattern"`            // Required: regex
	MetadataOperators []operators.Operator `yaml:"metadata_operators"` // Required: list of operators
}

// OrderingCriteria controls file processing order
type OrderingCriteria struct {
	Regex   string         `yaml:"regex,omitempty"`    // Regex with named capture groups
	GroupBy string         `yaml:"group_by,omitempty"` // Regex with named capture groups
	TopN    int            `yaml:"top_n,omitempty"`    // Number (default: 1)
	SortBy  []SortCriteria `yaml:"sort_by,omitempty"`
}

// SortCriteria defines sorting rules for file ordering
type SortCriteria struct {
	RegexKey  string `yaml:"regex_key,omitempty"` // Named capture group from regex
	SortType  string `yaml:"sort_type,omitempty"` // "numeric", "alphabetical", "timestamp", "mtime"
	Location  string `yaml:"location,omitempty"`  // For timestamp sort_type
	Format    string `yaml:"format,omitempty"`    // strptime format for timestamp sort_type
	Ascending bool   `yaml:"ascending,omitempty"` // Sort order (default: true)
}

// RetryOnFailure configures retry behavior for downstream errors
type RetryOnFailure struct {
	Enabled         bool   `yaml:"enabled,omitempty"`          // Default: false
	InitialInterval string `yaml:"initial_interval,omitempty"` // Duration (default: "1s")
	MaxInterval     string `yaml:"max_interval,omitempty"`     // Duration (default: "30s")
	MaxElapsedTime  string `yaml:"max_elapsed_time,omitempty"` // Duration (default: "5m")
}

// ID returns the receiver identifier in the format "type" or "type/name"
func (r *FileLog) ID() string {
	return r.id
}

// ReceiverType extracts the receiver type from the ID
func (r *FileLog) ReceiverType() types.ReceiverType {
	componentType, _ := types.ParseComponentID(r.id)
	return types.ReceiverType(componentType)
}

// NewFileLog creates a new FileLog receiver with the given name and include patterns
// If name is empty, the receiver ID will be just the type ("file_log")
// If name is provided, the receiver ID will be "file_log/name"
func NewFileLog(name string, include ...string) *FileLog {
	return &FileLog{
		id:      types.MakeComponentID(string(types.ReceiverTypeFileLog), name),
		Include: include,
	}
}
