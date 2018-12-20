package k8shandler

import (
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	APP   string = "app"
	INFRA string = "infra"

	PriorityClassName string = "cluster-logging"
	Curator           string = "curator"
	Fluentd           string = "fluentd"
	Kibana            string = "kibana"
	KibanaProxy       string = "kibana-proxy"
	Elasticsearch     string = "elasticsearch"
	Rsyslog           string = "rsyslog"
)

type ClusterLogging struct {
	*logging.ClusterLogging
}

func NewClusterLogging(logging *logging.ClusterLogging) *ClusterLogging {
	local := &ClusterLogging{logging}
	return local
}

func (logging *ClusterLogging) isSingleStack() bool {
	return len(logging.Spec.Stacks) == 1
}

func (logging *ClusterLogging) getStackNames() *sets.String {
	names := sets.NewString()
	for _, cluster := range logging.Spec.Stacks {
		names.Insert(cluster.Name)
	}
	return &names
}
func (logging *ClusterLogging) addOwnerRefTo(object metav1.Object) {
	ownerRef := utils.AsOwner(logging.ClusterLogging)
	utils.AddOwnerRefToObject(object, ownerRef)
}

func (logging *ClusterLogging) getCuratorName(name string) string {
	if logging.isSingleStack() {
		return Curator
	}
	return Curator + "-" + name
}

func (logging *ClusterLogging) getElasticsearchName(name string) string {
	if logging.isSingleStack() {
		return Elasticsearch
	}
	return Elasticsearch + "-" + name
}

func (logging *ClusterLogging) getKibanaName(name string) string {
	if logging.isSingleStack() {
		return Kibana
	}
	return Kibana + "-" + name
}
