package source

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func NewHttpSource(id, inputName string, input obs.InputSpec) (framework.Element, string) {
	return HttpReceiver{
		ID:            id,
		InputName:     inputName,
		ListenAddress: helpers.ListenOnAllLocalInterfacesAddress(),
		ListenPort:    input.Receiver.Port,
		Format:        string(input.Receiver.HTTP.Format),
	}, id
}

type HttpReceiver struct {
	ID            string
	InputName     string
	ListenAddress string
	ListenPort    int32
	Format        string
}

func (HttpReceiver) Name() string {
	return "httpReceiver"
}

func (i HttpReceiver) Template() string {
	return `
{{define "` + i.Name() + `" -}}
[sources.{{.ID}}]
type = "http_server"
address = "{{.ListenAddress}}:{{.ListenPort}}"
decoding.codec = "json"
{{end}}
`
}

func NewItemsTransform(id, inputs string) (framework.Element, string) {
	itemsID := helpers.MakeID(id, "items")
	return elements.Remap{
		ComponentID: itemsID,
		Inputs:      helpers.MakeInputs(inputs),
		VRL: `
if exists(.items) {
    r = array([])
    for_each(array!(.items)) -> |_index, i| {
      r = push(r, {"_internal": {"structured": i}})
    }
    . = r
} else {
  . = {"_internal": {"structured": .}}
}
`,
	}, itemsID
}
