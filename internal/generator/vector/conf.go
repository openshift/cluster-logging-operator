package vector

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	corev1 "k8s.io/api/core/v1"
)

//nolint:govet // using declarative style
func Conf(clspec *logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Section {
	return []generator.Section{
		{
			Sources(clfspec, op),
			"Set of all input sources",
		},
		{
			NormalizeLogs(clfspec, op),
			"set 'level' field, add metadata",
		},
		{
			SourcesToInputs(clfspec, op),
			"",
		},
		{
			InputsToPipelines(clfspec, op),
			"",
		},
		{
			Outputs(clspec, secrets, clfspec, op),
			"vector outputs",
		},
	}
}

func NormalizeLogs(spec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
	types := generator.GatherSources(spec, op)
	var el []generator.Element = make([]generator.Element, 0)
	if types.Has(logging.InputNameApplication) || types.Has(logging.InputNameInfrastructure) {
		el = append(el, Normalize("raw_container_logs", "container_logs")...)
	}
	if types.Has(logging.InputNameInfrastructure) {
		el = append(el, Normalize("raw_journal_logs", "journal_logs")...)
	}
	return el
}

func Normalize(inLabel, outLabel string) []generator.Element {
	return []generator.Element{
		generator.ConfLiteral{
			TemplateName: "transforms",
			InLabel:      inLabel,
			OutLabel:     outLabel,
			TemplateStr: `
{{define "transforms" -}}
[transforms.{{.OutLabel}}]
type = "remap"
inputs = ["{{.InLabel}}"]
source = '''
  level = "unknown"
  if match(.message,r'(Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn")'){
	level = "warn"
  } else if match(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"'){
	level = "info"
  } else if match(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"'){
	level = "error"
  } else if match(.message, r'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"'){
    level = "critical"
  } else if match(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"'){
	level = "debug"
  }
  .level = level

  .pipeline_metadata.collector.name = "vector"
  .pipeline_metadata.collector.version = "0.14.1"
  ip4, err = get_env_var("NODE_IPV4")
  .pipeline_metadata.collector.ipaddr4 = ip4
  received, err = format_timestamp(now(),"%+")
  .pipeline_metadata.collector.received_at = received
  .pipeline_metadata.collector.error = err
 '''
{{end}}
`,
		},
	}
}

func Sink() []generator.Element {
	return []generator.Element{
		generator.ConfLiteral{
			TemplateName: "stdout",
			TemplateStr: `
{{define "stdout" -}}
[sinks.console_logs]
  inputs = ["ocp_sys.infra","ocp_sys.app_special", "ocp_sys.app"]
  type = "console"
  encoding = "json"
{{end}}
			`,
		},
	}
}
