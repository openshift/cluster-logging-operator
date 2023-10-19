package viaq

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"strings"
)

func JournalLogs(inputs, id string) []generator.Element {
	return []generator.Element{
		elements.Remap{
			ComponentID: id,
			Inputs:      inputs,
			VRL: strings.Join(helpers.TrimSpaces([]string{
				normalize.ClusterID,
				normalize.AddJournalLogTag,
				normalize.DeleteJournalLogFields,
				normalize.FixJournalLogLevel,
				normalize.AddHostName,
				normalize.SystemK,
				normalize.SystemT,
				normalize.SystemU,
				normalize.AddTime,
				normalize.FixTimestampField,
			}), "\n\n"),
		},
	}
}

func DropJournalDebugLogs(inputs, id string) []generator.Element {
	return []generator.Element{
		elements.Filter{
			ComponentID: id,
			Inputs:      inputs,
			Condition:   `.PRIORITY != \"7\" && .PRIORITY != 7`,
		},
	}
}
