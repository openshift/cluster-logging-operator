package elasticsearch

import (
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
	corev1 "k8s.io/api/core/v1"
)

type Elasticsearch struct {
	ID_Key      string
	Desc        string
	ComponentID string
	Inputs      string
	Index       string
	Endpoint    string
}

func (e Elasticsearch) Name() string {
	return "elasticsearchTemplate"
}

func (e Elasticsearch) Template() string {
	return `{{define "` + e.Name() + `" -}}
[sinks.{{.ComponentID}}]
type = "elasticsearch"
inputs = {{.Inputs}}
endpoint = "{{.Endpoint}}"
bulk.index = "{{ "{{ write_index }}" }}"
bulk.action = "create"
request.timeout_secs = 2147483648
id_key = "_id"
{{end}}`
}

func SetESIndex(id string, inputs []string, o logging.OutputSpec, op Options) Element {

	setESIndex := `
index = "default"
if (.log_type == "application"){
  index = "app"
}
if (.log_type == "infrastructure"){
  index = "infra"
}
if (.log_type == "audit"){
  index = "audit"
}
.write_index = index + "-write"`

	addESId := `._id = encode_base64(uuid_v4())`
	removeFile := `del(.file)`
	removeTag := `del(.tag)`
	removeSourceType := `del(.source_type)`

	vrls := []string{setESIndex, addESId, removeFile, removeTag, removeSourceType}

	if o.Elasticsearch != nil {
		es := o.Elasticsearch
		switch {
		case es.StructuredTypeKey != "" && es.StructuredTypeName == "":
			changeIndexName := `
if .log_type == "application" && .structured != null {
  val = .%s
  if val != null {
    .write_index, err = "app-" + val + "-write"
  }
}  
`
			vrls = append(vrls, fmt.Sprintf(changeIndexName, o.Elasticsearch.StructuredTypeKey))
		case es.StructuredTypeKey == "" && es.StructuredTypeName != "":
			changeIndexName := `
if .log_type == "application" && .structured != null {
  .write_index = "app-%s-write"
}
`
			vrls = append(vrls, fmt.Sprintf(changeIndexName, o.Elasticsearch.StructuredTypeName))
		case es.StructuredTypeKey != "" && es.StructuredTypeName != "":
			changeIndexName := `
if .log_type == "application" && .structured != null {
  val = .%s
  if val != null {
    .write_index, err = "app-" + val + "-write"
  } else {
    .write_index = "app-%s-write"
  }
}
`
			vrls = append(vrls, fmt.Sprintf(changeIndexName, o.Elasticsearch.StructuredTypeKey, o.Elasticsearch.StructuredTypeName))
		}
		if es.EnableStructuredContainerLogs {
			vrls = append(vrls, `
  if .log_type == "application"  && .structured != null && .kubernetes.container_name != null && .kubernetes.annotations != null && length!(.kubernetes.annotations) > 0{
	key = join!(["containerType.logging.openshift.io", .kubernetes.container_name], separator: "/")
    index, err = get(value: .kubernetes.annotations, path: [key])
    if index != null && err == null {
      .write_index = join!(["app-",index,"-write"])
    } else {
       log(err, level: "error")
    }
  }
`)
		}
	}

	return Remap{
		Desc:        "Set Elasticsearch index",
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         strings.Join(helpers.TrimSpaces(vrls), "\n"),
	}
}

func DeDotAndFlattenLabels(id string, inputs []string) Element {
	return ConfLiteral{
		ComponentID:  id,
		InLabel:      helpers.MakeInputs(inputs...),
		TemplateName: "dedotTemplate",
		TemplateStr: `{{define "dedotTemplate" -}}
[transforms.{{.ComponentID}}]
type = "lua"
inputs = {{.InLabel}}
version = "2"
hooks.process = "process"
source = '''
    function process(event, emit)
        if event.log.kubernetes == nil then
            emit(event)
            return
        end
        if event.log.kubernetes.labels == nil then
            emit(event)
            return
        end
        dedot(event.log.kubernetes.labels)
        flatten_labels(event)
        prune_labels(event)
        emit(event)
    end

    function flatten_labels(event)
        -- create "flat_labels" key
        event.log.kubernetes.flat_labels = {}
        i = 1
        -- flatten the labels
        for k,v in pairs(event.log.kubernetes.labels) do
          event.log.kubernetes.flat_labels[i] = k.."="..v
          i=i+1
        end
    end 

	function prune_labels(event)
		local exclusions = {"app_kubernetes_io/name", "app_kubernetes_io/instance", "app_kubernetes_io/version", "app_kubernetes_io/component", "app_kubernetes_io/part-of", "app_kubernetes_io/managed-by", "app_kubernetes_io/created-by"}
		local keys = {}
		for k,v in pairs(event.log.kubernetes.labels) do
			for index, e in pairs(exclusions) do
				if k == e then
					keys[k] = v
				end
			end
		end
		event.log.kubernetes.labels = keys
	end

    function dedot(map)
        if map == nil then
            return
        end
        local new_map = {}
        local changed_keys = {}
        for k, v in pairs(map) do
            local dedotted = string.gsub(k, "%.", "_")
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

func ID(id1, id2 string) string {
	return fmt.Sprintf("%s_%s", id1, id2)
}

func Conf(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) []Element {
	outputs := []Element{}
	outputName := helpers.FormatComponentID(o.Name)
	if genhelper.IsDebugOutput(op) {
		return []Element{
			SetESIndex(ID(outputName, "add_es_index"), inputs, o, op),
			DeDotAndFlattenLabels(ID(outputName, "dedot_and_flatten"), []string{ID(outputName, "add_es_index")}),
			Debug(outputName, helpers.MakeInputs([]string{ID(outputName, "dedot_and_flatten")}...)),
		}
	}
	outputs = MergeElements(outputs,
		[]Element{
			SetESIndex(ID(outputName, "add_es_index"), inputs, o, op),
			DeDotAndFlattenLabels(ID(outputName, "dedot_and_flatten"), []string{ID(outputName, "add_es_index")}),
			Output(o, []string{ID(outputName, "dedot_and_flatten")}, secret, op),
		},
		TLSConf(o, secret),
		BasicAuth(o, secret),
	)

	return outputs
}

func Output(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) Element {

	return Elasticsearch{
		ComponentID: helpers.FormatComponentID(o.Name),
		Endpoint:    o.URL,
		Inputs:      helpers.MakeInputs(inputs...),
	}
}

func TLSConf(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}
	if o.Secret != nil {
		hasTLS := false
		conf = append(conf, security.TLSConf{
			Desc:        "TLS Config",
			ComponentID: helpers.FormatComponentID(o.Name),
		})
		if o.Name == logging.OutputNameDefault || security.HasTLSCertAndKey(secret) {
			hasTLS = true
			kc := TLSKeyCert{
				CertPath: security.SecretPath(o.Secret.Name, constants.ClientCertKey),
				KeyPath:  security.SecretPath(o.Secret.Name, constants.ClientPrivateKey),
			}
			conf = append(conf, kc)
		}
		if o.Name == logging.OutputNameDefault || security.HasCABundle(secret) {
			hasTLS = true
			ca := CAFile{
				CAFilePath: security.SecretPath(o.Secret.Name, constants.TrustedCABundleKey),
			}
			conf = append(conf, ca)
		}
		if !hasTLS {
			return []Element{}
		}
	}
	return conf
}

func BasicAuth(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}

	if o.Secret != nil {
		hasBasicAuth := false
		conf = append(conf, BasicAuthConf{
			Desc:        "Basic Auth Config",
			ComponentID: helpers.FormatComponentID(o.Name),
		})
		if security.HasUsernamePassword(secret) {
			hasBasicAuth = true
			up := UserNamePass{
				Username: security.GetFromSecret(secret, constants.ClientUsername),
				Password: security.GetFromSecret(secret, constants.ClientPassword),
			}
			conf = append(conf, up)
		}
		if !hasBasicAuth {
			return []Element{}
		}
	}

	return conf
}
