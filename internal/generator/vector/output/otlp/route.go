package otlp

import (
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

type Route struct {
	ComponentID string
	Desc        string
	Inputs      string
}

func (r Route) Name() string {
	return "routeTemplate"
}

func (r Route) Template() string {
	return `{{define "routeTemplate" -}}
{{if .Desc -}}
# {{.Desc}}
{{end -}}
[transforms.{{.ComponentID}}]
type = "route"
inputs = {{.Inputs}}
route.container = '.log_source == "container"'
route.node = '.log_source == "node"'
route.auditd = '.log_source == "auditd"'
route.kubeapi = '.log_source == "kubeAPI"'
route.openshiftapi = '.log_source == "openshiftAPI"'
route.ovn = '.log_source == "ovn"'
{{end}}
`
}

func RouteBySource(id string, inputs []string) Element {
	return Route{
		Desc:        "Route container, journal, and audit logs separately",
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
	}
}
