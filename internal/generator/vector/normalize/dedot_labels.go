package normalize

import (
	"fmt"

	. "github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func ID(id1, id2 string) string {
	return fmt.Sprintf("%s_%s", id1, id2)
}

// Dedotting namespace labels
// Replaces '.' and '/' with '_'
func DedotLabels(id string, inputs []string) Element {
	return ConfLiteral{
		ComponentID:  id,
		InLabel:      helpers.MakeInputs(inputs...),
		TemplateName: "dedotTemplate",
		TemplateStr: `{{define "dedotTemplate" -}}
[transforms.{{.ComponentID}}]
type = "lua"
inputs = {{.InLabel}}
version = "2"
hooks.init = "init"
hooks.process = "process"
source = '''
    function init()
        count = 0
    end
    function process(event, emit)
        count = count + 1
        event.log.openshift.sequence = count
        if event.log.kubernetes == nil then
            emit(event)
            return
        end
        if event.log.kubernetes.labels == nil then
            emit(event)
            return
        end
		dedot(event.log.kubernetes.namespace_labels)
        dedot(event.log.kubernetes.labels)
        emit(event)
    end
	
    function dedot(map)
        if map == nil then
            return
        end
        local new_map = {}
        local changed_keys = {}
        for k, v in pairs(map) do
            local dedotted = string.gsub(k, "[./]", "_")
            if dedotted ~= k then
                new_map[dedotted] = v
                changed_keys[k] = true
            end
        end
        for k in pairs(changed_keys) do
            map[k] = nil
        end
        for k, v in pairs(new_map) do
            map[k] = v
        end
    end
'''
{{end}}`,
	}
}
