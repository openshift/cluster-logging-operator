package cloudwatch

import (
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
	corev1 "k8s.io/api/core/v1"
)

type CloudWatch struct {
	ID_Key      string
	Desc        string
	ComponentID string
	Inputs      string
	Endpoint    string
	Region      string
	//if Endpoint provided then it overrides  Region         string
	Host           string
	GroupBy        string
	LogGroupPrefix string
	SecurityConfig Element
	EndpointConfig Element
}

func (e CloudWatch) Name() string {
	return "cloudwatchTemplate"
}

func (e CloudWatch) Template() string {

	return `{{define "` + e.Name() + `" -}}

[sinks.{{.ComponentID}}]
type = "aws_cloudwatch_logs"
inputs = {{.Inputs}}
region = "{{.Region}}"
create_missing_group = true
create_missing_stream = true
compression = "none"
encoding.codec = "json"
request.concurrency = 2
group_name = "{{"{{ Cw_groupName }}"}}-%h-%m"
stream_name = "{{"{{ Cw_streamName }}"}}"
{{- end}}
`
}

func AddCwGroupNameCw_streamName(groupbytype string, logGroupPrefix string, id string, inputs []string) Element {
	return RemapCW{
		Desc:           "Adding group_name and stream_name field",
		ComponentID:    id,
		Inputs:         helpers.MakeInputs(inputs...),
		GroupBy:        groupbytype,
		LogGroupPrefix: logGroupPrefix,
		VRL: strings.TrimSpace(`
.Cw_groupName =  "{{ LogGroupPrefix }}-{{ log_type }}"
.Cw_streamName = "{{ LogGroupPrefix }}-{{ log_type }}-{{ kubernetes.namespace_name }}"
`),
	}
}

func ID(id1, id2 string) string {
	return fmt.Sprintf("%s_%s", id1, id2)
}

func Conf(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) []Element {

	logGroupPrefix := ""
	Region := ""
	if o.Cloudwatch != nil {
		prefix := o.Cloudwatch.GroupPrefix
		if prefix != nil && strings.TrimSpace(*prefix) != "" {
			logGroupPrefix = *prefix
		}
		Region = o.Cloudwatch.Region
		if Region == "" {
			Region = "us-east-1"
		}
	}

	//need three syncs for individual input logs streams - application, infrastructure, audit
	outputs := []Element{}
	outputName := strings.ToLower(vectorhelpers.Replacer.Replace(o.Name))
	groupbytype := string(o.Cloudwatch.GroupBy)

	outputs = MergeElements(outputs,
		[]Element{
			AddCwGroupNameCw_streamName(groupbytype, logGroupPrefix, ID(outputName, "add_grpandstream"), inputs),
			Output(o, []string{ID(outputName, "add_grpandstream")}, secret, op, Region),
		},
		TLSConf(o, secret),
		BasicAuth(o, secret),
	)

	return outputs
}

func Output(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options, region string) Element {

	return CloudWatch{
		ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		Endpoint:    o.URL,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		Region:      region,
	}
}

func TLSConf(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}
	if o.Secret != nil {
		hasTLS := false
		if o.Name == logging.OutputNameDefault || security.HasAwsCredentials(secret) {
			hasTLS = true
			kc := AWSKey{
				AWSAccessKeyID:     security.GetFromSecret(secret, "aws_access_key_id"),
				AWSSecretAccessKey: security.GetFromSecret(secret, "aws_secret_access_key"),
			}
			conf = append(conf, kc)
		}
		if !hasTLS {
			return []Element{}
		}
	}
	return conf
}

//not sure if required for cloudwatch?
func BasicAuth(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}

	if o.Secret != nil {
		hasBasicAuth := false
		conf = append(conf, BasicAuthConf{
			Desc:        "Basic Auth Config",
			ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		})
		if !hasBasicAuth {
			return []Element{}
		}
	}

	return conf
}
