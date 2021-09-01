package output

import (
	. "github.com/openshift/cluster-logging-operator/pkg/generator"
)

type BufferConf struct {
	BufferKeys     []string
	BufferConfData Element
}

func (bc BufferConf) Name() string {
	return "bufferConfTemplate"
}

func (bc BufferConf) Template() string {
	return `{{define "` + bc.Name() + `" -}}
{{if .BufferKeys -}}
<buffer {{comma_separated .BufferKeys}}>
{{- else -}}
<buffer>
{{- end}}
{{compose_one .BufferConfData | indent 2}}
</buffer>
{{end}}`
}

type BufferConfData struct {
	BufferPath           string
	FlushMode            Element
	FlushInterval        Element
	FlushThreadCount     Element
	RetryType            Element
	RetryWait            Element
	RetryMaxInterval     Element
	RetryTimeout         Element
	QueuedChunkLimitSize Element
	TotalLimitSize       Element
	ChunkLimitSize       Element
	OverflowAction       Element
}

func (bc BufferConfData) Name() string {
	return "bufferConfDataTemplate"
}

func (bc BufferConfData) Template() string {
	return `{{define "` + bc.Name() + `" -}}
@type file
path '{{.BufferPath}}'
{{optional .FlushMode -}}
{{optional .FlushInterval -}}
{{optional .FlushThreadCount -}}
flush_at_shutdown true
{{optional .RetryType -}}
{{optional .RetryWait -}}
{{optional .RetryMaxInterval -}}
{{optional .RetryTimeout -}}
{{optional .QueuedChunkLimitSize -}}
{{optional .TotalLimitSize -}}
{{optional .ChunkLimitSize -}}
{{optional .OverflowAction -}}
{{end}}
`
}
