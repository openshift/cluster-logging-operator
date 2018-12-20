package k8shandler

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"reflect"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	batchv1 "k8s.io/api/batch/v1"
	batch "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	curationScheduleDefault string = "30 3 * * *"
)

// Note: this will eventually be deprecated and functionality will be moved into the ES operator
//   in the case of Curator. Other curation deployments may not be supported in the future

func (cluster *ClusterLogging) CreateOrUpdateCuration(stack *logging.StackSpec) (err error) {

	if stack.Curation == nil {
		logrus.Debugf("Curation is not spec'd for stack '%s'", stack.Name)
		return
	}

	if err = createOrUpdateCuratorServiceAccount(cluster); err != nil {
		return
	}

	if err = cluster.createOrUpdateCuratorConfigMap(stack); err != nil {
		return
	}

	if err = cluster.createOrUpdateCuratorCronJob(stack); err != nil {
		return
	}

	if err = createOrUpdateCuratorSecret(cluster); err != nil {
		return
	}

	curatorStatus, err := getCuratorStatus(cluster.Namespace)

	if err != nil {
		return fmt.Errorf("Failed to get status for Curator: %v", err)
	}

	printUpdateMessage := true
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if exists := utils.DoesClusterLoggingExist(cluster.ClusterLogging); exists {
			if !reflect.DeepEqual(curatorStatus, cluster.ClusterLogging.Status.Curation.CuratorStatus) {
				if printUpdateMessage {
					logrus.Info("Updating status of Curator")
					printUpdateMessage = false
				}
				cluster.ClusterLogging.Status.Curation.CuratorStatus = curatorStatus
				return sdk.Update(cluster)
			}
		}
		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("Failed to update Cluster Logging Curator status: %v", retryErr)
	}

	return nil
}

func createOrUpdateCuratorServiceAccount(logging *ClusterLogging) error {

	curatorServiceAccount := utils.NewServiceAccount(Curator, logging.Namespace)

	logging.addOwnerRefTo(curatorServiceAccount)

	err := sdk.Create(curatorServiceAccount)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Curator service account: %v", err)
	}

	return nil
}

func (logging *ClusterLogging) createOrUpdateCuratorConfigMap(stack *logging.StackSpec) error {

	name := logging.getCuratorName(stack.Name)
	curatorConfigMap := utils.NewConfigMap(
		name,
		logging.Namespace,
		map[string]string{
			"actions.yaml":  string(utils.GetFileContents("files/curator-actions.yaml")),
			"curator5.yaml": string(utils.GetFileContents("files/curator5-config.yaml")),
			"config.yaml":   string(utils.GetFileContents("files/curator-config.yaml")),
		},
	)

	logging.addOwnerRefTo(curatorConfigMap)

	err := sdk.Create(curatorConfigMap)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Curator configmap '%s': %v", name, err)
	}

	return nil
}

func createOrUpdateCuratorSecret(logging *ClusterLogging) error {

	curatorSecret := utils.NewSecret(
		Curator,
		logging.Namespace,
		map[string][]byte{
			"ca":       utils.GetFileContents("/tmp/_working_dir/ca.crt"),
			"key":      utils.GetFileContents("/tmp/_working_dir/system.logging.curator.key"),
			"cert":     utils.GetFileContents("/tmp/_working_dir/system.logging.curator.crt"),
			"ops-ca":   utils.GetFileContents("/tmp/_working_dir/ca.crt"),
			"ops-key":  utils.GetFileContents("/tmp/_working_dir/system.logging.curator.key"),
			"ops-cert": utils.GetFileContents("/tmp/_working_dir/system.logging.curator.crt"),
		})

	logging.addOwnerRefTo(curatorSecret)

	err := sdk.Create(curatorSecret)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Curator secret: %v", err)
	}

	return nil
}

func (logging *ClusterLogging) newCuratorCronJob(stack *logging.StackSpec) *batch.CronJob {
	curatorName := logging.getCuratorName(stack.Name)
	elasticsearchHost := logging.getElasticsearchName(stack.Name)
	curatorContainer := utils.NewContainer(Curator, v1.PullIfNotPresent, stack.Curation.Resources)

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

	curatorPodSpec := utils.NewPodSpec(
		Curator,
		[]v1.Container{curatorContainer},
		[]v1.Volume{
			{Name: "config", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: curatorName}}}},
			{Name: "certs", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "curator"}}},
		},
	)

	curatorPodSpec.RestartPolicy = v1.RestartPolicyNever
	curatorPodSpec.TerminationGracePeriodSeconds = utils.GetInt64(600)

	schedule := stack.Curation.Schedule
	if len(strings.TrimSpace(schedule)) == 0 {
		schedule = curationScheduleDefault
	}
	curatorCronJob := utils.CronJob(
		curatorName,
		logging.Namespace,
		Curator,
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
							Namespace: logging.Namespace,
							Labels: map[string]string{
								"provider":      "openshift",
								"component":     curatorName,
								"logging-infra": Curator,
							},
						},
						Spec: curatorPodSpec,
					},
				},
			},
		},
	)

	logging.addOwnerRefTo(curatorCronJob)

	return curatorCronJob
}

func (cluster *ClusterLogging) createOrUpdateCuratorCronJob(stack *logging.StackSpec) (err error) {
	curatorCronJob := cluster.newCuratorCronJob(stack)

	err = sdk.Create(curatorCronJob)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Curator cronjob: %v", err)
	}

	if cluster.Spec.ManagementState == logging.ManagementStateManaged {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return updateCuratorIfRequired(curatorCronJob)
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
}

func updateCuratorIfRequired(desired *batch.CronJob) (err error) {
	current := desired.DeepCopy()

	if err = sdk.Get(current); err != nil {
		if errors.IsNotFound(err) {
			// the object doesn't exist -- it was likely culled
			// recreate it on the next time through if necessary
			return nil
		}
		return fmt.Errorf("Failed to get Curator cronjob: %v", err)
	}

	current, different := isCuratorDifferent(current, desired)

	if different {
		if err = sdk.Update(current); err != nil {
			return err
		}
	}

	return nil
}

func isCuratorDifferent(current *batch.CronJob, desired *batch.CronJob) (*batch.CronJob, bool) {

	different := false

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

	return current, different
}
