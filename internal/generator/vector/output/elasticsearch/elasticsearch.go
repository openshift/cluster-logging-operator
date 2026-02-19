package elasticsearch

import (
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/common/tls"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	commontemplate "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

func New(id string, o *observability.Output, inputs []string, secrets observability.Secrets, op utils.Options) []framework.Element {
	if genhelper.IsDebugOutput(op) {
		return []framework.Element{
			elements.Debug(id, helpers.MakeInputs(inputs...)),
		}
	}
	componentID := helpers.MakeID(id, "index")
	outputs := []framework.Element{}
	if o.Elasticsearch.Version == 6 {
		addID := helpers.MakeID(id, "add_id")
		outputs = append(outputs, elements.Remap{
			ComponentID: addID,
			Inputs:      helpers.MakeInputs(inputs...),
			VRL: `._id = encode_base64(uuid_v4())
if exists(.kubernetes.event.metadata.uid) {
  ._id = .kubernetes.event.metadata.uid
}`,
		})
		inputs = []string{addID}
	}

	outputs = append(outputs,
		commontemplate.NewTemplateRemap(componentID, inputs, o.Elasticsearch.Index, componentID, "Elasticsearch Index"),
		api.NewConfig(func(c *api.Config) {
			c.Sinks[id] = sinks.NewElasticsearch(o.Elasticsearch.URL, func(s *sinks.Elasticsearch) {
				s.Bulk = &sinks.Bulk{
					Action: sinks.BulkActionCreate,
					Index:  fmt.Sprintf("{{ _internal.%s }}", componentID),
				}
				s.ApiVersion = fmt.Sprintf("v%d", o.Elasticsearch.Version)
				s.Encoding = common.NewApiEncoding("")
				s.Batch = common.NewApiBatch(o)
				s.Buffer = common.NewApiBuffer(o)
				s.Request = common.NewApiRequest(o)
				if len(o.Elasticsearch.Headers) > 0 {
					if s.Request == nil {
						s.Request = &sinks.Request{}
						s.Request.Headers = o.Elasticsearch.Headers
					}
				}
				elasticsearchAuth(s, o, op)
				if o.Elasticsearch.Version == 6 {
					s.IdKey = "_id"
				}
				s.TLS = tls.NewTls(o, secrets, op)
			}, componentID)
		}),
	)
	return outputs
}

func elasticsearchAuth(s *sinks.Elasticsearch, o *observability.Output, op utils.Options) {
	if o.Elasticsearch.Authentication != nil && o.Elasticsearch.Authentication.Token != nil {
		if s.Request == nil {
			s.Request = &sinks.Request{}
		}
		if s.Request.Headers == nil {
			s.Request.Headers = map[string]string{}
		}
		var token string
		key := o.Elasticsearch.Authentication.Token
		switch o.Elasticsearch.Authentication.Token.From {
		case obs.BearerTokenFromSecret:
			if key.Secret != nil {
				token = helpers.SecretFrom(&obs.SecretReference{
					SecretName: key.Secret.Name,
					Key:        key.Secret.Key,
				})
			}
		case obs.BearerTokenFromServiceAccount:
			if name, found := utils.GetOption(op, framework.OptionServiceAccountTokenSecretName, ""); found {
				token = helpers.SecretFrom(&obs.SecretReference{
					Key:        constants.TokenKey,
					SecretName: name,
				})
			}
		}
		s.Request.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	} else if o.Elasticsearch.Authentication != nil && o.Elasticsearch.Authentication.Username != nil && o.Elasticsearch.Authentication.Password != nil {
		s.Auth = &sinks.ElasticsearchAuth{
			Strategy: sinks.HttpAuthStrategyBasic,
			HttpAuthBasic: sinks.HttpAuthBasic{
				User:     helpers.SecretFrom(o.Elasticsearch.Authentication.Username),
				Password: helpers.SecretFrom(o.Elasticsearch.Authentication.Password),
			},
		}
	}
}
