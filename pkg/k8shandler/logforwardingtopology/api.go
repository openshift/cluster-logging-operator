package logforwardingtopology

import (
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-logging-operator/pkg/status"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	//LogForwardingTopologyAnnotation describes the topology to use for log collection.  The absence of the
	//annotation or a recognized value results in the default topology
	LogForwardingTopologyAnnotation = "clusterlogging.openshift.io/logforwardingTopology"

	//EnableDevPreviewLogForwardingTopologyAnnotation setting the value to 'true' will cause the operator to evalute
	//the topology to be used in log forwarding
	EnableDevPreviewLogForwardingTopologyAnnotation = "clusterlogging.openshift.io/enableDevPreviewTopology"

	//LogForwardingDualEdgeNormalizationTopology deploys multiple containers to each node to collect and normalize log messagges
	LogForwardingDualEdgeNormalizationTopology = "dualEdgeNormalization"

	//LogForwardingEdgeNormalizationTopology is the default (legacy) topology to deploy a single container to each node to collect and normalize log messages.
	LogForwardingEdgeNormalizationTopology = "edgeNormalization"

	//LogForwardingCentralNormalizationTopology deploys a single container to each node to collect and forward messages
	//to a centralized log normalizer
	LogForwardingCentralNormalizationTopology                      = "centralNormalization"
	LogForwardingTopologyCondition            status.ConditionType = "LogForwardingTopology"
)

type APIClient interface {
	Create(object runtime.Object) error
	Update(object runtime.Object) (err error)
	Get(objectName string, object runtime.Object) error
	Delete(object runtime.Object) error
}

type LogForwarderTopology interface {
	Reconcile(proxyConfig *configv1.Proxy) error
	Undeploy() error
}

func NewLogForwardingTopologyCondition(topology string) status.Condition {
	return status.Condition{
		Type:               LogForwardingTopologyCondition,
		Status:             core.ConditionTrue,
		Reason:             status.ConditionReason(topology),
		Message:            "This is the enabled collector and normalization topology",
		LastTransitionTime: meta.Now(),
	}
}
