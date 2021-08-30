package source

import (
	. "github.com/openshift/cluster-logging-operator/pkg/generator"
)

type ContainerLogs struct {
	OutLabel
	Desc         string
	Paths        string
	ExcludePaths string
	PosFile      string
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
  skip_refresh_on_startup true
  @label @{{.OutLabel}}
  <parse>
    @type multi_format
    <pattern>
      format json
      time_format '%Y-%m-%dT%H:%M:%S.%N%Z'
      keep_time_key true
    </pattern>
    <pattern>
      format regexp
      expression /^(?<time>[^\s]+) (?<stream>stdout|stderr)( (?<logtag>.))? (?<log>.*)$/
      time_format '%Y-%m-%dT%H:%M:%S.%N%:z'
      keep_time_key true
    </pattern>
  </parse>
</source>
{{end}}`
}
