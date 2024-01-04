package source

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"regexp"
	"strings"
)

type KubernetesLogs struct {
	framework.ComponentID
	Desc               string
	IncludePaths       string
	ExcludePaths       string
	ExtraLabelSelector string
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
{{- if gt (len .IncludePaths) 0}}
include_paths_glob_patterns = {{.IncludePaths}}
{{- end}}
{{- if gt (len .ExcludePaths) 0 }}
exclude_paths_glob_patterns = {{.ExcludePaths}}
{{- end}}
{{- if gt (len .ExtraLabelSelector) 0 }}
extra_label_selector = "{{.ExtraLabelSelector}}"
{{- end}}
pod_annotation_fields.pod_labels = "kubernetes.labels"
pod_annotation_fields.pod_namespace = "kubernetes.namespace_name"
pod_annotation_fields.pod_annotations = "kubernetes.annotations"
pod_annotation_fields.pod_uid = "kubernetes.pod_id"
pod_annotation_fields.pod_node_name = "hostname"
namespace_annotation_fields.namespace_uid = "kubernetes.namespace_id"
rotate_wait_ms = 5000
{{end}}`
}

// NewKubernetesLogs element which always excludes temp and gzip files
func NewKubernetesLogs(id, includes, excludes string) KubernetesLogs {
	return KubernetesLogs{
		ComponentID:  id,
		Desc:         "Logs from containers (including openshift containers)",
		IncludePaths: includes,
		ExcludePaths: excludes,
	}
}

const (
	crioPodPathFmt       = `"/var/log/pods/%s"`
	crioNamespacePathFmt = `"/var/log/pods/%s/*/*.log"`
	crioContainerPathFmt = `"/var/log/pods/*/%s/*.log"`
	crioPathExtFmt       = `"/var/log/pods/*/*/*.%s"`
)

// ContainerPathGlobFrom formats a list of kubernetes container file paths to include/exclude for
// collection given a list of namespaces and containers and return a string that
// is in a form directly usable by a vector kubernetes_log config. The result is
// a set of file paths assumed to be at the well known location and structure of
// CRIO pod logs.
// The format rules:
//
//	namespaces:
//	  namespace:     /var/log/pods/namespace_*/*/*.log
//	  **namespace:   /var/log/pods/*namespace_*/*/*.log
//	  **name*pace**: /var/log/pods/*name*pace*/*/*.log
//	  namespace**:   /var/log/pods/namespace*/*/*.log
//	containers:
//	  container:    /var/log/pods/*/container/*.log
//	  *cont**iner*:    /var/log/pods/*/*cont*iner*/*.log
//	  cont**iner*:    /var/log/pods/*/cont*iner*/*.log
func ContainerPathGlobFrom(namespaces, containers []string, extensions ...string) string {
	paths := []string{}
	for _, n := range namespaces {
		paths = append(paths, fmt.Sprintf(crioNamespacePathFmt, normalizeNamespace(n)))
	}
	for _, c := range containers {
		paths = append(paths, fmt.Sprintf(crioContainerPathFmt, collapseWildcards(c)))
	}
	for _, e := range extensions {
		paths = append(paths, fmt.Sprintf(crioPathExtFmt, collapseWildcards(e)))
	}
	if len(paths) == 0 {
		return ""
	}
	return joinContainerPathsForVector(paths)
}

type ContainerPathGlobBuilder struct {
	paths []string
}

func NewContainerPathGlobBuilder() *ContainerPathGlobBuilder {
	return &ContainerPathGlobBuilder{}
}

// AddOther takes an argument and joins it with the well known container path
func (b *ContainerPathGlobBuilder) AddOther(other ...string) *ContainerPathGlobBuilder {
	for _, n := range other {
		b.paths = append(b.paths, fmt.Sprintf(crioPodPathFmt, collapseWildcards(n)))
	}
	return b
}
func (b *ContainerPathGlobBuilder) AddNamespaces(namespaces ...string) *ContainerPathGlobBuilder {
	for _, n := range namespaces {
		b.paths = append(b.paths, fmt.Sprintf(crioNamespacePathFmt, normalizeNamespace(n)))
	}
	return b
}
func (b *ContainerPathGlobBuilder) AddContainers(containers ...string) *ContainerPathGlobBuilder {
	for _, c := range containers {
		b.paths = append(b.paths, fmt.Sprintf(crioContainerPathFmt, collapseWildcards(c)))
	}
	return b
}
func (b *ContainerPathGlobBuilder) AddExtensions(extensions ...string) *ContainerPathGlobBuilder {
	for _, e := range extensions {
		b.paths = append(b.paths, fmt.Sprintf(crioPathExtFmt, collapseWildcards(e)))
	}
	return b
}
func (b *ContainerPathGlobBuilder) Build() string {
	paths := joinContainerPathsForVector(b.paths)
	if paths == "" {
		return ""
	}
	return paths
}

func joinContainerPathsForVector(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	return fmt.Sprintf("[%s]", strings.Join(paths, ", "))
}

func normalizeNamespace(ns string) string {
	if !strings.Contains(ns, "*") {
		return fmt.Sprintf("%s_*", ns)
	}
	return fmt.Sprintf("%s_*", collapseWildcards(ns))
}

var consecutiveWildcards = regexp.MustCompile(`\*+`)

func collapseWildcards(entry string) string {
	return consecutiveWildcards.ReplaceAllString(entry, "*")
}
