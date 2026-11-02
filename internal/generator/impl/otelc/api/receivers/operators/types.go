package operators

// OperatorType represents the type of operator in the log processing pipeline
type OperatorType string

// Parser Operators - extract structured data from log entries
const (
	OperatorTypeRegexParser     OperatorType = "regex_parser"
	OperatorTypeJSONParser      OperatorType = "json_parser"
	OperatorTypeJSONArrayParser OperatorType = "json_array_parser"
	OperatorTypeCSVParser       OperatorType = "csv_parser"
	OperatorTypeSyslogParser    OperatorType = "syslog_parser"
	OperatorTypeSeverityParser  OperatorType = "severity_parser"
	OperatorTypeTimeParser      OperatorType = "time_parser"
	OperatorTypeTraceParser     OperatorType = "trace_parser"
	OperatorTypeURIParser       OperatorType = "uri_parser"
	OperatorTypeKeyValueParser  OperatorType = "key_value_parser"
	OperatorTypeContainer       OperatorType = "container"
	OperatorTypeScopeNameParser OperatorType = "scope_name_parser"
)

// Transformer Operators - modify log data
const (
	OperatorTypeMove         OperatorType = "move"
	OperatorTypeAdd          OperatorType = "add"
	OperatorTypeCopy         OperatorType = "copy"
	OperatorTypeRemove       OperatorType = "remove"
	OperatorTypeFlatten      OperatorType = "flatten"
	OperatorTypeRecombine    OperatorType = "recombine"
	OperatorTypeRegexReplace OperatorType = "regex_replace"
	OperatorTypeRetain       OperatorType = "retain"
	OperatorTypeAssignKeys   OperatorType = "assign_keys"
	OperatorTypeSanitizeUTF8 OperatorType = "sanitize_utf8"
	OperatorTypeUnquote      OperatorType = "unquote"
)

// Filter/Router Operators - control flow
const (
	OperatorTypeFilter OperatorType = "filter"
	OperatorTypeRouter OperatorType = "router"
)

// Utility Operators
const (
	OperatorTypeNoop OperatorType = "noop"
)

// Input Operators (typically not used in filelog operators array, but defined for completeness)
const (
	OperatorTypeFileInput       OperatorType = "file_input"
	OperatorTypeJournaldInput   OperatorType = "journald_input"
	OperatorTypeStdin           OperatorType = "stdin"
	OperatorTypeSyslogInput     OperatorType = "syslog_input"
	OperatorTypeTCPInput        OperatorType = "tcp_input"
	OperatorTypeUDPInput        OperatorType = "udp_input"
	OperatorTypeWindowsEventLog OperatorType = "windows_eventlog_input"
)

// Output Operators (typically not used in filelog operators array, but defined for completeness)
const (
	OperatorTypeFileOutput OperatorType = "file_output"
	OperatorTypeStdout     OperatorType = "stdout"
)
