package utils

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	WORKING_DIR = "/tmp/_working_dir"
	OsNodeLabel = "kubernetes.io/os"
	LinuxValue  = "linux"
)

// COMPONENT_IMAGES are thee keys are based on the "container name" + "-{image,version}"
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
		Kind:       "ClusterLogging",
		Name:       o.Name,
		UID:        o.UID,
		Controller: GetBool(true),
	}
}

func AreSelectorsSame(lhs, rhs map[string]string) bool {

	if len(lhs) != len(rhs) {
		return false
	}

	for lhsKey, lhsVal := range lhs {
		rhsVal, ok := rhs[lhsKey]
		if !ok || lhsVal != rhsVal {
			return false
		}
	}

	return true
}

// EnsureLinuxNodeSelector takes given selector map and returns a selector map with linux node selector added into it.
// If there is already a node type selector and is different from "linux" then it is overridden and warning is logged.
// See https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#interlude-built-in-node-labels
func EnsureLinuxNodeSelector(selectors map[string]string) map[string]string {
	if selectors == nil {
		return map[string]string{OsNodeLabel: LinuxValue}
	}
	if os, ok := selectors[OsNodeLabel]; ok {
		if os == LinuxValue {
			return selectors
		}
		// Selector is provided but is not "linux"
		logrus.Warnf("Overriding node selector value: %s=%s to %s", OsNodeLabel, os, LinuxValue)
	}
	selectors[OsNodeLabel] = LinuxValue
	return selectors
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

func GetShareDir() string {
	shareDir := os.Getenv("LOGGING_SHARE_DIR")
	if shareDir == "" {
		return "/usr/share/logging"
	}
	return shareDir
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
