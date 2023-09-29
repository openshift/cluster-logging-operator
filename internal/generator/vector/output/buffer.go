package output

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"reflect"
	"strings"
)

type bufferConf struct {
	ComponentID string
	Settings    map[string]interface{}
}

func BufferConf(o logging.OutputSpec, op generator.Options) []generator.Element {
	conf := bufferConf{
		ComponentID: helpers.FormatComponentID(o.Name),
		Settings:    map[string]interface{}{},
	}
	o.Tuning.Range(func(key, value string) {
		if strings.HasPrefix(key, "buffer.") {
			key = strings.TrimPrefix(key, "buffer.")
			formatter := "%s"
			if logging.VariantKind(value) == reflect.String {
				formatter = "%q"
			}
			conf.Settings[key] = fmt.Sprintf(formatter, value)
		}
	})
	return []generator.Element{
		conf,
	}
}

func (t bufferConf) Name() string {
	return "outputBufferConf"
}

func (t bufferConf) Template() string {
	return `
{{define "outputBufferConf" -}}
[sinks.{{.ComponentID}}.buffer]
{{- range $key, $value := .Settings }}
{{ $key }} = {{ $value -}}
{{ end }}
{{ end }}`
}
