package helpers

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	receiverses "github.com/openshift/cluster-logging-operator/test/framework/e2e/receivers/elasticsearch"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/ViaQ/logerr/v2/log/static"
	cl "github.com/openshift/cluster-logging-operator/api/logging/v1"
	elasticsearch "github.com/openshift/elasticsearch-operator/apis/logging/v1"
)

type LogComponentType string

const (
	ComponentTypeStore                          LogComponentType = "LogStore"
	ComponentTypeVisualization                  LogComponentType = "Visualization"
	ComponentTypeCollector                      LogComponentType = "collector"
	ComponentTypeCollectorFluentd               LogComponentType = "collector-fluentd"
	ComponentTypeCollectorVector                LogComponentType = "collector-vector"
	ComponentTypeCollectorDeployment            LogComponentType = "collector-deployment"
	ComponentTypeReceiverElasticsearchRHManaged LogComponentType = receiverses.ManagedLogStore
)

func NewClusterLogging(componentTypes ...LogComponentType) *cl.ClusterLogging {
	log.Info("NewClusterLogging ", "componentTypes", componentTypes)
	instance := &cl.ClusterLogging{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterLogging",
			APIVersion: cl.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.SingletonName,
			Namespace: constants.OpenshiftNS,
		},
		Spec: cl.ClusterLoggingSpec{
			ManagementState: cl.ManagementStateManaged,
		},
	}

	for _, compType := range componentTypes {
		switch compType {
		case ComponentTypeStore:
			instance.Spec.LogStore = &cl.LogStoreSpec{
				Type: cl.LogStoreTypeElasticsearch,
				Elasticsearch: &cl.ElasticsearchSpec{
					Resources: &v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("2Gi"),
							v1.ResourceCPU:    resource.MustParse("100m"),
						},
					},
					ProxySpec: cl.ProxySpec{
						Resources: &v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("512Mi"),
								v1.ResourceCPU:    resource.MustParse("200m"),
							},
						},
					},
					NodeCount:        1,
					RedundancyPolicy: elasticsearch.ZeroRedundancy,
				},
			}
		case ComponentTypeCollector, ComponentTypeCollectorFluentd:
			instance.Spec.Collection = &cl.CollectionSpec{
				Type: cl.LogCollectionTypeFluentd,
				CollectorSpec: cl.CollectorSpec{
					Resources: &v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("500Mi"),
							v1.ResourceCPU:    resource.MustParse("200m"),
						},
					},
				},
			}
		case ComponentTypeCollectorVector:
			instance.Spec.Collection = &cl.CollectionSpec{
				Type: cl.LogCollectionTypeVector,
			}

		}
	}
	return instance
}
