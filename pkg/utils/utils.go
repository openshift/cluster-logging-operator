package utils

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"

	route "github.com/openshift/api/route/v1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1beta1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

const WORKING_DIR = "/tmp/_working_dir"
const SPLIT_ANNOTATION = "io.openshift.clusterlogging.alpha/splitinstallation"

// COMPONENT_IMAGES are thee keys are based on the "container name" + "-{image,version}"
var COMPONENT_IMAGES = map[string]string{
	"kibana":        "KIBANA_IMAGE",
	"kibana-proxy":  "OAUTH_PROXY_IMAGE",
	"curator":       "CURATOR_IMAGE",
	"fluentd":       "FLUENTD_IMAGE",
	"elasticsearch": "ELASTICSEARCH_IMAGE",
	"rsyslog":       "RSYSLOG_IMAGE",
}

//KubernetesResource is an adapter between public and private impl of ClusterLogging
type KubernetesResource interface {
	SchemeGroupVersion() string
	Type() metav1.TypeMeta
	Meta() metav1.ObjectMeta
}

func AsOwner(o KubernetesResource) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: o.SchemeGroupVersion(),
		Kind:       o.Type().Kind,
		Name:       o.Meta().Name,
		UID:        o.Meta().UID,
		Controller: GetBool(true),
	}
}

//AddOwnerRefToObject adds the parent as an owner to the child
func AddOwnerRefToObject(object metav1.Object, ownerRef metav1.OwnerReference) {
	if (metav1.OwnerReference{}) != ownerRef {
		object.SetOwnerReferences(append(object.GetOwnerReferences(), ownerRef))
	}
}

// GetComponentImage returns a full image pull spec for a given component
// based on the component type
func GetComponentImage(component string) string {

	env_var_name, ok := COMPONENT_IMAGES[component]
	if !ok {
		logrus.Errorf("Environment variable name mapping missing for component: %s", component)
		return ""
	}
	imageTag := os.Getenv(env_var_name)
	if imageTag == "" {
		logrus.Errorf("No image tag defined for component '%s' by environment variable '%s'", component, env_var_name)
	}
	logrus.Debugf("Setting component image for '%s' to: '%s'", component, imageTag)
	return imageTag
}

// CreateClusterRole creates a cluser role or returns error
func CreateClusterRole(name string, rules []rbac.PolicyRule, cluster *logging.ClusterLogging) (*rbac.ClusterRole, error) {
	clusterRole := &rbac.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbac.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: rules,
	}

	AddOwnerRefToObject(clusterRole, AsOwner(cluster))

	err := sdk.Create(clusterRole)
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("Failure creating '%s' clusterrole: %v", name, err)
	}
	return clusterRole, nil
}

func GetFileContents(filePath string) []byte {

	if filePath == "" {
		logrus.Debug("Empty file path provided for retrieving file contents")
		return nil
	}

	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		logrus.Errorf("Unable to read file to get contents: %v", err)
		return nil
	}

	return contents
}

func GetWorkingDirFileContents(filePath string) []byte {
	return GetFileContents(GetWorkingDirFilePath(filePath))
}

func GetWorkingDirFilePath(toFile string) string {
	return path.Join(WORKING_DIR, toFile)
}

func WriteToWorkingDirFile(toFile string, value []byte) error {

	if err := ioutil.WriteFile(GetWorkingDirFilePath(toFile), value, 0644); err != nil {
		return fmt.Errorf("Unable to write to working dir: %v", err)
	}

	return nil
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

func GetDeploymentList(namespace string, selector string) (*apps.DeploymentList, error) {
	list := &apps.DeploymentList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: apps.SchemeGroupVersion.String(),
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
			APIVersion: apps.SchemeGroupVersion.String(),
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

func GetPodList(namespace string, selector string) (*core.PodList, error) {
	list := &core.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: core.SchemeGroupVersion.String(),
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
			APIVersion: batch.SchemeGroupVersion.String(),
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
			APIVersion: apps.SchemeGroupVersion.String(),
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

func GetRouteURL(routeName, namespace string) (string, error) {

	foundRoute := &route.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: route.SchemeGroupVersion.String(),
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
	}

	if err := sdk.Get(foundRoute); err != nil {
		if !errors.IsNotFound(err) {
			logrus.Errorf("Failed to check for ClusterLogging object: %v", err)
		}
		return "", err
	}

	return fmt.Sprintf("%s%s", "https://", foundRoute.Spec.Host), nil
}

func CreateOrUpdateSecret(secret *core.Secret) (err error) {
	err = sdk.Create(secret)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing %v secret: %v", secret.Name, err)
		}

		current := secret.DeepCopy()
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err = sdk.Get(current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %v secret: %v", secret.Name, err)
			}

			current.Data = secret.Data
			if err = sdk.Update(current); err != nil {
				return err
			}
			return nil
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
}

func RemoveServiceAccount(namespace string, serviceAccountName string) error {

	serviceAccount := ServiceAccount(serviceAccountName, namespace)

	err := sdk.Delete(serviceAccount)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v service account: %v", serviceAccountName, err)
	}

	return nil
}

func RemoveConfigMap(namespace string, configmapName string) error {

	configMap := ConfigMap(
		configmapName,
		namespace,
		map[string]string{},
	)

	err := sdk.Delete(configMap)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v configmap: %v", configmapName, err)
	}

	return nil
}

func RemoveSecret(namespace string, secretName string) error {

	secret := Secret(
		secretName,
		namespace,
		map[string][]byte{},
	)

	err := sdk.Delete(secret)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v secret: %v", secretName, err)
	}

	return nil
}

func RemoveRoute(namespace string, routeName string) error {

	route := Route(
		routeName,
		namespace,
		routeName,
		"",
	)

	err := sdk.Delete(route)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v route %v", routeName, err)
	}

	return nil
}

func RemoveService(namespace string, serviceName string) error {

	service := Service(
		serviceName,
		namespace,
		serviceName,
		[]core.ServicePort{},
	)

	err := sdk.Delete(service)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v service %v", serviceName, err)
	}

	return nil
}

func RemoveOAuthClient(namespace string, clientName string) error {

	oauthClient := OAuthClient(
		clientName,
		namespace,
		"",
		[]string{},
		[]string{},
	)

	err := sdk.Delete(oauthClient)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v oauthclient %v", clientName, err)
	}

	return nil
}

func RemoveDaemonset(namespace string, daemonsetName string) error {

	daemonset := DaemonSet(
		daemonsetName,
		namespace,
		daemonsetName,
		daemonsetName,
		core.PodSpec{},
	)

	err := sdk.Delete(daemonset)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v daemonset %v", daemonsetName, err)
	}

	return nil
}

func RemoveDeployment(namespace string, deploymentName string) error {

	deployment := Deployment(
		deploymentName,
		namespace,
		deploymentName,
		deploymentName,
		core.PodSpec{},
	)

	err := sdk.Delete(deployment)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v deployment %v", deploymentName, err)
	}

	return nil
}

func RemoveCronJob(namespace string, cronjobName string) error {

	cronjob := CronJob(
		cronjobName,
		namespace,
		cronjobName,
		cronjobName,
		batch.CronJobSpec{},
	)

	err := sdk.Delete(cronjob)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v cronjob %v", cronjobName, err)
	}

	return nil
}

func RemovePriorityClass(priorityclassName string) error {
	collectionPriorityClass := PriorityClass(
		priorityclassName,
		1000000,
		false,
		"This priority class is for the Cluster-Logging Collector",
	)

	err := sdk.Delete(collectionPriorityClass)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v priority class %v", priorityclassName, err)
	}

	return nil
}
