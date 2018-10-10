package utils

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	route "github.com/openshift/api/route/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1beta1"
	"k8s.io/api/core/v1"
	scheduling "k8s.io/api/scheduling/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const WORKING_DIR = "/tmp/_working_dir"
const ALLINONE_ANNOTATION = "io.openshift.clusterlogging.alpha/allinone"

func AllInOne(logging *logging.ClusterLogging) bool {

	_, ok := logging.ObjectMeta.Annotations[ALLINONE_ANNOTATION]
	return ok
}

// These keys are based on the "container name" + "-{image,version}"
var COMPONENT_IMAGES = map[string]string{
	"kibana-image":          "logging-kibana5",
	"kibana-version":        "latest",
	"kibana-proxy-image":    "oauth-proxy",
	"kibana-proxy-version":  "latest",
	"curator-image":         "logging-curator5",
	"curator-version":       "latest",
	"fluentd-image":         "logging-fluentd",
	"fluentd-version":       "latest",
	"elasticsearch-image":   "logging-elasticsearch5",
	"elasticsearch-version": "latest",
}

func getImageName(component string) string {

	var componentKey bytes.Buffer
	componentKey.WriteString(component)
	componentKey.WriteString("-image")

	imageName, ok := COMPONENT_IMAGES[componentKey.String()]

	if ok {
		return imageName
	}

	return ""
}

func getImageVersion(component string) string {

	var componentKey bytes.Buffer
	componentKey.WriteString(component)
	componentKey.WriteString("-version")

	imageVersion, ok := COMPONENT_IMAGES[componentKey.String()]

	if ok {
		return imageVersion
	}

	return ""
}

func getImagePrefix(component string) string {

	repoPrefix := os.Getenv("REPO_PREFIX")
	imagePrefix := os.Getenv("IMAGE_PREFIX")

	var prefix bytes.Buffer

	switch component {
	case "kibana":
		prefix.WriteString(repoPrefix)
		prefix.WriteString(imagePrefix)

	case "kibana-proxy":
		prefix.WriteString(repoPrefix)

	case "curator":
		prefix.WriteString(repoPrefix)
		prefix.WriteString(imagePrefix)

	case "fluentd":
		prefix.WriteString(repoPrefix)
		prefix.WriteString(imagePrefix)

	case "elasticsearch":
		prefix.WriteString(repoPrefix)
		prefix.WriteString(imagePrefix)

	default:
		return ""
	}

	return prefix.String()
}

func AsOwner(o *logging.ClusterLogging) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: logging.SchemeGroupVersion.String(),
		Kind:       o.Kind,
		Name:       o.Name,
		UID:        o.UID,
		Controller: GetBool(true),
	}
}

func AddOwnerRefToObject(object metav1.Object, ownerRef metav1.OwnerReference) {
	if (metav1.OwnerReference{}) != ownerRef {
		object.SetOwnerReferences(append(object.GetOwnerReferences(), ownerRef))
	}
}

func GetComponentImage(component string) string {

	var image bytes.Buffer

	image.WriteString(getImagePrefix(component))
	image.WriteString(getImageName(component))
	image.WriteString(":")
	image.WriteString(getImageVersion(component))

	return image.String()
}

func GetFileContents(filePath string) []byte {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		logrus.Errorf("Unable to read file to get contents: %v", err)
		return nil
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
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Type: "Opaque",
		Data: data,
	}
}

func ServiceAccount(accountName string, namespace string) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      accountName,
			Namespace: namespace,
		},
	}
}

func Service(serviceName string, namespace string, selectorComponent string, servicePorts []v1.ServicePort) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels: map[string]string{
				"logging-infra": "support",
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"component": selectorComponent,
				"provider":  "openshift",
			},
			Ports: servicePorts,
		},
	}
}

func Route(routeName string, namespace string, hostName string, serviceName string) *route.Route {
	return &route.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      routeName,
			Namespace: namespace,
			Labels: map[string]string{
				"component":     "support",
				"logging-infra": "support",
				"provider":      "openshift",
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

func PodSpec(serviceAccountName string, containers []v1.Container, volumes []v1.Volume) v1.PodSpec {
	return v1.PodSpec{
		Containers:         containers,
		ServiceAccountName: serviceAccountName,
		Volumes:            volumes,
	}
}

func Container(containerName string, pullPolicy v1.PullPolicy, resources v1.ResourceRequirements) v1.Container {
	return v1.Container{
		Name:            containerName,
		Image:           GetComponentImage(containerName),
		ImagePullPolicy: pullPolicy,
		Resources:       resources,
	}
}

// loggingComponent = kibana
// component = kibana{,-ops}
func Deployment(deploymentName string, namespace string, loggingComponent string, component string, podSpec v1.PodSpec) *apps.Deployment {
	return &apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
			Labels: map[string]string{
				"provider":      "openshift",
				"component":     component,
				"logging-infra": loggingComponent,
			},
		},
		Spec: apps.DeploymentSpec{
			Replicas: GetInt32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"provider":      "openshift",
					"component":     component,
					"logging-infra": loggingComponent,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: deploymentName,
					Labels: map[string]string{
						"provider":      "openshift",
						"component":     component,
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

func DaemonSet(daemonsetName string, namespace string, loggingComponent string, component string, podSpec v1.PodSpec) *apps.DaemonSet {
	return &apps.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      daemonsetName,
			Namespace: namespace,
			Labels: map[string]string{
				"provider":      "openshift",
				"component":     component,
				"logging-infra": loggingComponent,
			},
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/port":   "24231",
				"prometheus.io/scheme": "http",
			},
		},
		Spec: apps.DaemonSetSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: daemonsetName,
					Labels: map[string]string{
						"provider":      "openshift",
						"component":     component,
						"logging-infra": loggingComponent,
					},
					Annotations: map[string]string{
						"scheduler.alpha.kubernetes.io/critical-pod": "''",
					},
				},
				Spec: podSpec,
			},
			UpdateStrategy: apps.DaemonSetUpdateStrategy{
				Type:          apps.RollingUpdateDaemonSetStrategyType,
				RollingUpdate: &apps.RollingUpdateDaemonSet{},
			},
			MinReadySeconds: 600,
		},
	}
}

func ConfigMap(configmapName string, namespace string, data map[string]string) *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmapName,
			Namespace: namespace,
		},
		Data: data,
	}
}

func PriorityClass(priorityclassName string, priorityValue int32, globalDefault bool, description string) *scheduling.PriorityClass {
	return &scheduling.PriorityClass{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PriorityClass",
			APIVersion: "scheduling.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: priorityclassName,
		},
		Value:         priorityValue,
		GlobalDefault: globalDefault,
		Description:   description,
	}
}

func GetBool(value bool) *bool {
	b := value
	return &b
}

func GetInt32(value int32) *int32 {
	i := value
	return &i
}

func GetInt64(value int64) *int64 {
	i := value
	return &i
}

func CronJob(cronjobName string, namespace string, loggingComponent string, component string, cronjobSpec batch.CronJobSpec) *batch.CronJob {
	return &batch.CronJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: "batch/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cronjobName,
			Namespace: namespace,
			Labels: map[string]string{
				"provider":      "openshift",
				"component":     component,
				"logging-infra": loggingComponent,
			},
		},
		Spec: cronjobSpec,
	}
}

func GetDeploymentList(namespace string, selector string) (*apps.DeploymentList, error) {
	list := &apps.DeploymentList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
	}

	err := sdk.List(
		namespace,
		list,
		sdk.WithListOptions(&metav1.ListOptions{
			LabelSelector: selector,
		}),
	)

	return list, err
}

func GetReplicaSetList(namespace string, selector string) (*apps.ReplicaSetList, error) {
	list := &apps.ReplicaSetList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicaSet",
			APIVersion: "apps/v1",
		},
	}

	err := sdk.List(
		namespace,
		list,
		sdk.WithListOptions(&metav1.ListOptions{
			LabelSelector: selector,
		}),
	)

	return list, err
}

func GetPodList(namespace string, selector string) (*v1.PodList, error) {
	list := &v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}

	err := sdk.List(
		namespace,
		list,
		sdk.WithListOptions(&metav1.ListOptions{
			LabelSelector: selector,
		}),
	)

	return list, err
}

func GetCronJobList(namespace string, selector string) (*batch.CronJobList, error) {
	list := &batch.CronJobList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: "batch/v1beta1",
		},
	}

	err := sdk.List(
		namespace,
		list,
		sdk.WithListOptions(&metav1.ListOptions{
			LabelSelector: selector,
		}),
	)

	return list, err
}

func GetDaemonSetList(namespace string, selector string) (*apps.DaemonSetList, error) {
	list := &apps.DaemonSetList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "extensions/v1beta1",
		},
	}

	err := sdk.List(
		namespace,
		list,
		sdk.WithListOptions(&metav1.ListOptions{
			LabelSelector: selector,
		}),
	)

	return list, err
}
