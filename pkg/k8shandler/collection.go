package k8shandler

import (
	"fmt"
	"reflect"

	"strings"
	"time"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"

	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	apps "k8s.io/api/apps/v1"
)

const (
	clusterLoggingPriorityClassName = "cluster-logging"
)

var (
	retryInterval = time.Second * 30
	timeout       = time.Second * 1800
)

//CreateOrUpdateCollection component of the cluster
func (cluster *ClusterLogging) CreateOrUpdateCollection() (err error) {

	// there is no easier way to check this in golang without writing a helper function
	// TODO: write a helper function to validate Type is a valid option for common setup or tear down
	if cluster.Spec.Collection.Logs.Type == logging.LogCollectionTypeFluentd || cluster.Spec.Collection.Logs.Type == logging.LogCollectionTypeRsyslog {
		if err = createOrUpdateCollectionPriorityClass(cluster); err != nil {
			return
		}

		if err = cluster.createOrUpdateCollectorServiceAccount(); err != nil {
			return
		}
	} else {
		if err = utils.RemoveServiceAccount(cluster.Namespace, "logcollector"); err != nil {
			return
		}

		if err = utils.RemovePriorityClass(clusterLoggingPriorityClassName); err != nil {
			return
		}
	}

	if cluster.Spec.Collection.Logs.Type == logging.LogCollectionTypeFluentd {
		if err = createOrUpdateFluentdService(cluster); err != nil {
			return
		}

		if err = createOrUpdateFluentdServiceMonitor(cluster); err != nil {
			return
		}

		if err = createOrUpdateFluentdPrometheusRule(cluster); err != nil {
			return
		}

		if err = createOrUpdateFluentdConfigMap(cluster); err != nil {
			return
		}

		if err = createOrUpdateFluentdSecret(cluster); err != nil {
			return
		}

		if err = createOrUpdateFluentdDaemonset(cluster); err != nil {
			return
		}

		fluentdStatus, err := getFluentdCollectorStatus(cluster.Namespace)
		if err != nil {
			return fmt.Errorf("Failed to get status of Fluentd: %v", err)
		}

		printUpdateMessage := true
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if exists := cluster.Exists(); exists {
				if !reflect.DeepEqual(fluentdStatus, cluster.Status.Collection.Logs.FluentdStatus) {
					if printUpdateMessage {
						logrus.Info("Updating status of Fluentd")
						printUpdateMessage = false
					}
					cluster.Status.Collection.Logs.FluentdStatus = fluentdStatus
					return sdk.Update(cluster.ClusterLogging)
				}
			}
			return nil
		})
		if retryErr != nil {
			return fmt.Errorf("Failed to update Cluster Logging Fluentd status: %v", retryErr)
		}
	} else {
		removeFluentd(cluster)
	}

	if cluster.Spec.Collection.Logs.Type == logging.LogCollectionTypeRsyslog {
		if err = createOrUpdateRsyslogConfigMap(cluster); err != nil {
			return
		}

		if err = createOrUpdateRsyslogSecret(cluster); err != nil {
			return
		}

		if err = createOrUpdateRsyslogDaemonset(cluster); err != nil {
			return
		}

		rsyslogStatus, err := getRsyslogCollectorStatus(cluster.Namespace)
		if err != nil {
			return fmt.Errorf("Failed to get status of Rsyslog: %v", err)
		}

		printUpdateMessage := true
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if exists := cluster.Exists(); exists {
				if !reflect.DeepEqual(rsyslogStatus, cluster.Status.Collection.Logs.RsyslogStatus) {
					if printUpdateMessage {
						logrus.Info("Updating status of Rsyslog")
						printUpdateMessage = false
					}
					cluster.Status.Collection.Logs.RsyslogStatus = rsyslogStatus
					return sdk.Update(cluster.ClusterLogging)
				}
			}
			return nil
		})
		if retryErr != nil {
			return fmt.Errorf("Failed to update Cluster Logging Rsyslog status: %v", retryErr)
		}
	} else {
		removeRsyslog(cluster)
	}

	return nil
}

func createOrUpdateCollectionPriorityClass(logging *ClusterLogging) error {

	collectionPriorityClass := utils.NewPriorityClass(clusterLoggingPriorityClassName, 1000000, false, "This priority class is for the Cluster-Logging Collector")

	logging.AddOwnerRefTo(collectionPriorityClass)

	err := sdk.Create(collectionPriorityClass)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Collection priority class: %v", err)
	}

	return nil
}

func (cluster *ClusterLogging) createOrUpdateCollectorServiceAccount() error {

	collectorServiceAccount := utils.NewServiceAccount("logcollector", cluster.Namespace)

	utils.AddOwnerRefToObject(collectorServiceAccount, utils.AsOwner(cluster))

	err := sdk.Create(collectorServiceAccount)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Log Collector service account: %v", err)
	}

	// Also create the role and role binding so that the service account has host read access
	collectorRole := utils.NewRole(
		"log-collector-privileged",
		cluster.Namespace,
		utils.NewPolicyRules(
			utils.NewPolicyRule(
				[]string{"security.openshift.io"},
				[]string{"securitycontextconstraints"},
				[]string{"privileged"},
				[]string{"use"},
			),
		),
	)

	utils.AddOwnerRefToObject(collectorRole, utils.AsOwner(cluster))

	err = sdk.Create(collectorRole)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Log collector privileged role: %v", err)
	}

	subject := utils.NewSubject(
		"ServiceAccount",
		"logcollector",
	)
	subject.APIGroup = ""

	collectorRoleBinding := utils.NewRoleBinding(
		"log-collector-privileged-binding",
		cluster.Namespace,
		"log-collector-privileged",
		utils.NewSubjects(
			subject,
		),
	)

	utils.AddOwnerRefToObject(collectorRoleBinding, utils.AsOwner(cluster))

	err = sdk.Create(collectorRoleBinding)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Log collector privileged role binding: %v", err)
	}

	// create clusterrole for logcollector to retrieve metadata
	clusterrules := utils.NewPolicyRules(
		utils.NewPolicyRule(
			[]string{""},
			[]string{"pods", "namespaces"},
			nil,
			[]string{"get", "list", "watch"},
		),
	)
	clusterRole, err := utils.CreateClusterRole("metadata-reader", clusterrules, cluster.ClusterLogging)
	if err != nil {
		return err
	}
	subject = utils.NewSubject(
		"ServiceAccount",
		"logcollector",
	)
	subject.Namespace = cluster.Namespace
	subject.APIGroup = ""

	collectorReaderClusterRoleBinding := utils.NewClusterRoleBinding(
		"cluster-logging-metadata-reader",
		clusterRole.Name,
		utils.NewSubjects(
			subject,
		),
	)

	utils.AddOwnerRefToObject(collectorReaderClusterRoleBinding, utils.AsOwner(cluster))

	err = sdk.Create(collectorReaderClusterRoleBinding)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Log collector %q cluster role binding: %v", collectorReaderClusterRoleBinding.Name, err)
	}

	return nil
}

func isDaemonsetDifferent(current *apps.DaemonSet, desired *apps.DaemonSet) (*apps.DaemonSet, bool) {

	different := false

	if isDaemonsetImageDifference(current, desired) {
		logrus.Infof("Daemonset image change found, updating %q", current.Name)
		current = updateCurrentDaemonsetImages(current, desired)
		different = true
	}

	return current, different
}

func waitForDaemonSetReady(ds *apps.DaemonSet) error {

	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		err = sdk.Get(ds)
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

func isDaemonsetImageDifference(current *apps.DaemonSet, desired *apps.DaemonSet) bool {

	for _, curr := range current.Spec.Template.Spec.Containers {
		for _, des := range desired.Spec.Template.Spec.Containers {
			// Only compare the images of containers with the same name
			if curr.Name == des.Name {
				if curr.Image != des.Image {
					return true
				}
			}
		}
	}

	return false
}

func updateCurrentDaemonsetImages(current *apps.DaemonSet, desired *apps.DaemonSet) *apps.DaemonSet {

	containers := current.Spec.Template.Spec.Containers

	for index, curr := range current.Spec.Template.Spec.Containers {
		for _, des := range desired.Spec.Template.Spec.Containers {
			// Only compare the images of containers with the same name
			if curr.Name == des.Name {
				if curr.Image != des.Image {
					containers[index].Image = des.Image
				}
			}
		}
	}

	return current
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
