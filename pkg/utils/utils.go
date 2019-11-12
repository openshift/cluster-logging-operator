package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultWorkingDir = "/tmp/_working_dir"
	OsNodeLabel       = "kubernetes.io/os"
	LinuxValue        = "linux"
)

// COMPONENT_IMAGES are thee keys are based on the "container name" + "-{image,version}"
var COMPONENT_IMAGES = map[string]string{
	"kibana":        "KIBANA_IMAGE",
	"kibana-proxy":  "OAUTH_PROXY_IMAGE",
	"curator":       "CURATOR_IMAGE",
	"fluentd":       "FLUENTD_IMAGE",
	"elasticsearch": "ELASTICSEARCH_IMAGE",
	"promtail":      "PROMTAIL_IMAGE",
}

// GetAnnotation returns the value of an annoation for a given key and true if the key was found
func GetAnnotation(key string, meta metav1.ObjectMeta) (string, bool) {
	for k, value := range meta.Annotations {
		if k == key {
			return value, true
		}
	}
	return "", false
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

//CalculateMD5Hash returns a MD5 hash of the give text
func CalculateMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func AreMapsSame(lhs, rhs map[string]string) bool {
	return reflect.DeepEqual(lhs, rhs)
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

func AreTolerationsSame(lhs, rhs []v1.Toleration) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for _, lhsToleration := range lhs {
		if !containsToleration(lhsToleration, rhs) {
			return false
		}
	}

	return true

}

func containsToleration(toleration v1.Toleration, tolerations []v1.Toleration) bool {
	for _, t := range tolerations {
		if isTolerationSame(t, toleration) {
			return true
		}
	}

	return false
}

func isTolerationSame(lhs, rhs v1.Toleration) bool {

	tolerationSecondsBool := false
	// check that both are either null or not null
	if (lhs.TolerationSeconds == nil) == (rhs.TolerationSeconds == nil) {
		if lhs.TolerationSeconds != nil {
			// only compare values (attempt to dereference) if pointers aren't nil
			tolerationSecondsBool = (*lhs.TolerationSeconds == *rhs.TolerationSeconds)
		} else {
			tolerationSecondsBool = true
		}
	}

	return (lhs.Key == rhs.Key) &&
		(lhs.Operator == rhs.Operator) &&
		(lhs.Value == rhs.Value) &&
		(lhs.Effect == rhs.Effect) &&
		tolerationSecondsBool
}

func AppendTolerations(lhsTolerations, rhsTolerations []v1.Toleration) []v1.Toleration {
	if lhsTolerations == nil {
		lhsTolerations = []v1.Toleration{}
	}

	return append(lhsTolerations, rhsTolerations...)
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

	contents, err := ioutil.ReadFile(filepath.Clean(filePath))
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
	workingDir := os.Getenv("WORKING_DIR")
	if workingDir == "" {
		workingDir = DefaultWorkingDir
	}
	return path.Join(workingDir, toFile)
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

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

/**
EnvValueEqual - check if 2 EnvValues are equal or not
Notes:
- reflect.DeepEqual does not return expected results if the to-be-compared value is a pointer.
- needs to adjust with k8s.io/api/core/v#/types.go when the types are updated.
**/
func EnvValueEqual(env1, env2 []v1.EnvVar) bool {
	var found bool
	if len(env1) != len(env2) {
		return false
	}
	for _, elem1 := range env1 {
		found = false
		for _, elem2 := range env2 {
			if elem1.Name == elem2.Name {
				if elem1.Value != elem2.Value {
					return false
				}
				if (elem1.ValueFrom != nil && elem2.ValueFrom == nil) ||
					(elem1.ValueFrom == nil && elem2.ValueFrom != nil) {
					return false
				}
				if elem1.ValueFrom != nil {
					found = EnvVarSourceEqual(*elem1.ValueFrom, *elem2.ValueFrom)
				} else {
					found = true
				}
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func EnvVarSourceEqual(esource1, esource2 v1.EnvVarSource) bool {
	if (esource1.FieldRef != nil && esource2.FieldRef == nil) ||
		(esource1.FieldRef == nil && esource2.FieldRef != nil) ||
		(esource1.ResourceFieldRef != nil && esource2.ResourceFieldRef == nil) ||
		(esource1.ResourceFieldRef == nil && esource2.ResourceFieldRef != nil) ||
		(esource1.ConfigMapKeyRef != nil && esource2.ConfigMapKeyRef == nil) ||
		(esource1.ConfigMapKeyRef == nil && esource2.ConfigMapKeyRef != nil) ||
		(esource1.SecretKeyRef != nil && esource2.SecretKeyRef == nil) ||
		(esource1.SecretKeyRef == nil && esource2.SecretKeyRef != nil) {
		return false
	}
	if esource1.FieldRef != nil {
		return reflect.DeepEqual(*esource1.FieldRef, *esource2.FieldRef)
	}
	if esource1.ResourceFieldRef != nil {
		return reflect.DeepEqual(*esource1.ResourceFieldRef, *esource2.ResourceFieldRef)
	}
	if esource1.ConfigMapKeyRef != nil {
		return reflect.DeepEqual(*esource1.ConfigMapKeyRef, *esource2.ConfigMapKeyRef)
	}
	if esource1.SecretKeyRef != nil {
		return reflect.DeepEqual(*esource1.SecretKeyRef, *esource2.SecretKeyRef)
	}
	return true
}
