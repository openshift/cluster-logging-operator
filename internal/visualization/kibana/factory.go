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

func New(namespace, name string, visSpec *logging.VisualizationSpec, logStore *logging.LogStoreSpec, owner metav1.OwnerReference) *es.Kibana {

	resources := &v1.ResourceRequirements{}
	proxyResources := &v1.ResourceRequirements{} //nolint:staticcheck
	nodeSelector := map[string]string{}
	tolerations := []v1.Toleration{}
	replicas := int32(0)
	if visSpec != nil {
		nodeSelector = visSpec.NodeSelector
		tolerations = visSpec.Tolerations
		resources = visSpec.Resources
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

		if visSpec.Replicas != nil {
			replicas = *visSpec.Replicas
		} else {
			if logStore != nil && logStore.Elasticsearch.NodeCount > 0 {
				replicas = 1
			}
		}

		if proxyResources == nil { //nolint:staticcheck
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
