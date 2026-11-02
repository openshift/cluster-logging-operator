package container

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/receivers/operators"
)

// NewOperators creates a complete operator pipeline for parsing CRI-O container logs
// and mapping them to non-deprecated attributes as defined in:
// https://github.com/rhobs/observability-data-model/blob/main/cluster-logging.md
//
// CRI-O log format: <timestamp> <stream> <tags> <log>
// Example: 2023-03-15T10:30:45.123456789+00:00 stdout F log message here
//
// File path format: /var/log/pods/<namespace>_<pod_name>_<uid>/<container_name>/<restart_count>.log
func NewOperators(clusterUID, nodeName, logSource, logType string) []operators.Operator {
	return []operators.Operator{
		// Step 1: Parse CRI-O format
		{
			Type:   operators.OperatorTypeRegexParser,
			ID:     "parser-crio",
			Output: "extract_metadata_from_filepath",
			Config: map[string]interface{}{
				"regex": `^(?P<time>[^ Z]+) (?P<stream>stdout|stderr) (?P<logtag>[^ ]*) ?(?P<log>.*)$`,
				"timestamp": map[string]interface{}{
					"parse_from":  "attributes.time",
					"layout_type": "gotime",
					"layout":      "2006-01-02T15:04:05.999999999Z07:00",
				},
			},
		},

		// Step 2: Extract Kubernetes metadata from file path
		// Path: /var/log/pods/<namespace>_<pod_name>_<uid>/<container_name>/<restart_count>.log
		{
			Type:   operators.OperatorTypeRegexParser,
			ID:     "extract_metadata_from_filepath",
			Output: "move_log_to_body",
			Config: map[string]interface{}{
				"parse_from": "attributes[\"log.file.path\"]",
				"regex":      `^.*\/(?P<namespace>[^_]+)_(?P<pod_name>[^_]+)_(?P<uid>[^\/]+)\/(?P<container_name>[^\/]+)\/(?P<restart_count>\d+)\.log$`,
				"cache": map[string]interface{}{
					"size": 128,
				},
			},
		},

		// Step 3: Move parsed log content to body
		{
			Type:   operators.OperatorTypeMove,
			ID:     "move_log_to_body",
			Output: "set_iostream",
			Config: map[string]interface{}{
				"from": "attributes.log",
				"to":   "body",
			},
		},

		// Step 4: Set log.iostream (non-deprecated attribute)
		{
			Type:   operators.OperatorTypeMove,
			ID:     "set_iostream",
			Output: "move_namespace_to_resource",
			Config: map[string]interface{}{
				"from": "attributes.stream",
				"to":   "attributes[\"log.iostream\"]",
			},
		},

		// Step 5: Move namespace to resource attributes (k8s.namespace.name)
		{
			Type:   operators.OperatorTypeMove,
			ID:     "move_namespace_to_resource",
			Output: "move_pod_name_to_resource",
			Config: map[string]interface{}{
				"from": "attributes.namespace",
				"to":   "resource[\"k8s.namespace.name\"]",
			},
		},

		// Step 6: Move pod name to resource attributes (k8s.pod.name)
		{
			Type:   operators.OperatorTypeMove,
			ID:     "move_pod_name_to_resource",
			Output: "move_pod_uid_to_resource",
			Config: map[string]interface{}{
				"from": "attributes.pod_name",
				"to":   "resource[\"k8s.pod.name\"]",
			},
		},

		// Step 7: Move pod UID to resource attributes (k8s.pod.uid)
		{
			Type:   operators.OperatorTypeMove,
			ID:     "move_pod_uid_to_resource",
			Output: "move_container_name_to_resource",
			Config: map[string]interface{}{
				"from": "attributes.uid",
				"to":   "resource[\"k8s.pod.uid\"]",
			},
		},

		// Step 8: Move container name to resource attributes (k8s.container.name)
		{
			Type:   operators.OperatorTypeMove,
			ID:     "move_container_name_to_resource",
			Output: "remove_restartcount",
			Config: map[string]interface{}{
				"from": "attributes.container_name",
				"to":   "resource[\"k8s.container.name\"]",
			},
		},

		// Step 9: Move restart count to resource attributes (k8s.container.restart_count)
		{
			Type:   operators.OperatorTypeRemove,
			ID:     "remove_restartcount",
			Output: "remove_logtag",
			Config: map[string]interface{}{
				"field": "attributes.restart_count",
			},
		},

		// Step 10: Remove the logtag field (not needed in final output)
		{
			Type:   operators.OperatorTypeRemove,
			ID:     "remove_logtag",
			Output: "remove_time",
			Config: map[string]interface{}{
				"field": "attributes.logtag",
			},
		},

		// Step 11: Remove the time field (already parsed to timestamp)
		{
			Type: operators.OperatorTypeRemove,
			ID:   "remove_time",
			Config: map[string]interface{}{
				"field": "attributes.time",
			},
		},
		{
			Type: operators.OperatorTypeAdd,
			ID:   "add_node_name",
			Config: map[string]interface{}{
				"field": "resource[\"k8s.node.name\"]",
				"value": nodeName,
			},
		},
		{
			Type:   operators.OperatorTypeAdd,
			ID:     "add_cluster_uid",
			Output: "add_log_source",
			Config: map[string]interface{}{
				"field": "resource[\"openshift.cluster.uid\"]",
				"value": clusterUID,
			},
		},
		{
			Type:   operators.OperatorTypeAdd,
			ID:     "add_log_source",
			Output: "add_log_type",
			Config: map[string]interface{}{
				"field": "resource[\"openshift.log.source\"]",
				"value": logSource,
			},
		},
		{
			Type: operators.OperatorTypeAdd,
			ID:   "add_log_type",
			Config: map[string]interface{}{
				"field": `resource["openshift.log.type"]`,
				"value": logType,
			},
		},
	}
}
