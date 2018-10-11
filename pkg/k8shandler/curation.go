package k8shandler

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	batchv1 "k8s.io/api/batch/v1"
	batch "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Note: this will eventually be deprecated and functionality will be moved into the ES operator
//   in the case of Curator. Other curation deployments may not be supported in the future

func CreateOrUpdateCuration(logging *logging.ClusterLogging) (err error) {
	if err = createOrUpdateCuratorServiceAccount(logging); err != nil {
		return
	}

	if err = createOrUpdateCuratorConfigMap(logging); err != nil {
		return
	}

	if err = createOrUpdateCuratorCronJob(logging); err != nil {
		return
	}

	return createOrUpdateCuratorSecret(logging)
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

func createOrUpdateCuratorCronJob(logging *logging.ClusterLogging) error {

	if utils.AllInOne(logging) {
		curatorCronJob := getCuratorCronJob(logging, "curator", "elasticsearch")

		err := sdk.Create(curatorCronJob)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing Curator cronjob: %v", err)
		}
	} else {
		curatorCronJob := getCuratorCronJob(logging, "curator-app", "elasticsearch-app")

		err := sdk.Create(curatorCronJob)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing Curator App cronjob: %v", err)
		}

		curatorInfraCronJob := getCuratorCronJob(logging, "curator-infra", "elasticsearch-infra")

		err = sdk.Create(curatorInfraCronJob)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing Curator Infra cronjob: %v", err)
		}
	}

	return nil
}
