package k8shandler

import (
  "k8s.io/apimachinery/pkg/api/errors"
  "github.com/sirupsen/logrus"
  "github.com/openshift/cluster-logging-operator/pkg/utils"
  "k8s.io/api/core/v1"

  sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
  logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
)

func CreateOrUpdateCollection(logging *logging.ClusterLogging) error {
  createOrUpdateFluentdServiceAccount(logging)

  createOrUpdateCollectionPriorityClass(logging)
  createOrUpdateFluentdConfigMap(logging)
  createOrUpdateFluentdDaemonset(logging)

  return createOrUpdateFluentdSecret(logging)
}

func createOrUpdateCollectionPriorityClass(logging *logging.ClusterLogging) error {

  collectionPriorityClass := utils.PriorityClass("cluster-logging", 1000000, false, "This priority class is for the Cluster-Logging Collector")

  sdk.Create(collectionPriorityClass)
  // Currently we cannot do validation checking on this due to version issues with the scheduling api and the version we create the priorityclass with
  /*err := sdk.Create(collectionPriorityClass)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure creating Collection priority class: %v", err)
  }*/

  return nil
}

func createOrUpdateFluentdServiceAccount(logging *logging.ClusterLogging) error {

  fluentdServiceAccount := utils.ServiceAccount("aggregated-logging-fluentd", logging.Namespace)

  err := sdk.Create(fluentdServiceAccount)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure creating Fluentd service account: %v", err)
  }

  return nil
}

func createOrUpdateFluentdConfigMap(logging *logging.ClusterLogging) error {

  // TODO: update to use the actual files we want...
  fluentdConfigMap := utils.ConfigMap(
    "logging-fluentd",
    logging.Namespace,
    map[string]string{
      "fluent.conf": string(utils.GetFileContents("files/fluent.conf")),
      "throttle-config.yaml": string(utils.GetFileContents("files/fluentd-throttle-config.yaml")),
      "secure-forward.conf": string(utils.GetFileContents("files/secure-forward.conf")),
    },
  )

  err := sdk.Create(fluentdConfigMap)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure constructing Fluentd configmap: %v", err)
  }

  return nil
}

func createOrUpdateFluentdSecret(logging *logging.ClusterLogging) error {

  fluentdSecret := utils.Secret(
    "logging-fluentd",
    logging.Namespace,
    map[string][]byte{
      "ca": utils.GetFileContents("/tmp/_working_dir/ca.crt"),
      "key": utils.GetFileContents("/tmp/_working_dir/system.logging.fluentd.key"),
      "cert": utils.GetFileContents("/tmp/_working_dir/system.logging.fluentd.crt"),
      "ops-ca": utils.GetFileContents("/tmp/_working_dir/ca.crt"),
      "ops-key": utils.GetFileContents("/tmp/_working_dir/system.logging.fluentd.key"),
      "ops-cert": utils.GetFileContents("/tmp/_working_dir/system.logging.fluentd.crt"),
    }  )

  err := sdk.Create(fluentdSecret)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure constructing Fluentd secret: %v", err)
  }

  return nil
}

func createOrUpdateFluentdDaemonset(logging *logging.ClusterLogging) error {

  fluentdPodSpec := fluentdPodSpec(logging)
  fluentdDaemonset := utils.DaemonSet("logging-fluentd", logging.Namespace, "fluentd", "fluentd", fluentdPodSpec)

  err := sdk.Create(fluentdDaemonset)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure creating Fluentd Daemonset %v", err)
  }

  return nil
}

func fluentdPodSpec(logging *logging.ClusterLogging) v1.PodSpec {
  return v1.PodSpec{
    NodeSelector: map[string]string {
      "logging-infra-fluentd": "true",
    },
    PriorityClassName: "cluster-logging",
    Containers: []v1.Container{
      {Name: "fluentd", Image: "openshift/origin-logging-fluentd:latest", ImagePullPolicy: v1.PullIfNotPresent,
        SecurityContext: &v1.SecurityContext{
          Privileged: utils.GetBool(true),
        },
        Env: []v1.EnvVar{
          {Name: "MERGE_JSON_LOG", Value: ""},
          {Name: "K8S_HOST_URL", Value: ""},
          {Name: "ES_HOST", Value: ""},
          {Name: "ES_PORT", Value: ""},
          {Name: "ES_CLIENT_CERT", Value: ""},
          {Name: "ES_CLIENT_KEY", Value: ""},
          {Name: "ES_CA", Value: ""},
          {Name: "OPS_HOST", Value: ""},
          {Name: "OPS_PORT", Value: ""},
          {Name: "OPS_CLIENT_CERT", Value: ""},
          {Name: "OPS_CLIENT_KEY", Value: ""},
          {Name: "OPS_CA", Value: ""},
          {Name: "JOURNAL_SOURCE", Value: ""},
          {Name: "JOURNAL_READ_FROM_HEAD", Value: ""},
          {Name: "BUFFER_QUEUE_LIMIT", Value: ""},
          {Name: "BUFFER_SIZE_LIMIT", Value: ""},
          {Name: "FILE_BUFFER_LIMIT", Value: ""},
          {Name: "FLUENTD_CPU_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "fluentd", Resource: "limits.cpu"}}},
          {Name: "FLUENTD_MEMORY_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "fluentd", Resource: "limits.memory"}}},
        },
        Resources: logging.Spec.Collection.Resources,
        VolumeMounts: []v1.VolumeMount{
          {Name: "runlogjournal", MountPath: "/run/log/journal"},
          {Name: "varlog", ReadOnly: true, MountPath: "/var/log"},
          {Name: "varlibdockercontainers", ReadOnly: true, MountPath: "/var/lib/docker"},
          {Name: "config", ReadOnly: true, MountPath: "/etc/fluent/configs.d/user"},
          {Name: "certs", ReadOnly: true, MountPath: "/etc/fluent/keys"},
          {Name: "dockerhostname", ReadOnly: true, MountPath: "/etc/docker-hostname"},
          {Name: "localtime", ReadOnly: true, MountPath: "/etc/localtime"},
          {Name: "dockercfg", ReadOnly: true, MountPath: "/etc/sysconfig/docker"},
          {Name: "dockerdaemoncfg", ReadOnly: true, MountPath: "/etc/docker"},
          {Name: "filebufferstorage", MountPath: "/var/lib/fluentd"},
        },
      },
    },
    ServiceAccountName: "aggregated-logging-fluentd",
    Volumes: []v1.Volume{
      {Name: "runlogjournal", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/run/log/journal"}}},
      {Name: "varlog", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/log"}}},
      {Name: "varlibdockercontainers", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/docker"}}},
      {Name: "config", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "logging-fluentd"}}}},
      {Name: "certs", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "logging-fluentd"}}},
      {Name: "dockerhostname", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/hostname"}}},
      {Name: "localtime", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/localtime"}}},
      {Name: "dockercfg", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/sysconfig/docker"}}},
      {Name: "dockerdaemoncfg", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/docker"}}},
      {Name: "filebufferstorage", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/fluentd"}}},
    },
  }
}
