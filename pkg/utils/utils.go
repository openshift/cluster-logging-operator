package utils

import (
  "math/rand"
  "time"
  "io/ioutil"
  "github.com/sirupsen/logrus"

  "k8s.io/api/core/v1"
  apps "k8s.io/api/apps/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  route "github.com/openshift/api/route/v1"
)

const WORKING_DIR = "/tmp/_working_dir"

func GetFileContents(filePath string) []byte {
  contents, err := ioutil.ReadFile(filePath)
  if err != nil {
    logrus.Fatalf("Unable to read file to get contents: %v", err)
  }

  return contents
}

func init() {
  rand.Seed(time.Now().UnixNano())
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func GetRandomWord(wordSize int) []byte {
  b := make([]rune, wordSize)
  for i := range b {
    b[i] = letters[rand.Intn(len(letters))]
  }
  return []byte(string(b))
}

func Secret(secretName string, namespace string, data map[string][]byte) *v1.Secret {
  return &v1.Secret{
    TypeMeta: metav1.TypeMeta{
      Kind: "Secret",
      APIVersion: "v1",
    },
    ObjectMeta: metav1.ObjectMeta{
      Name: secretName,
      Namespace: namespace,
    },
    Type: "Opaque",
    Data: data,
  }
}

func ServiceAccount(accountName string, namespace string) *v1.ServiceAccount {
  return &v1.ServiceAccount{
    TypeMeta: metav1.TypeMeta{
      Kind: "ServiceAccount",
      APIVersion: "v1",
    },
    ObjectMeta: metav1.ObjectMeta{
      Name: accountName,
      Namespace: namespace,
    },
  }
}

func Service(serviceName string, namespace string, selectorComponent string, servicePorts []v1.ServicePort) *v1.Service {
  return &v1.Service{
    TypeMeta: metav1.TypeMeta{
      Kind: "Service",
      APIVersion: "v1",
    },
    ObjectMeta: metav1.ObjectMeta{
      Name: serviceName,
      Namespace: namespace,
      Labels: map[string]string{
        "logging-infra": "support",
      },
    },
    Spec: v1.ServiceSpec{
      Selector: map[string]string{
        "component": selectorComponent,
        "provider": "openshift",
      },
      Ports: servicePorts,
    },
  }
}

func Route(routeName string, namespace string, hostName string, serviceName string) *route.Route {
  return &route.Route{
    TypeMeta: metav1.TypeMeta{
      Kind: "Route",
      APIVersion: "route.openshift.io/v1",
    },
    ObjectMeta: metav1.ObjectMeta{
      Name: routeName,
      Namespace: namespace,
      Labels: map[string]string{
        "component": "support",
        "logging-infra": "support",
        "provider": "openshift",
      },
    },
    Spec: route.RouteSpec{
      Host: hostName,
      To: route.RouteTargetReference{
        Name: serviceName,
        Kind: "Service",
      },
    },
  }
}

// loggingComponent = kibana
// component = kibana{,-ops}
func Deployment(deploymentName string, namespace string, loggingComponent string, component string, podSpec v1.PodSpec) *apps.Deployment {
  return &apps.Deployment{
    TypeMeta: metav1.TypeMeta{
      Kind: "Deployment",
      APIVersion: "extensions/v1beta1",
    },
    ObjectMeta: metav1.ObjectMeta{
      Name: deploymentName,
      Namespace: namespace,
      Labels: map[string]string{
        "provider": "openshift",
        "component": component,
        "logging-infra": loggingComponent,
      },
    },
    Spec: apps.DeploymentSpec{
      //Replicas: 0,
      Selector: &metav1.LabelSelector{
        MatchLabels: map[string]string {
          "provider": "openshift",
          "component": component,
          "logging-infra": loggingComponent,
        },
      },
      Template: v1.PodTemplateSpec{
        ObjectMeta: metav1.ObjectMeta{
          Name: deploymentName,
          Labels: map[string]string{
            "provider": "openshift",
            "component": component,
            "logging-infra": loggingComponent,
          },
        },
        Spec: podSpec,
      },
      Strategy: apps.DeploymentStrategy{
        Type: apps.RollingUpdateDeploymentStrategyType,
        //RollingUpdate: {}
      },
    },
  }
}
