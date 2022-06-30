package functional

import (
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/runtime"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	ElasticSearchTag   = "7.10.1"
	ElasticSearchImage = "elasticsearch:" + ElasticSearchTag
)

func (f *CollectorFunctionalFramework) addES7Output(b *runtime.PodBuilder, output logging.OutputSpec) error {
	log.V(2).Info("Adding elasticsearc7 output", "name", output.Name)
	name := strings.ToLower(output.Name)
	log.V(2).Info("Adding container", "name", name)
	log.V(2).Info("Adding ElasticSearch output container", "name", logging.OutputTypeElasticsearch)
	b.AddContainer(logging.OutputTypeElasticsearch, ElasticSearchImage).
		AddEnvVar("discovery.type", "single-node").
		AddRunAsUser(2000).
		WithPullPolicy(corev1.PullIfNotPresent). // Docker image, excessive pulling will exceed quota and fail.
		End()
	return nil
}
