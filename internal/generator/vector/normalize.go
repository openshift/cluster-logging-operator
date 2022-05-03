package vector

import (
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

const (
	FixLogLevel = `
level = "unknown"
if match!(.message,r'(Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn")'){
  level = "warn"
} else if match!(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"'){
  level = "info"
} else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"'){
  level = "error"
} else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"'){
  level = "debug"
}
.level = level
`

	RemoveFile = `
del(.file)
`
	RemoveSourceType = `
del(.source_type)
`
	RemoveStream = `
del(.stream)
`
	RemovePodIPs = `
del(.kubernetes.pod_ips)
`
)

func NormalizeLogs(spec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
	types := generator.GatherSources(spec, op)
	var el []generator.Element = make([]generator.Element, 0)
	if types.Has(logging.InputNameApplication) || types.Has(logging.InputNameInfrastructure) {
		el = append(el, NormalizeContainerLogs("raw_container_logs", "container_logs")...)
	}
	if types.Has(logging.InputNameInfrastructure) {
		el = append(el, NormalizeJournalLogs("raw_journal_logs", "journal_logs")...)
	}
	return el
}

func NormalizeContainerLogs(inLabel, outLabel string) []generator.Element {
	return []generator.Element{
		Remap{
			ComponentID: outLabel,
			Inputs:      helpers.MakeInputs(inLabel),
			VRL: strings.Join(helpers.TrimSpaces([]string{
				FixLogLevel,
				RemoveFile,
				RemoveSourceType,
				RemoveStream,
				RemovePodIPs,
			}), "\n\n"),
		},
	}
}

func NormalizeJournalLogs(inLabel, outLabel string) []generator.Element {
	return []generator.Element{
		Remap{
			ComponentID: outLabel,
			Inputs:      helpers.MakeInputs(inLabel),
			VRL:         SrcPassThrough,
		},
	}
}
