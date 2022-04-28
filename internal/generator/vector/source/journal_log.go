package source

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
)

const JournalLogTemplate = `
{{define "inputSourceJournalTemplate" -}}
[sources.{{.ComponentID}}]
type = "journald"
journal_directory = "/var/log/journal"
{{end}}`

type JournalLog = generator.ConfLiteral
