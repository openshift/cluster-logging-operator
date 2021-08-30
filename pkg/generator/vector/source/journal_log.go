package source

import (
	. "github.com/openshift/cluster-logging-operator/pkg/generator"
)

const JournalLogTemplate = `
{{define "inputSourceJournalTemplate" -}}
[sources.{{.ComponentID}}]
  type = "journald"
{{end}}`

type JournalLog = ConfLiteral
