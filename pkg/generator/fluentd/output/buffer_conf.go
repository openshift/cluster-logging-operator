package output

type BufferConfig struct {
	BufferKeys           []string
	BufferPath           string
	FlushMode            string
	FlushInterval        string
	FlushThreadCount     string
	RetryType            string
	RetryWait            string
	RetryMaxInterval     string
	RetryTimeout         string
	QueuedChunkLimitSize string
	TotalLimitSize       string
	ChunkLimitSize       string
	OverflowAction       string
}

func (bc BufferConfig) Name() string {
	return "bufferConfigTemplate"
}

func (bc BufferConfig) Template() string {
	return `{{define "` + bc.Name() + `" -}}
{{if .BufferKeys -}}
<buffer {{comma_separated .BufferKeys}}>
{{- else -}}
<buffer>
{{- end}}
  @type file
  path '{{.BufferPath}}'
  flush_mode {{.FlushMode}}
  {{.FlushInterval}}
  flush_thread_count {{.FlushThreadCount}}
  flush_at_shutdown true
  retry_type {{.RetryType}}
  retry_wait {{.RetryWait}}
  retry_max_interval {{.RetryMaxInterval}}
  retry_timeout {{.RetryTimeout}}
  queued_chunks_limit_size {{.QueuedChunkLimitSize}}
  total_limit_size {{.TotalLimitSize}}
  chunk_limit_size {{.ChunkLimitSize}}
  overflow_action {{.OverflowAction}}
</buffer>
{{- end}}
`
}
