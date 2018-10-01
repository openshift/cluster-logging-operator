package k8shandler

import (
  "k8s.io/apimachinery/pkg/api/errors"
  "github.com/sirupsen/logrus"
  "github.com/openshift/cluster-logging-operator/pkg/utils"
  "k8s.io/api/core/v1"

  sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
  logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  batch "k8s.io/api/batch/v1beta1"
  batchv1 "k8s.io/api/batch/v1"
)

// Note: this will eventually be deprecated and functionality will be moved into the ES operator
//   in the case of Curator. Other curation deployments may not be supported in the future

func CreateOrUpdateCuration(logging *logging.ClusterLogging) error {
  createOrUpdateCuratorServiceAccount(logging)

  createOrUpdateCuratorConfigMap(logging)
  createOrUpdateCuratorCronJob(logging)

  return createOrUpdateCuratorSecret(logging)
}

func createOrUpdateCuratorServiceAccount(logging *logging.ClusterLogging) error {

  curatorServiceAccount := utils.ServiceAccount("aggregated-logging-curator", logging.Namespace)

  err := sdk.Create(curatorServiceAccount)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure creating Curator service account: %v", err)
  }

  return nil
}

func createOrUpdateCuratorConfigMap(logging *logging.ClusterLogging) error {

  curatorConfigMap := utils.ConfigMap(
    "logging-curator",
    logging.Namespace,
    map[string]string{
      "actions.yaml": string(utils.GetFileContents("files/curator-actions.yaml")),
      "curator5.yaml": string(utils.GetFileContents("files/curator5-config.yaml")),
      "config.yaml": string(utils.GetFileContents("files/curator-config.yaml")),
    },
  )

  err := sdk.Create(curatorConfigMap)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure constructing Curator configmap: %v", err)
  }

  return nil
}

func createOrUpdateCuratorSecret(logging *logging.ClusterLogging) error {

  curatorSecret := utils.Secret(
    "logging-curator",
    logging.Namespace,
    map[string][]byte{
      "ca": utils.GetFileContents("/tmp/_working_dir/ca.crt"),
      "key": utils.GetFileContents("/tmp/_working_dir/system.logging.curator.key"),
      "cert": utils.GetFileContents("/tmp/_working_dir/system.logging.curator.crt"),
      "ops-ca": utils.GetFileContents("/tmp/_working_dir/ca.crt"),
      "ops-key": utils.GetFileContents("/tmp/_working_dir/system.logging.curator.key"),
      "ops-cert": utils.GetFileContents("/tmp/_working_dir/system.logging.curator.crt"),
    }  )

  err := sdk.Create(curatorSecret)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure constructing Curator secret: %v", err)
  }

  return nil
}

func createOrUpdateCuratorCronJob(logging *logging.ClusterLogging) error {

  curatorCronJob := utils.CronJob(
    "logging-curator",
    logging.Namespace,
    "curator",
    "curator",
    batch.CronJobSpec{
      SuccessfulJobsHistoryLimit: utils.GetInt32(1),
      FailedJobsHistoryLimit: utils.GetInt32(1),
      Schedule: "30 3 * * *",
      JobTemplate: batch.JobTemplateSpec{
        Spec: batchv1.JobSpec{
          BackoffLimit: utils.GetInt32(0),
          Parallelism: utils.GetInt32(1),
          Template: v1.PodTemplateSpec{
            ObjectMeta: metav1.ObjectMeta{
              Name: "logging-curator",
              Namespace: logging.Namespace,
              Labels: map[string]string{
                "provider": "openshift",
                "component": "curator",
                "logging-infra": "curator",
              },
            },
            Spec: v1.PodSpec{
              RestartPolicy: v1.RestartPolicyNever,
              TerminationGracePeriodSeconds: utils.GetInt64(600),
              ServiceAccountName: "aggregated-logging-curator",
              Containers: []v1.Container{
                {Name: "curator", Image: "openshift/origin-logging-curator5:latest", ImagePullPolicy: v1.PullIfNotPresent,
                  Env: []v1.EnvVar{
                    {Name: "K8S_HOST_URL", Value: ""},
                    {Name: "ES_HOST", Value: ""},
                    {Name: "ES_PORT", Value: ""},
                    {Name: "ES_CLIENT_CERT", Value: ""},
                    {Name: "ES_CLIENT_KEY", Value: ""},
                    {Name: "ES_CA", Value: ""},
                    {Name: "CURATOR_DEFAULT_DAYS", Value: ""},
                    {Name: "CURATOR_SCRIPT_LOG_LEVEL", Value: ""},
                    {Name: "CURATOR_LOG_LEVEL", Value: ""},
                    {Name: "CURATOR_TIMEOUT", Value: ""},
                  },
                  Resources: logging.Spec.Collection.Resources,
                  VolumeMounts: []v1.VolumeMount{
                    {Name: "certs", ReadOnly: true, MountPath: "/etc/curator/keys"},
                    {Name: "config", ReadOnly: true, MountPath: "/etc/curator/settings"},
                  },
                },
              },
              Volumes: []v1.Volume{
                {Name: "config", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "logging-curator"}}}},
                {Name: "certs", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "logging-curator"}}},
              },
            },
          },
        },
      },
    },
  )

  err := sdk.Create(curatorCronJob)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure constructing Curator cronjob: %v", err)
  }

  return nil
}
