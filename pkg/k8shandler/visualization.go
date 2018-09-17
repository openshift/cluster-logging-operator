package k8shandler

import (
  "k8s.io/apimachinery/pkg/api/errors"
  "github.com/sirupsen/logrus"
  "github.com/openshift/cluster-logging-operator/pkg/utils"
  "k8s.io/api/core/v1"
  "k8s.io/apimachinery/pkg/util/intstr"

  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
  logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
)

func CreateOrUpdateVisualization(logging *logging.ClusterLogging) error {
  createOrUpdateKibanaServiceAccount(logging)
  createOrUpdateKibanaService(logging)
  createOrUpdateKibanaRoute(logging)
  createOrUpdateKibanaDeployment(logging)
  return createOrUpdateKibanaSecret(logging)
}

func createOrUpdateKibanaServiceAccount(logging *logging.ClusterLogging) error {

  kibanaServiceAccount := utils.ServiceAccount("aggregated-logging-kibana", logging.Namespace)

  err := sdk.Create(kibanaServiceAccount)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure creating Kibana service account: %v", err)
  }

  return nil
}

func createOrUpdateKibanaDeployment(logging *logging.ClusterLogging) error {

  kibanaPodSpec := kibanaPodSpec(logging)
  kibanaDeployment := utils.Deployment("logging-kibana", logging.Namespace, "kibana", "kibana", kibanaPodSpec)

  err := sdk.Create(kibanaDeployment)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure creating Kibana deployment: %v", err)
  }

  return nil
}

func createOrUpdateKibanaRoute(logging *logging.ClusterLogging) error {

  kibanaRoute := utils.Route(
    "logging-kibana",
    logging.Namespace,
    "logging.example.com",
    "logging-kibana",
  )

  err := sdk.Create(kibanaRoute)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure creating Kibana route: %v", err)
  }

  return nil
}

func createOrUpdateKibanaService(logging *logging.ClusterLogging) error {

  kibanaService := utils.Service(
    "logging-kibana",
    logging.Namespace,
    "kibana",
    []v1.ServicePort{
      { Port: 443, TargetPort: intstr.IntOrString{
        Type: intstr.String,
        StrVal: "oaproxy",
        } },
    })

  err := sdk.Create(kibanaService)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure constructing Kibana service: %v", err)
  }

  return nil
}

func createOrUpdateKibanaSecret(logging *logging.ClusterLogging) error {

  kibanaSecret := utils.Secret(
    "logging-kibana",
    logging.Namespace,
    map[string][]byte{
      "ca": utils.GetFileContents("/tmp/_working_dir/ca.crt"),
      "key": utils.GetFileContents("/tmp/_working_dir/system.logging.kibana.key"),
      "cert": utils.GetFileContents("/tmp/_working_dir/system.logging.kibana.crt"),
    }  )

  err := sdk.Create(kibanaSecret)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure constructing Kibana secret: %v", err)
  }

  proxySecret := utils.Secret(
    "logging-kibana-proxy",
    logging.Namespace,
    map[string][]byte{
      "oauth-secret": utils.GetRandomWord(64),
      "session-secret": utils.GetRandomWord(32),
      "server-key": utils.GetFileContents("/tmp/_working_dir/kibana-internal.key"),
      "server-cert": utils.GetFileContents("/tmp/_working_dir/kibana-internal.crt"),
    }  )

  err = sdk.Create(proxySecret)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure constructing Kibana Proxy secret: %v", err)
  }

  return nil
}

func kibanaPodSpec(logging *logging.ClusterLogging) v1.PodSpec {
  return v1.PodSpec{
    Affinity: &v1.Affinity{
      PodAntiAffinity: &v1.PodAntiAffinity{
        PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
          {
            Weight: 100,
            PodAffinityTerm: v1.PodAffinityTerm{
              LabelSelector: &metav1.LabelSelector{
                MatchExpressions: []metav1.LabelSelectorRequirement{
                  {Key: "logging-infra", Operator: metav1.LabelSelectorOpIn, Values: []string{"kibana"}},
                },
              },
              TopologyKey: "kubernetes.io/hostname",
            },
          },
        },
      },
    },
    Containers: []v1.Container{
      {Name: "kibana", Image: "openshift/origin-logging-kibana5:latest", ImagePullPolicy: v1.PullIfNotPresent,
        Env: []v1.EnvVar{
          {Name: "ELASTICSEARCH_URL", Value: "https://logging-es:9200"},
          {Name: "KIBANA_MEMORY_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "kibana", Resource: "limits.memory"}}},
        },
        Resources: logging.Spec.Visualization.Resources,
        VolumeMounts: []v1.VolumeMount{
          {Name: "kibana", ReadOnly: true, MountPath: "/etc/kibana/keys"},
        },
        ReadinessProbe: &v1.Probe{
          Handler: v1.Handler{
            Exec: &v1.ExecAction{
              Command: []string{
                "/usr/share/kibana/probe/readiness.sh",
              },
            },
          },
          InitialDelaySeconds: 5, TimeoutSeconds: 4, PeriodSeconds: 5,
        },
      },
      {Name: "kibana-proxy", Image: "openshift/oauth-proxy:latest", ImagePullPolicy: v1.PullIfNotPresent,
        Args: []string{
          "--upstream-ca=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
          "--https-address=:3000",
          "-provider=openshift",
          "-client-id=kibana-proxy",
          "-client-secret-file=/secret/oauth-secret",
          "-cookie-secret-file=/secret/session-secret",
          "-upstream=http://localhost:5601",
          "-scope=user:info user:check-access user:list-projects",
          "--tls-cert=/secret/server-cert",
          "-tls-key=/secret/server-key",
          "-pass-access-token",
          "-skip-provider-button",
        },
        Env: []v1.EnvVar{
          {Name: "OAP_DEBUG", Value: "false"},
          {Name: "OCP_AUTH_PROXY_MEMORY_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "kibana-proxy", Resource: "limits.memory"}}},
        },
        Resources: logging.Spec.Visualization.Proxy.Resources,
        Ports: []v1.ContainerPort{
          {Name: "oaproxy", ContainerPort: 3000},
        },
        VolumeMounts: []v1.VolumeMount{
          {Name: "kibana-proxy", ReadOnly: true, MountPath: "/secret"},
        },
      },
    },
    ServiceAccountName: "aggregated-logging-kibana",
    Volumes: []v1.Volume{
      {Name: "kibana", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "logging-kibana"}}},
      {Name: "kibana-proxy", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "logging-kibana-proxy"}}},
    },
  }
}
