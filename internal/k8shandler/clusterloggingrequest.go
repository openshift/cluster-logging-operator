package k8shandler

import (
	"context"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	generatorUtils "github.com/openshift/cluster-logging-operator/internal/generator/utils"
	"github.com/openshift/cluster-logging-operator/internal/k8s/loader"
	"github.com/openshift/cluster-logging-operator/internal/utils"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterLoggingRequest struct {
	Client client.Client
	Reader client.Reader

	ClusterVersion string
	//ClusterID is the unique identifier of the cluster in which the operator is deployed
	ClusterID string

	Cluster       *logging.ClusterLogging
	EventRecorder record.EventRecorder

	// Forwarder is a logforwarder instance
	Forwarder *logging.ClusterLogForwarder

	// OutputSecrets are retrieved during validation and used for generation.
	OutputSecrets map[string]*corev1.Secret

	// Custom resource names for custom CLF
	ResourceNames *factory.ForwarderResourceNames

	// Owner of collector resources
	ResourceOwner metav1.OwnerReference

	// Determine the collector deployment strategy; Daemonset || Deployment
	isDaemonset bool

	// Vector log level
	// Default is `warn`: https://issues.redhat.com/browse/LOG-3435
	LogLevel string
}

func NewClusterLoggingRequest(cl *logging.ClusterLogging, forwarder *logging.ClusterLogForwarder, requestClient client.Client, reader client.Reader, r record.EventRecorder, clusterVersion, clusterID string, resourceNames *factory.ForwarderResourceNames) ClusterLoggingRequest {
	ownerRef := metav1.OwnerReference{}
	if forwarder != nil {
		ownerRef = utils.AsOwner(forwarder)
	}
	request := ClusterLoggingRequest{
		Cluster:        cl,
		Client:         requestClient,
		Reader:         reader,
		EventRecorder:  r,
		Forwarder:      forwarder,
		ClusterVersion: clusterVersion,
		ClusterID:      clusterID,
		ResourceNames:  resourceNames,
		ResourceOwner:  ownerRef,
		isDaemonset:    true,
		LogLevel:       "warn",
	}

	// Owner is always the ClusterLogging instance for legacy ClusterLogging
	if cl.Namespace == constants.OpenshiftNS && cl.Name == constants.SingletonName {
		request.ResourceOwner = utils.AsOwner(cl)
	}

	if request.Forwarder != nil {
		// Collector is not a daemonset if the only input source is an HTTP receiver
		// Enabled through an annotation
		if _, ok := request.Forwarder.Annotations[constants.AnnotationEnableCollectorAsDeployment]; ok {
			inputs := generatorUtils.GatherSources(&forwarder.Spec, framework.NoOptions)
			if inputs.Len() == 1 && inputs.Has(logging.InputNameReceiver) {
				request.isDaemonset = false
			}
		}
	}

	if level, ok := request.Forwarder.Annotations[constants.AnnotationVectorLogLevel]; ok {
		request.LogLevel = level
	}

	return request
}

func (clusterRequest *ClusterLoggingRequest) IncludesManagedStorage() bool {
	return clusterRequest.Cluster != nil && clusterRequest.Cluster.Spec.LogStore != nil
}

func (clusterRequest *ClusterLoggingRequest) IsLegacyDeployment() bool {
	return clusterRequest.Cluster != nil &&
		clusterRequest.Cluster.Namespace == constants.OpenshiftNS &&
		clusterRequest.Cluster.Name == constants.SingletonName
}

// true if equals "Managed" or empty
func (clusterRequest *ClusterLoggingRequest) isManaged() bool {
	return clusterRequest.Cluster.Spec.ManagementState == logging.ManagementStateManaged ||
		clusterRequest.Cluster.Spec.ManagementState == ""
}

// Update the runtime Object or return error
func (clusterRequest *ClusterLoggingRequest) Update(object client.Object) (err error) {
	if err = clusterRequest.Client.Update(context.TODO(), object); err != nil {
		log.Error(err, "Error updating ", object.GetObjectKind())
	}
	return err
}

// UpdateStatus modifies the status sub-resource or returns an error.
func (clusterRequest *ClusterLoggingRequest) UpdateStatus(object client.Object) (err error) {
	if err = clusterRequest.Client.Status().Update(context.TODO(), object); err != nil {
		// making this debug because we should be throwing the returned error if we are never
		// able to update the status
		if strings.Contains(err.Error(), constants.OptimisticLockErrorMsg) {
			// we can skip this error, so rise login level, more info here: https://github.com/kubernetes/kubernetes/issues/28149
			log.V(5).Error(err, "Error updating status")
		} else {
			log.V(2).Error(err, "Error updating status")
		}
	}
	return err
}

func (clusterRequest *ClusterLoggingRequest) Get(objectName string, object client.Object) error {
	namespacedName := types.NamespacedName{Name: objectName, Namespace: clusterRequest.Forwarder.Namespace}

	log.V(3).Info("Getting object", "namespacedName", namespacedName, "object", object)

	return clusterRequest.Client.Get(context.TODO(), namespacedName, object)
}

func (clusterRequest *ClusterLoggingRequest) GetClusterResource(objectName string, object client.Object) error {
	namespacedName := types.NamespacedName{Name: objectName}
	log.V(3).Info("Getting ClusterResource object", "namespacedName", namespacedName, "object", object)
	err := clusterRequest.Client.Get(context.TODO(), namespacedName, object)
	log.V(3).Error(err, "Response")
	return err
}

func (clusterRequest *ClusterLoggingRequest) List(selector map[string]string, object client.ObjectList) error {
	log.V(3).Info("Listing selector object", "selector", selector, "object", object)

	listOpts := []client.ListOption{
		client.InNamespace(clusterRequest.Forwarder.Namespace),
		client.MatchingLabels(selector),
	}

	return clusterRequest.Client.List(
		context.TODO(),
		object,
		listOpts...,
	)
}

func (clusterRequest *ClusterLoggingRequest) Delete(object client.Object) error {
	log.V(3).Info("Deleting", "object", object)
	return clusterRequest.Client.Delete(context.TODO(), object)
}

func (clusterRequest *ClusterLoggingRequest) UpdateCondition(t logging.ConditionType, message string, reason logging.ConditionReason, status corev1.ConditionStatus) error {
	instance, err, _ := loader.FetchClusterLogging(clusterRequest.Client, clusterRequest.Cluster.Namespace, clusterRequest.Cluster.Name, true)
	if err != nil {
		return err
	}

	if logging.SetCondition(&instance.Status.Conditions, t, status, reason, message) {
		return clusterRequest.UpdateStatus(&instance)
	}
	return nil
}
