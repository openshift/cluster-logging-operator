package kibana

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	es "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	DefaultKibanaMemory     = resource.MustParse("736Mi")
	DefaultKibanaCpuRequest = resource.MustParse("100m")

	DefaultKibanaProxyMemory     = resource.MustParse("256Mi")
	DefaultKibanaProxyCpuRequest = resource.MustParse("100m")
)

func New(namespace, name string, kibanaSpec *logging.KibanaSpec, logStore *logging.LogStoreSpec, owner metav1.OwnerReference) *es.Kibana {

	resources := &v1.ResourceRequirements{}
	proxyResources := &v1.ResourceRequirements{}
	nodeSelector := map[string]string{}
	tolerations := []v1.Toleration{}
	replicas := int32(0)
	if kibanaSpec != nil {
		nodeSelector = kibanaSpec.NodeSelector
		tolerations = kibanaSpec.Tolerations
		resources = kibanaSpec.Resources
		if resources == nil {
			resources = &v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceMemory: DefaultKibanaMemory,
				},
				Requests: v1.ResourceList{
					v1.ResourceMemory: DefaultKibanaMemory,
					v1.ResourceCPU:    DefaultKibanaCpuRequest,
				},
			}
		}

		if kibanaSpec.Replicas != nil {
			replicas = *kibanaSpec.Replicas
		} else {
			if logStore != nil && logStore.Elasticsearch != nil && logStore.Elasticsearch.NodeCount > 0 {
				replicas = 1
			}
		}

		proxyResources = kibanaSpec.ProxySpec.Resources
		if proxyResources == nil {
			proxyResources = &v1.ResourceRequirements{
				Limits: v1.ResourceList{v1.ResourceMemory: DefaultKibanaProxyMemory},
				Requests: v1.ResourceList{
					v1.ResourceMemory: DefaultKibanaProxyMemory,
					v1.ResourceCPU:    DefaultKibanaProxyCpuRequest,
				},
			}
		}
	}
	cr := &es.Kibana{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Kibana",
			APIVersion: es.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: es.KibanaSpec{
			ManagementState: es.ManagementStateManaged,
			Replicas:        replicas,
			Resources:       resources,
			NodeSelector:    nodeSelector,
			Tolerations:     tolerations,
			ProxySpec: es.ProxySpec{
				Resources: proxyResources,
			},
		},
	}

	utils.AddOwnerRefToObject(cr, owner)
	return cr
}
