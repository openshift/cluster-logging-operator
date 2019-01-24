package utils

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"os"
	"time"
	"path"

	route "github.com/openshift/api/route/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1beta1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const WORKING_DIR = "/tmp/_working_dir"
const SPLIT_ANNOTATION = "io.openshift.clusterlogging.alpha/splitinstallation"

func AllInOne(cluster *logging.ClusterLogging) bool {

	//_, ok := cluster.ObjectMeta.Annotations[SPLIT_ANNOTATION]
	return true
}

// These keys are based on the "container name" + "-{image,version}"
var COMPONENT_IMAGES = map[string]string{
	"kibana":        "KIBANA_IMAGE",
	"kibana-proxy":  "OAUTH_PROXY_IMAGE",
	"curator":       "CURATOR_IMAGE",
	"fluentd":       "FLUENTD_IMAGE",
	"elasticsearch": "ELASTICSEARCH_IMAGE",
	"rsyslog":       "RSYSLOG_IMAGE",
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

func DoesClusterLoggingExist(cluster *logging.ClusterLogging) (bool, *logging.ClusterLogging) {

	// this is to avoid an invalid memory address panic
	if cluster == nil {
		return false, nil
	}

	if err := sdk.Get(cluster); err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		logrus.Errorf("Failed to check for ClusterLogging object: %v", err)
		return false, nil
	}

	return true, cluster
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

func RemoveServiceAccount(cluster *logging.ClusterLogging, serviceAccountName string) error {

	serviceAccount := ServiceAccount(serviceAccountName, cluster.Namespace)

	err := sdk.Delete(serviceAccount)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v service account: %v", serviceAccountName, err)
	}

	return nil
}

func RemoveConfigMap(cluster *logging.ClusterLogging, configmapName string) error {

	configMap := ConfigMap(
		configmapName,
		cluster.Namespace,
		map[string]string{ },
	)

	err := sdk.Delete(configMap)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v configmap: %v", configmapName, err)
	}

	return nil
}

func RemoveSecret(cluster *logging.ClusterLogging, secretName string) error {

	secret := Secret(
		secretName,
		cluster.Namespace,
		map[string][]byte{ },
	)

	err := sdk.Delete(secret)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v secret: %v", secretName, err)
	}

	return nil
}

func RemoveRoute(cluster *logging.ClusterLogging, routeName string) error {

	route := Route(
		routeName,
		cluster.Namespace,
		routeName,
		"",
	)

	err := sdk.Delete(route)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v route %v", routeName, err)
	}

	return nil
}

func RemoveService(cluster *logging.ClusterLogging, serviceName string) error {

	service := Service(
		serviceName,
		cluster.Namespace,
		serviceName,
		[]core.ServicePort{ },
	)

	err := sdk.Delete(service)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v service %v", serviceName, err)
	}

	return nil
}

func RemoveOAuthClient(cluster *logging.ClusterLogging, clientName string) error {

	oauthClient := OAuthClient(
		clientName,
		cluster.Namespace,
		"",
		[]string{ },
		[]string{ },
	)

	err := sdk.Delete(oauthClient)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v oauthclient %v", clientName, err)
	}

	return nil
}

func RemoveDaemonset(cluster *logging.ClusterLogging, daemonsetName string) error {

	daemonset := DaemonSet(
		daemonsetName,
		cluster.Namespace,
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

func RemoveDeployment(cluster *logging.ClusterLogging, deploymentName string) error {

	deployment := Deployment(
		deploymentName,
		cluster.Namespace,
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

func RemoveCronJob(cluster *logging.ClusterLogging, cronjobName string) error {

	cronjob := CronJob(
		cronjobName,
		cluster.Namespace,
		cronjobName,
		cronjobName,
		batch.CronJobSpec{ },
	)

	err := sdk.Delete(cronjob)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v cronjob %v", cronjobName, err)
	}

	return nil
}
