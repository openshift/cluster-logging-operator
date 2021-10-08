package vector

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Vector Config Generation", func() {
	var f = func(clspec logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		sources := LogSources(&clfspec, op)
		merged := make([]generator.Element, 0)
		for _, source := range sources {
			element := generator.Element(source)
			merged = append(merged, element)
		}
		return merged
	}
	DescribeTable("Source(s)", generator.TestGenerateConfWith(f),
		Entry("Only Application", generator.ConfGenerateTest{
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
			ExpectedConf: `# Logs from containers
[sources.kubernetes_logs]
  type = "kubernetes_logs"
  auto_partial_merge = true
  exclude_paths_glob_patterns = ["/var/log/pods/collector-*_openshift-logging_*.log", "/var/log/pods/elasticsearch-*_openshift-logging_*.log", "/var/log/pods/kibana-*_openshift-logging_*.log"]`,
		}),
		Entry("Only Infrastructure", generator.ConfGenerateTest{
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
			ExpectedConf: `# Logs from containers
[sources.kubernetes_logs]
  type = "kubernetes_logs"
  auto_partial_merge = true
  exclude_paths_glob_patterns = ["/var/log/pods/collector-*_openshift-logging_*.log", "/var/log/pods/elasticsearch-*_openshift-logging_*.log", "/var/log/pods/kibana-*_openshift-logging_*.log"]

# Logs from journald
[sources.journald]
  type = "journald"`,
		}),
		PEntry("Only Audit", generator.ConfGenerateTest{
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
[sources.host_audit_logs]
  type = "file"
  ignore_older_secs = 600
  include = ["/var/log/audit/audit.log"]

# Logs from kubernetes audit
[sources.k8s_audit_logs]
  type = "file"
  ignore_older_secs = 600
  include = ["/var/log/kube-apiserver/audit.log"]

# Logs from openshift audit
[sources.openshift_audit_logs]
  type = "file"
  ignore_older_secs = 600
  include = ["/var/log/oauth-apiserver.audit.log"]
`,
		}),
		Entry("Application and Infrastructure Log Sources", generator.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
							logging.InputNameInfrastructure,
							//logging.InputNameAudit,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `# Logs from containers
[sources.kubernetes_logs]
  type = "kubernetes_logs"
  auto_partial_merge = true
  exclude_paths_glob_patterns = ["/var/log/pods/collector-*_openshift-logging_*.log", "/var/log/pods/elasticsearch-*_openshift-logging_*.log", "/var/log/pods/kibana-*_openshift-logging_*.log"]

# Logs from journald
[sources.journald]
  type = "journald"`,
		}))
})
