package source

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"strings"
)

type KubernetesLogs struct {
	framework.ComponentID
	Desc         string
	ExcludePaths string
}

func (kl KubernetesLogs) Name() string {
	return "k8s_logs_template"
}

func (kl KubernetesLogs) Template() string {
	return `{{define "` + kl.Name() + `" -}}
# {{.Desc}}
[sources.{{.ComponentID}}]
type = "kubernetes_logs"
max_read_bytes = 3145728
glob_minimum_cooldown_ms = 15000
auto_partial_merge = true
exclude_paths_glob_patterns = {{.ExcludePaths}}
pod_annotation_fields.pod_labels = "kubernetes.labels"
pod_annotation_fields.pod_namespace = "kubernetes.namespace_name"
pod_annotation_fields.pod_annotations = "kubernetes.annotations"
pod_annotation_fields.pod_uid = "kubernetes.pod_id"
pod_annotation_fields.pod_node_name = "hostname"
namespace_annotation_fields.namespace_uid = "kubernetes.namespace_id"
rotate_wait_ms = 5000
{{end}}`
}

func NewKubernetesLogs(id, namespace string) KubernetesLogs {
	return KubernetesLogs{
		ComponentID:  id,
		Desc:         "Logs from containers (including openshift containers)",
		ExcludePaths: ExcludeContainerPaths(namespace),
	}
}

func LogFilesMetricExporterLogsPath(namespace string) string {
	return fmt.Sprintf("/var/log/pods/%s_%%s-*/*/*.log", namespace)
}

func ElasticSearchLogStoreLogsPath(namespace string) string {
	return fmt.Sprintf("/var/log/pods/%s_%%s-*/*/*.log", namespace)
}

func LokiLogStoreLogsPath(namespace string) string {
	return fmt.Sprintf("/var/log/pods/%s_*/%%s*/*.log", namespace)
}

func VisualizationLogsPath(namespace string) string {
	return fmt.Sprintf("/var/log/pods/%s_%%s-*/*/*.log", namespace)
}

func ExcludeContainerPaths(namespace string) string {
	return fmt.Sprintf("[%s]", strings.Join(
		[]string{
			fmt.Sprintf("%q", fmt.Sprintf(LogFilesMetricExporterLogsPath(namespace), constants.LogfilesmetricexporterName)),
			fmt.Sprintf("%q", fmt.Sprintf(ElasticSearchLogStoreLogsPath(namespace), constants.ElasticsearchName)),
			fmt.Sprintf("%q", fmt.Sprintf(LokiLogStoreLogsPath(namespace), constants.LokiName)),
			fmt.Sprintf("%q", fmt.Sprintf(VisualizationLogsPath(namespace), constants.KibanaName)),
			fmt.Sprintf("%q", fmt.Sprintf("/var/log/pods/%s_*/%s/*.log", namespace, "gateway")),
			fmt.Sprintf("%q", fmt.Sprintf("/var/log/pods/%s_*/%s/*.log", namespace, "opa")),
			fmt.Sprintf("%q", "/var/log/pods/*/*/*.gz"),
			fmt.Sprintf("%q", "/var/log/pods/*/*/*.tmp"),
		},
		", ",
	))
}
