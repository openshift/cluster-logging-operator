package source

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

const JournalLogTemplate = `
{{define "inputSourceJournalTemplate" -}}
[sources.{{.ComponentID}}]
type = "journald"
journal_directory = "/var/log/journal"
{{end}}`

type JournalLog = framework.ConfLiteral

func NewJournalLog(id string) JournalLog {
	return JournalLog{
		ComponentID:  id,
		Desc:         "Logs from linux journal",
		TemplateName: "inputSourceJournalTemplate",
		TemplateStr:  JournalLogTemplate,
	}
}
