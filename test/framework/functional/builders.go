package functional

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/client"
	corev1 "k8s.io/api/core/v1"
)

type CollectorFunctionalFrameworkBuilder struct {
	builder     *ClusterLogForwarderBuilder
	Framework   *CollectorFunctionalFramework
	podVisitors []runtime.PodBuilderVisitor
}

type FrameworkPipelineBuilder struct {
	frameworkBuilder *CollectorFunctionalFrameworkBuilder
	pipelineBuilder  *PipelineBuilder
}

func NewCollectorFunctionalFrameworkUsingCollectorBuilder(logCollectorType logging.LogCollectionType, testOptions ...client.TestOption) *CollectorFunctionalFrameworkBuilder {
	framework := NewCollectorFunctionalFrameworkUsingCollector(logCollectorType, testOptions...)

	return &CollectorFunctionalFrameworkBuilder{
		builder:     NewClusterLogForwarderBuilder(framework.Forwarder),
		Framework:   framework,
		podVisitors: []runtime.PodBuilderVisitor{},
	}
}

func (b *CollectorFunctionalFrameworkBuilder) FromInput(inputName string) *FrameworkPipelineBuilder {
	return &FrameworkPipelineBuilder{
		frameworkBuilder: b,
		pipelineBuilder:  b.builder.FromInput(inputName),
	}
}

func (b *CollectorFunctionalFrameworkBuilder) Deploy() error {
	return b.Framework.DeployWithVisitors(b.podVisitors)
}

type ElasticsearchVersion int

const (
	ElasticsearchVersion6 ElasticsearchVersion = 6
	ElasticsearchVersion7 ElasticsearchVersion = 7
	ElasticsearchVersion8 ElasticsearchVersion = 8
)

func (p *FrameworkPipelineBuilder) ToElasticSearchOutput(version ElasticsearchVersion, secret *corev1.Secret) *CollectorFunctionalFrameworkBuilder {
	p.pipelineBuilder.ToOutputWithVisitor(func(output *logging.OutputSpec) {
		if version > ElasticsearchVersion6 {
			output.Elasticsearch.Version = int(version)
		}
		if secret != nil {
			p.frameworkBuilder.Framework.Secrets = append(p.frameworkBuilder.Framework.Secrets, secret)
			output.Secret = &logging.OutputSecretSpec{
				Name: secret.Name,
			}
		}
		deploy := func(b *runtime.PodBuilder) error {
			return AddESOutput(version, b, *output)
		}
		p.frameworkBuilder.podVisitors = append(p.frameworkBuilder.podVisitors, deploy)
	}, logging.OutputTypeElasticsearch)

	return p.frameworkBuilder
}
