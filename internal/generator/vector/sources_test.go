package vector

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Vector Config Generation", func() {
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return generator.MergeElements(
			LogSources(&clfspec, constants.OpenshiftNS, op),
		)
	}
	DescribeTable("Source(s)", helpers.TestGenerateConfWith(f),
		Entry("Only Application", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
# Logs from containers (including openshift containers)
[sources.raw_container_logs]
type = "kubernetes_logs"
glob_minimum_cooldown_ms = 15000
auto_partial_merge = true
exclude_paths_glob_patterns = ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
pod_annotation_fields.pod_labels = "kubernetes.labels"
pod_annotation_fields.pod_namespace = "kubernetes.namespace_name"
pod_annotation_fields.pod_annotations = "kubernetes.annotations"
pod_annotation_fields.pod_uid = "kubernetes.pod_id"
pod_annotation_fields.pod_node_name = "hostname"
`,
		}),
		Entry("Only Infrastructure", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameInfrastructure,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
# Logs from containers (including openshift containers)
[sources.raw_container_logs]
type = "kubernetes_logs"
glob_minimum_cooldown_ms = 15000
auto_partial_merge = true
exclude_paths_glob_patterns = ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
pod_annotation_fields.pod_labels = "kubernetes.labels"
pod_annotation_fields.pod_namespace = "kubernetes.namespace_name"
pod_annotation_fields.pod_annotations = "kubernetes.annotations"
pod_annotation_fields.pod_uid = "kubernetes.pod_id"
pod_annotation_fields.pod_node_name = "hostname"

[sources.raw_journal_logs]
type = "journald"
journal_directory = "/var/log/journal"
`,
		}),
		Entry("Only Audit", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameAudit,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
# Logs from host audit
[sources.raw_host_audit_logs]
type = "file"
include = ["/var/log/audit/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

# Logs from kubernetes audit
[sources.raw_k8s_audit_logs]
type = "file"
include = ["/var/log/kube-apiserver/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

# Logs from openshift audit
[sources.raw_openshift_audit_logs]
type = "file"
include = ["/var/log/oauth-apiserver/audit.log","/var/log/openshift-apiserver/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

# Logs from ovn audit
[sources.raw_ovn_audit_logs]
type = "file"
include = ["/var/log/ovn/acl-audit-log.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000
`,
		}),
		Entry("All Log Sources", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
							logging.InputNameInfrastructure,
							logging.InputNameAudit,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
# Logs from containers (including openshift containers)
[sources.raw_container_logs]
type = "kubernetes_logs"
glob_minimum_cooldown_ms = 15000
auto_partial_merge = true
exclude_paths_glob_patterns = ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
pod_annotation_fields.pod_labels = "kubernetes.labels"
pod_annotation_fields.pod_namespace = "kubernetes.namespace_name"
pod_annotation_fields.pod_annotations = "kubernetes.annotations"
pod_annotation_fields.pod_uid = "kubernetes.pod_id"
pod_annotation_fields.pod_node_name = "hostname"

[sources.raw_journal_logs]
type = "journald"
journal_directory = "/var/log/journal"

# Logs from host audit
[sources.raw_host_audit_logs]
type = "file"
include = ["/var/log/audit/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

# Logs from kubernetes audit
[sources.raw_k8s_audit_logs]
type = "file"
include = ["/var/log/kube-apiserver/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

# Logs from openshift audit
[sources.raw_openshift_audit_logs]
type = "file"
include = ["/var/log/oauth-apiserver/audit.log","/var/log/openshift-apiserver/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

# Logs from ovn audit
[sources.raw_ovn_audit_logs]
type = "file"
include = ["/var/log/ovn/acl-audit-log.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000
`,
		}),
	)
})
