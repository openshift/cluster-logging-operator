package utils

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"

	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func CheckFileExists(filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("'%s' not found", filePath)
		}
		return err
	}
	return nil
}
