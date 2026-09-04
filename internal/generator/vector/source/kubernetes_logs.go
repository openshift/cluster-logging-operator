package source

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
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
rotate_wait_secs = 5
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
	crioPodPathFmt                   = `"/var/log/pods/%s"`
	crioNamespacePathFmt             = `"/var/log/pods/%s/*/*.log"`
	crioNamespaceAndContainerPathFmt = `"/var/log/pods/%s/%s/*.log"`
	crioNamespaceContainerCombined   = `"/var/log/pods/%s/*.log"`
	crioContainerPathFmt             = `"/var/log/pods/*/%s/*.log"`
	crioPathExtFmt                   = `"/var/log/pods/*/*/*.%s"`
	crioEverything                   = `["/var/log/pods/*/*/*.log"]`
)

// ContainerPathGlobFrom formats a list of kubernetes container file paths to include/exclude for
// collection given a list of namespaces and containers and return a string that
// is in a form directly usable by a vector kubernetes_log config. The result is
// a set of file paths assumed to be at the well known location and structure of
// CRIO pod logs. Container and namespace includes are combined in their various permutations
// as well as excludes to allow collection (or exclusion) of specific containers from specific namespaces
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
	containers   *sets.String
	namespaces   *sets.String
	nsContainers *sets.String
	paths        []string
}

type NamespaceContainer struct {
	Namespace string
	Container string
}

func NewContainerPathGlobBuilder() *ContainerPathGlobBuilder {
	return &ContainerPathGlobBuilder{
		containers:   sets.NewString(),
		namespaces:   sets.NewString(),
		nsContainers: sets.NewString(),
	}
}

func (b *ContainerPathGlobBuilder) AddCombined(ncs ...NamespaceContainer) *ContainerPathGlobBuilder {
	for _, n := range ncs {
		if n.Namespace == "" {
			n.Namespace = "*"
		}
		if n.Container == "" {
			n.Container = "*"
		}
		combined := fmt.Sprintf("%s/%s", normalizeNamespace(n.Namespace), collapseWildcards(n.Container))
		b.nsContainers.Insert(combined)
	}
	return b
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
		if n != "" {
			b.namespaces.Insert(normalizeNamespace(n))
		}
	}
	return b
}
func (b *ContainerPathGlobBuilder) AddContainers(containers ...string) *ContainerPathGlobBuilder {
	for _, c := range containers {
		if c != "" {
			b.containers.Insert(collapseWildcards(c))
		}
	}
	return b
}
func (b *ContainerPathGlobBuilder) AddExtensions(extensions ...string) *ContainerPathGlobBuilder {
	for _, e := range extensions {
		b.paths = append(b.paths, fmt.Sprintf(crioPathExtFmt, collapseWildcards(e)))
	}
	return b
}
func (b *ContainerPathGlobBuilder) Build(excludeNSFromContainers ...string) string {
	namespacesNotToCombine := sets.NewString()
	for _, ns := range excludeNSFromContainers {
		namespacesNotToCombine.Insert(normalizeNamespace(ns))
	}
	uniq := sets.NewString(b.paths...)
	if b.nsContainers.Len() > 0 {
		for _, ncs := range b.nsContainers.List() {
			uniq.Insert(fmt.Sprintf(crioNamespaceContainerCombined, ncs))
		}
	}
	switch {
	case b.containers.Len() == 0 && b.namespaces.Len() > 0:
		for _, n := range b.namespaces.List() {
			uniq.Insert(fmt.Sprintf(crioNamespaceAndContainerPathFmt, n, "*"))
		}
	case b.namespaces.Len() == 0 && b.containers.Len() > 0:
		for _, c := range b.containers.List() {
			uniq.Insert(fmt.Sprintf(crioNamespaceAndContainerPathFmt, "*", c))
		}
	case b.namespaces.Len() > 0 && b.containers.Len() > 0:
		for _, c := range b.containers.List() {
			for _, n := range b.namespaces.List() {
				cFinal := c
				if namespacesNotToCombine.Has(n) {
					cFinal = "*"
				}
				uniq.Insert(fmt.Sprintf(crioNamespaceAndContainerPathFmt, n, cFinal))
			}
		}
	}
	paths := uniq.List()
	sort.Strings(paths)
	pathString := joinContainerPathsForVector(paths)
	if pathString == "" || pathString == crioEverything {
		return ""
	}
	return pathString
}

func joinContainerPathsForVector(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	return fmt.Sprintf("[%s]", strings.Join(paths, ", "))
}

func normalizeNamespace(ns string) string {
	if len(ns) == 1 && ns == "*" {
		return ns
	}
	if !strings.Contains(ns, "*") {
		return fmt.Sprintf("%s_*", ns)
	}
	return fmt.Sprintf("%s_*", collapseWildcards(ns))
}

var consecutiveWildcards = regexp.MustCompile(`\*+`)

func collapseWildcards(entry string) string {
	return consecutiveWildcards.ReplaceAllString(entry, "*")
}
