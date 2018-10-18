package k8shandler

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"reflect"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	batchv1 "k8s.io/api/batch/v1"
	batch "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Note: this will eventually be deprecated and functionality will be moved into the ES operator
//   in the case of Curator. Other curation deployments may not be supported in the future

func CreateOrUpdateCuration(cluster *logging.ClusterLogging) (err error) {

	if cluster.Spec.Curation.Type == logging.CurationTypeCurator {

		if err = createOrUpdateCuratorServiceAccount(cluster); err != nil {
			return
		}

		if err = createOrUpdateCuratorConfigMap(cluster); err != nil {
			return
		}

		if err = createOrUpdateCuratorCronJob(cluster); err != nil {
			return
		}

		if err = createOrUpdateCuratorSecret(cluster); err != nil {
			return
		}

		curatorStatus, err := getCuratorStatus(cluster.Namespace)

		if err != nil {
			return fmt.Errorf("Failed to get status for Curator: %v", err)
		}

		if !reflect.DeepEqual(curatorStatus, cluster.Status.Curation.CuratorStatus) {
			logrus.Infof("Updating status of Curator")
			cluster.Status.Curation.CuratorStatus = curatorStatus

			if err = sdk.Update(cluster); err != nil {
				return fmt.Errorf("Failed to update Cluster Logging Curator status: %v", err)
			}
		}
	}

	return nil
}

func createOrUpdateCuratorServiceAccount(logging *logging.ClusterLogging) error {

	curatorServiceAccount := utils.ServiceAccount("curator", logging.Namespace)

	utils.AddOwnerRefToObject(curatorServiceAccount, utils.AsOwner(logging))

	err := sdk.Create(curatorServiceAccount)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Curator service account: %v", err)
	}

	return nil
}

func createOrUpdateCuratorConfigMap(logging *logging.ClusterLogging) error {

	curatorConfigMap := utils.ConfigMap(
		"curator",
		logging.Namespace,
		map[string]string{
			"actions.yaml":  string(utils.GetFileContents("files/curator-actions.yaml")),
			"curator5.yaml": string(utils.GetFileContents("files/curator5-config.yaml")),
			"config.yaml":   string(utils.GetFileContents("files/curator-config.yaml")),
		},
	)

	utils.AddOwnerRefToObject(curatorConfigMap, utils.AsOwner(logging))

	err := sdk.Create(curatorConfigMap)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Curator configmap: %v", err)
	}

	return nil
}

func createOrUpdateCuratorSecret(logging *logging.ClusterLogging) error {

	curatorSecret := utils.Secret(
		"curator",
		logging.Namespace,
		map[string][]byte{
			"ca":       utils.GetFileContents("/tmp/_working_dir/ca.crt"),
			"key":      utils.GetFileContents("/tmp/_working_dir/system.logging.curator.key"),
			"cert":     utils.GetFileContents("/tmp/_working_dir/system.logging.curator.crt"),
			"ops-ca":   utils.GetFileContents("/tmp/_working_dir/ca.crt"),
			"ops-key":  utils.GetFileContents("/tmp/_working_dir/system.logging.curator.key"),
			"ops-cert": utils.GetFileContents("/tmp/_working_dir/system.logging.curator.crt"),
		})

	utils.AddOwnerRefToObject(curatorSecret, utils.AsOwner(logging))

	err := sdk.Create(curatorSecret)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Curator secret: %v", err)
	}

	return nil
}

func getCuratorCronJob(logging *logging.ClusterLogging, curatorName string, elasticsearchHost string) *batch.CronJob {
	curatorContainer := utils.Container("curator", v1.PullIfNotPresent, logging.Spec.Curation.CuratorSpec.Resources)

	curatorContainer.Env = []v1.EnvVar{
		{Name: "K8S_HOST_URL", Value: ""},
		{Name: "ES_HOST", Value: elasticsearchHost},
		{Name: "ES_PORT", Value: "9200"},
		{Name: "ES_CLIENT_CERT", Value: ""},
		{Name: "ES_CLIENT_KEY", Value: ""},
		{Name: "ES_CA", Value: ""},
		{Name: "CURATOR_DEFAULT_DAYS", Value: ""},
		{Name: "CURATOR_SCRIPT_LOG_LEVEL", Value: ""},
		{Name: "CURATOR_LOG_LEVEL", Value: ""},
		{Name: "CURATOR_TIMEOUT", Value: ""},
	}

	curatorContainer.VolumeMounts = []v1.VolumeMount{
		{Name: "certs", ReadOnly: true, MountPath: "/etc/curator/keys"},
		{Name: "config", ReadOnly: true, MountPath: "/etc/curator/settings"},
	}

	curatorPodSpec := utils.PodSpec(
		curatorName,
		[]v1.Container{curatorContainer},
		[]v1.Volume{
			{Name: "config", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "curator"}}}},
			{Name: "certs", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "curator"}}},
		},
	)

	curatorPodSpec.RestartPolicy = v1.RestartPolicyNever
	curatorPodSpec.TerminationGracePeriodSeconds = utils.GetInt64(600)

	curatorCronJob := utils.CronJob(
		curatorName,
		logging.Namespace,
		"curator",
		curatorName,
		batch.CronJobSpec{
			SuccessfulJobsHistoryLimit: utils.GetInt32(1),
			FailedJobsHistoryLimit:     utils.GetInt32(1),
			Schedule:                   "30 3 * * *",
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
								"logging-infra": "curator",
							},
						},
						Spec: curatorPodSpec,
					},
				},
			},
		},
	)

	utils.AddOwnerRefToObject(curatorCronJob, utils.AsOwner(logging))

	return curatorCronJob
}

func createOrUpdateCuratorCronJob(logging *logging.ClusterLogging) (err error) {

	if utils.AllInOne(logging) {
		curatorCronJob := getCuratorCronJob(logging, "curator", "elasticsearch")

		err = sdk.Create(curatorCronJob)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing Curator cronjob: %v", err)
		}

		if err = updateCuratorIfRequired(curatorCronJob); err != nil {
			return
		}
	} else {
		curatorCronJob := getCuratorCronJob(logging, "curator-app", "elasticsearch-app")

		err = sdk.Create(curatorCronJob)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing Curator App cronjob: %v", err)
		}

		if err = updateCuratorIfRequired(curatorCronJob); err != nil {
			return
		}

		curatorInfraCronJob := getCuratorCronJob(logging, "curator-infra", "elasticsearch-infra")

		err = sdk.Create(curatorInfraCronJob)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing Curator Infra cronjob: %v", err)
		}

		if err = updateCuratorIfRequired(curatorInfraCronJob); err != nil {
			return
		}
	}

	return nil
}

func updateCuratorIfRequired(desired *batch.CronJob) (err error) {
	current := desired.DeepCopy()

	if err = sdk.Get(current); err != nil {
		return fmt.Errorf("Failed to get Curator cronjob: %v", err)
	}

	current, different := isCuratorDifferent(current, desired)

	if different {
		if err = sdk.Update(current); err != nil {
			return fmt.Errorf("Failed to update Curator cronjob: %v", err)
		}
	}

	return nil
}

func isCuratorDifferent(current *batch.CronJob, desired *batch.CronJob) (*batch.CronJob, bool) {

	different := false

	// Check schedule
	if current.Spec.Schedule != desired.Spec.Schedule {
		current.Spec.Schedule = desired.Spec.Schedule
		logrus.Infof("Invalid Curator schedule found, updating %q", current.Name)
		different = true
	}

	// Check suspended
	if current.Spec.Suspend != nil && desired.Spec.Suspend != nil && *current.Spec.Suspend != *desired.Spec.Suspend {
		current.Spec.Suspend = desired.Spec.Suspend
		logrus.Infof("Invalid Curator suspend value found, updating %q", current.Name)
		different = true
	}

	return current, different
}
