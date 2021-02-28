package k8shandler

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	clusterLoggingPriorityClassName = "cluster-logging"
	metricsPort                     = int32(24231)
	metricsPortName                 = "metrics"
	metricsVolumeName               = "collector-metrics"
	prometheusCAFile                = "/etc/prometheus/configmaps/serving-certs-ca-bundle/service-ca.crt"
)

var (
	retryInterval = time.Second * 30
	timeout       = time.Second * 1800
)

var serviceAccountLogCollectorUID types.UID

//CreateOrUpdateCollection component of the cluster
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateCollection() (err error) {
	cluster := clusterRequest.Cluster
	collectorConfig := ""
	collectorConfHash := ""

	var collectorServiceAccount *core.ServiceAccount

	// there is no easier way to check this in golang without writing a helper function
	// TODO: write a helper function to validate Type is a valid option for common setup or tear down
	if cluster.Spec.Collection != nil && cluster.Spec.Collection.Logs.Type == logging.LogCollectionTypeFluentd {
		if err = clusterRequest.createOrUpdateCollectionPriorityClass(); err != nil {
			return
		}

		if collectorServiceAccount, err = clusterRequest.createOrUpdateCollectorServiceAccount(); err != nil {
			return
		}

		if collectorConfig, err = clusterRequest.generateCollectorConfig(); err != nil {
			return
		}
		log.V(3).Info("Generated collector config", "config", collectorConfig)
		collectorConfHash, err = utils.CalculateMD5Hash(collectorConfig)
		if err != nil {
			log.Error(err, "unable to calculate MD5 hash")
			return
		}
		if err = clusterRequest.reconcileFluentdService(); err != nil {
			return
		}

		if err = clusterRequest.reconcileFluentdServiceMonitor(); err != nil {
			return
		}

		if err = clusterRequest.createOrUpdateFluentdPrometheusRule(); err != nil {
			log.Error(err, "unable to create or update fluentd prometheus rule")
		}

		if err = clusterRequest.createOrUpdateFluentdConfigMap(collectorConfig); err != nil {
			return
		}

		if err = clusterRequest.createOrUpdateFluentdSecret(); err != nil {
			return
		}

		if err = clusterRequest.createOrUpdateFluentdDaemonset(collectorConfHash); err != nil {
			return
		}

		if err = clusterRequest.UpdateFluentdStatus(); err != nil {
			log.Error(err, "unable to update status for fluentd")
		}

		if collectorServiceAccount != nil {

			// remove our finalizer from the list and update it.
			collectorServiceAccount.ObjectMeta.Finalizers = utils.RemoveString(collectorServiceAccount.ObjectMeta.Finalizers, metav1.FinalizerDeleteDependents)
			if err = clusterRequest.Update(collectorServiceAccount); err != nil {
				log.Info("Unable to update the collector serviceaccount finalizers", "collectorServiceAccount.Name", collectorServiceAccount.Name)
				return nil
			}
		}
	} else {
		if err = clusterRequest.RemoveServiceAccount("logcollector"); err != nil {
			return
		}

		if err = clusterRequest.RemovePriorityClass(clusterLoggingPriorityClassName); err != nil {
			return
		}

		if err = clusterRequest.removeFluentd(); err != nil {
			return
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) UpdateFluentdStatus() (err error) {

	cluster := clusterRequest.Cluster

	fluentdStatus, err := clusterRequest.getFluentdCollectorStatus()
	if err != nil {
		return fmt.Errorf("Failed to get status of Fluentd: %v", err)
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if !compareFluentdCollectorStatus(fluentdStatus, cluster.Status.Collection.Logs.FluentdStatus) {
			cluster.Status.Collection.Logs.FluentdStatus = fluentdStatus
			return clusterRequest.UpdateStatus(cluster)
		}
		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("Failed to update Cluster Logging Fluentd status: %v", retryErr)
	}

	return nil
}

func compareFluentdCollectorStatus(lhs, rhs logging.FluentdCollectorStatus) bool {
	if lhs.DaemonSet != rhs.DaemonSet {
		return false
	}

	if len(lhs.Conditions) != len(rhs.Conditions) {
		return false
	}

	if len(lhs.Conditions) > 0 {
		if !reflect.DeepEqual(lhs.Conditions, rhs.Conditions) {
			return false
		}
	}

	if len(lhs.Nodes) != len(rhs.Nodes) {
		return false
	}

	if len(lhs.Nodes) > 0 {
		if !reflect.DeepEqual(lhs.Nodes, rhs.Nodes) {

			return false
		}
	}

	if len(lhs.Pods) != len(rhs.Pods) {
		return false
	}

	if len(lhs.Pods) > 0 {
		if !reflect.DeepEqual(lhs.Pods, rhs.Pods) {
			return false
		}
	}

	return true
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateCollectionPriorityClass() error {

	collectionPriorityClass := NewPriorityClass(clusterLoggingPriorityClassName, 1000000, false, "This priority class is for the Cluster-Logging Collector")

	utils.AddOwnerRefToObject(collectionPriorityClass, utils.AsOwner(clusterRequest.Cluster))

	err := clusterRequest.Create(collectionPriorityClass)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Collection priority class: %v", err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateCollectorServiceAccount() (*core.ServiceAccount, error) {

	cluster := clusterRequest.Cluster

	collectorServiceAccount := NewServiceAccount("logcollector", cluster.Namespace)

	utils.AddOwnerRefToObject(collectorServiceAccount, utils.AsOwner(clusterRequest.Cluster))

	delfinalizer := false
	if collectorServiceAccount.ObjectMeta.DeletionTimestamp.IsZero() {
		// This object is not being deleted.
		if !utils.ContainsString(collectorServiceAccount.ObjectMeta.Finalizers, metav1.FinalizerDeleteDependents) {
			collectorServiceAccount.ObjectMeta.Finalizers = append(collectorServiceAccount.ObjectMeta.Finalizers, metav1.FinalizerDeleteDependents)
		}
		err := clusterRequest.Create(collectorServiceAccount)
		if err != nil && !errors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("Failure creating Log Collector service account: %v", err)
		}
		if len(collectorServiceAccount.ObjectMeta.UID) != 0 {
			serviceAccountLogCollectorUID = collectorServiceAccount.ObjectMeta.UID
		}
	} else if utils.ContainsString(collectorServiceAccount.ObjectMeta.Finalizers, metav1.FinalizerDeleteDependents) {
		// This object is being deleted.
		// our finalizer is present, so lets handle any dependency
		delfinalizer = true
	}

	// Also create the role and role binding so that the service account has host read access
	collectorRole := NewRole(
		"log-collector-privileged",
		cluster.Namespace,
		NewPolicyRules(
			NewPolicyRule(
				[]string{"security.openshift.io"},
				[]string{"securitycontextconstraints"},
				[]string{"privileged"},
				[]string{"use"},
			),
		),
	)

	utils.AddOwnerRefToObject(collectorRole, utils.AsOwner(cluster))

	err := clusterRequest.Create(collectorRole)
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("Failure creating Log collector privileged role: %v", err)
	}

	subject := NewSubject(
		"ServiceAccount",
		"logcollector",
	)
	subject.APIGroup = ""

	collectorRoleBinding := NewRoleBinding(
		"log-collector-privileged-binding",
		cluster.Namespace,
		"log-collector-privileged",
		NewSubjects(
			subject,
		),
	)

	utils.AddOwnerRefToObject(collectorRoleBinding, utils.AsOwner(cluster))

	err = clusterRequest.Create(collectorRoleBinding)
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("Failure creating Log collector privileged role binding: %v", err)
	}

	// create clusterrole for logcollector to retrieve metadata
	clusterrules := NewPolicyRules(
		NewPolicyRule(
			[]string{""},
			[]string{"pods", "namespaces"},
			nil,
			[]string{"get", "list", "watch"},
		),
	)
	clusterRole, err := clusterRequest.CreateClusterRole("metadata-reader", clusterrules, cluster)
	if err != nil {
		return nil, err
	}
	subject = NewSubject(
		"ServiceAccount",
		"logcollector",
	)
	subject.Namespace = cluster.Namespace
	subject.APIGroup = ""

	collectorReaderClusterRoleBinding := NewClusterRoleBinding(
		"cluster-logging-metadata-reader",
		clusterRole.Name,
		NewSubjects(
			subject,
		),
	)

	err = clusterRequest.Create(collectorReaderClusterRoleBinding)
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("Failure creating Log collector %q cluster role binding: %v", collectorReaderClusterRoleBinding.Name, err)
	}

	if delfinalizer {
		return collectorServiceAccount, nil
	} else {
		return nil, nil
	}
}

func (clusterRequest *ClusterLoggingRequest) waitForDaemonSetReady(ds *apps.DaemonSet) error {

	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		err = clusterRequest.Get(ds.Name, ds)
		if err != nil {
			if errors.IsNotFound(err) {
				return false, fmt.Errorf("Failed to get Fluentd daemonset: %v", err)
			}
			return false, err
		}

		if int(ds.Status.DesiredNumberScheduled) == int(ds.Status.NumberReady) {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return err
	}

	return nil
}

func isBufferFlushRequired(current *apps.DaemonSet, desired *apps.DaemonSet) bool {

	currImage := strings.Split(current.Spec.Template.Spec.Containers[0].Image, ":")
	desImage := strings.Split(desired.Spec.Template.Spec.Containers[0].Image, ":")

	if len(currImage) != 2 || len(desImage) != 2 {
		// we don't have versions here -- not sure how we would compare versions to determine
		// need to flush buffers
		return false
	}

	currVersion := currImage[1]
	desVersion := desImage[1]

	if strings.HasPrefix(currVersion, "v") {
		currVersion = strings.Split(currVersion, "v")[1]
	}

	if strings.HasPrefix(desVersion, "v") {
		desVersion = strings.Split(desVersion, "v")[1]
	}

	return (currVersion == "3.11" && desVersion == "4.0.0")
}

func getServiceAccountLogCollectorUID() types.UID {
	return serviceAccountLogCollectorUID
}
