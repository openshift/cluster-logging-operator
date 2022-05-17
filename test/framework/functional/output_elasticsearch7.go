package functional

import (
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/runtime"

	"github.com/ViaQ/logerr/v2/log"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

const (
	ElasticSearchTag   = "7.10.1"
	ElasticSearchImage = "elasticsearch:" + ElasticSearchTag
)

func (f *CollectorFunctionalFramework) addES7Output(b *runtime.PodBuilder, output logging.OutputSpec) error {

	logger := log.NewLogger("output-elasticsearch7-testing")

	logger.V(2).Info("Adding elasticsearc7 output", "name", output.Name)
	name := strings.ToLower(output.Name)
	logger.V(2).Info("Adding container", "name", name)
	logger.V(2).Info("Adding ElasticSearch output container", "name", logging.OutputTypeElasticsearch)
	b.AddContainer(logging.OutputTypeElasticsearch, ElasticSearchImage).
		AddEnvVar("discovery.type", "single-node").
		AddRunAsUser(2000).
		End()
	return nil
}
