package k8shandler

import (
	"fmt"
	"reflect"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	batchv1 "k8s.io/api/batch/v1"
	batch "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Note: this will eventually be deprecated and functionality will be moved into the ES operator
//   in the case of Curator. Other curation deployments may not be supported in the future

const defaultSchedule = "30 3,9,15,21 * * *"

//CreateOrUpdateCuration reconciles curation for an instance of ClusterLogging
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateCuration() (err error) {

	cluster := clusterRequest.cluster

	if cluster.Spec.Curation.Type == logging.CurationTypeCurator {

		if err = clusterRequest.createOrUpdateCuratorServiceAccount(); err != nil {
			return
		}

		if err = clusterRequest.createOrUpdateCuratorConfigMap(); err != nil {
			return
		}

		if err = clusterRequest.createOrUpdateCuratorCronJob(); err != nil {
			return
		}

		if err = clusterRequest.createOrUpdateCuratorSecret(); err != nil {
			return
		}

		curatorStatus, err := clusterRequest.getCuratorStatus()

		if err != nil {
			return fmt.Errorf("Failed to get status for Curator: %v", err)
		}

		printUpdateMessage := true
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if !reflect.DeepEqual(curatorStatus, cluster.Status.Curation.CuratorStatus) {
				if printUpdateMessage {
					logrus.Info("Updating status of Curator")
					printUpdateMessage = false
				}
				cluster.Status.Curation.CuratorStatus = curatorStatus
				return clusterRequest.Update(cluster)
			}
			return nil
		})
		if retryErr != nil {
			return fmt.Errorf("Failed to update Cluster Logging Curator status: %v", retryErr)
		}
	} else {
		clusterRequest.removeCurator()
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) removeCurator() (err error) {
	if clusterRequest.isManaged() {
		if err = clusterRequest.RemoveServiceAccount("curator"); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap("curator"); err != nil {
			return
		}

		if err = clusterRequest.RemoveSecret("curator"); err != nil {
			return
		}

		if err = clusterRequest.RemoveCronJob("curator"); err != nil {
			return
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateCuratorServiceAccount() error {

	curatorServiceAccount := NewServiceAccount("curator", clusterRequest.cluster.Namespace)
	utils.AddOwnerRefToObject(curatorServiceAccount, utils.AsOwner(clusterRequest.cluster))

	err := clusterRequest.Create(curatorServiceAccount)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Curator service account for %q: %v", clusterRequest.cluster.Name, err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateCuratorConfigMap() error {

	curatorConfigMap := NewConfigMap(
		"curator",
		clusterRequest.cluster.Namespace,
		map[string]string{
			"actions.yaml":  string(utils.GetFileContents(utils.GetShareDir() + "/curator/curator-actions.yaml")),
			"curator5.yaml": string(utils.GetFileContents(utils.GetShareDir() + "/curator/curator5-config.yaml")),
			"config.yaml":   string(utils.GetFileContents(utils.GetShareDir() + "/curator/curator-config.yaml")),
		},
	)

	utils.AddOwnerRefToObject(curatorConfigMap, utils.AsOwner(clusterRequest.cluster))

	err := clusterRequest.Create(curatorConfigMap)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Curator configmap for %q: %v", clusterRequest.cluster.Name, err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateCuratorSecret() error {

	curatorSecret := NewSecret(
		"curator",
		clusterRequest.cluster.Namespace,
		map[string][]byte{
			"ca":       utils.GetWorkingDirFileContents("ca.crt"),
			"key":      utils.GetWorkingDirFileContents("system.logging.curator.key"),
			"cert":     utils.GetWorkingDirFileContents("system.logging.curator.crt"),
			"ops-ca":   utils.GetWorkingDirFileContents("ca.crt"),
			"ops-key":  utils.GetWorkingDirFileContents("system.logging.curator.key"),
			"ops-cert": utils.GetWorkingDirFileContents("system.logging.curator.crt"),
		})

	utils.AddOwnerRefToObject(curatorSecret, utils.AsOwner(clusterRequest.cluster))

	err := clusterRequest.CreateOrUpdateSecret(curatorSecret)
	if err != nil {
		return err
	}

	return nil
}

//TODO: refactor elasticsearchHost to get from cluster
func newCuratorCronJob(cluster *logging.ClusterLogging, curatorName string, elasticsearchHost string) *batch.CronJob {

	var resources = cluster.Spec.Curation.Resources
	if resources == nil {
		resources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceMemory: defaultCuratorMemory},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultCuratorMemory,
				v1.ResourceCPU:    defaultCuratorCpuRequest,
			},
		}
	}
	curatorContainer := NewContainer("curator", "curator", v1.PullIfNotPresent, *resources)

	curatorContainer.Env = []v1.EnvVar{
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc.cluster.local"},
		{Name: "ES_HOST", Value: elasticsearchHost},
		{Name: "ES_PORT", Value: "9200"},
		{Name: "ES_CLIENT_CERT", Value: "/etc/curator/keys/cert"},
		{Name: "ES_CLIENT_KEY", Value: "/etc/curator/keys/key"},
		{Name: "ES_CA", Value: "/etc/curator/keys/ca"},
		{Name: "CURATOR_DEFAULT_DAYS", Value: "30"},
		{Name: "CURATOR_SCRIPT_LOG_LEVEL", Value: "INFO"},
		{Name: "CURATOR_LOG_LEVEL", Value: "ERROR"},
		{Name: "CURATOR_TIMEOUT", Value: "300"},
	}

	curatorContainer.VolumeMounts = []v1.VolumeMount{
		{Name: "certs", ReadOnly: true, MountPath: "/etc/curator/keys"},
		{Name: "config", ReadOnly: true, MountPath: "/etc/curator/settings"},
	}

	curatorPodSpec := NewPodSpec(
		"curator",
		[]v1.Container{curatorContainer},
		[]v1.Volume{
			{Name: "config", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "curator"}}}},
			{Name: "certs", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "curator"}}},
		},
		cluster.Spec.Curation.NodeSelector,
	)

	curatorPodSpec.RestartPolicy = v1.RestartPolicyNever
	curatorPodSpec.TerminationGracePeriodSeconds = utils.GetInt64(600)

	schedule := cluster.Spec.Curation.CuratorSpec.Schedule
	if schedule == "" {
		schedule = defaultSchedule
	}

	curatorCronJob := NewCronJob(
		curatorName,
		cluster.Namespace,
		"curator",
		curatorName,
		batch.CronJobSpec{
			SuccessfulJobsHistoryLimit: utils.GetInt32(1),
			FailedJobsHistoryLimit:     utils.GetInt32(1),
			Schedule:                   schedule,
			JobTemplate: batch.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					BackoffLimit: utils.GetInt32(0),
					Parallelism:  utils.GetInt32(1),
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Name:      curatorName,
							Namespace: cluster.Namespace,
							Labels: map[string]string{
								"provider":      "openshift",
								"component":     curatorName,
								"logging-infra": "curator",
							},
						},
						Spec: curatorPodSpec,
					},
				},
			},
		},
	)

	utils.AddOwnerRefToObject(curatorCronJob, utils.AsOwner(cluster))

	return curatorCronJob
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateCuratorCronJob() (err error) {

	curatorCronJob := newCuratorCronJob(clusterRequest.cluster, "curator", "elasticsearch")

	err = clusterRequest.Create(curatorCronJob)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Curator cronjob: %v", err)
	}

	if clusterRequest.isManaged() {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return clusterRequest.updateCuratorIfRequired(curatorCronJob)
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) updateCuratorIfRequired(desired *batch.CronJob) (err error) {
	current := desired.DeepCopy()

	if err = clusterRequest.Get(desired.Name, current); err != nil {
		if errors.IsNotFound(err) {
			// the object doesn't exist -- it was likely culled
			// recreate it on the next time through if necessary
			return nil
		}
		return fmt.Errorf("Failed to get Curator cronjob: %v", err)
	}

	current, different := isCuratorDifferent(current, desired)

	if different {
		if err = clusterRequest.Update(current); err != nil {
			return err
		}
	}

	return nil
}

func isCuratorDifferent(current *batch.CronJob, desired *batch.CronJob) (*batch.CronJob, bool) {

	different := false

	if !utils.AreSelectorsSame(current.Spec.JobTemplate.Spec.Template.Spec.NodeSelector, desired.Spec.JobTemplate.Spec.Template.Spec.NodeSelector) {
		logrus.Infof("Invalid Curator nodeSelector change found, updating '%s'", current.Name)
		current.Spec.JobTemplate.Spec.Template.Spec.NodeSelector = desired.Spec.JobTemplate.Spec.Template.Spec.NodeSelector
		different = true
	}

	// Check schedule
	if current.Spec.Schedule != desired.Spec.Schedule {
		logrus.Infof("Invalid Curator schedule found, updating %q", current.Name)
		current.Spec.Schedule = desired.Spec.Schedule
		different = true
	}

	// Check suspended
	if current.Spec.Suspend != nil && desired.Spec.Suspend != nil && *current.Spec.Suspend != *desired.Spec.Suspend {
		logrus.Infof("Invalid Curator suspend value found, updating %q", current.Name)
		current.Spec.Suspend = desired.Spec.Suspend
		different = true
	}

	if current.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image != desired.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image {
		logrus.Infof("Curator image change found, updating %q", current.Name)
		current.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image = desired.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image
		different = true
	}

	if utils.AreResourcesDifferent(current, desired) {
		logrus.Infof("Curator resources change found, updating %q", current.Name)
		different = true
	}

	return current, different
}
