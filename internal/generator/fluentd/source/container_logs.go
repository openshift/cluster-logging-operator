package source

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"strconv"
)

type ContainerLogs struct {
	generator.OutLabel
	Desc         string
	Paths        string
	ExcludePaths string
	PosFile      string
	Tunings      *logging.FluentdInFileSpec
}

func (cl ContainerLogs) ReadLinesLimit() string {
	if cl.Tunings == nil || cl.Tunings.ReadLinesLimit <= 0 {
		return ""
	}
	return "\n  read_lines_limit " + strconv.Itoa(cl.Tunings.ReadLinesLimit)
}

func (cl ContainerLogs) Name() string {
	return "inputContainerSourceTemplate"
}

func (cl ContainerLogs) Template() string {
	return `{{define "` + cl.Name() + `" -}}
# {{.Desc}}
<source>
  @type tail
  @id container-input
  path {{.Paths}}
  exclude_path {{.ExcludePaths}}
  pos_file "{{.PosFile}}"
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  {{- .ReadLinesLimit }}
  skip_refresh_on_startup true
  @label @{{.OutLabel}}
  <parse>
    @type regexp
    expression /^(?<@timestamp>[^\s]+) (?<stream>stdout|stderr) (?<logtag>[F|P]) (?<message>.*)$/
    time_key '@timestamp'
    keep_time_key true
  </parse>
</source>
{{end}}`
}
