package functional

import (
	"strings"

	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/runtime"
)

const (
	ElasticSearchTag   = "7.10.1"
	ElasticSearchImage = "elasticsearch:" + ElasticSearchTag
)

func (f *FluentdFunctionalFramework) addES7Output(b *runtime.PodBuilder, output logging.OutputSpec) error {

	log.V(2).Info("Adding elasticsearc7 output", "name", output.Name)
	name := strings.ToLower(output.Name)
	log.V(2).Info("Adding container", "name", name)
	log.V(2).Info("Adding ElasticSearch output container", "name", logging.OutputTypeElasticsearch)
	b.AddContainer(logging.OutputTypeElasticsearch, ElasticSearchImage).
		AddEnvVar("discovery.type", "single-node").
		AddRunAsUser(2000).
		End()
	return nil
}
