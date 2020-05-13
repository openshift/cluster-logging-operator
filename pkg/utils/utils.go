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

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultWorkingDir = "/tmp/ocp-clo"
	OsNodeLabel       = "kubernetes.io/os"
	LinuxValue        = "linux"
)

// COMPONENT_IMAGES are thee keys are based on the "container name" + "-{image,version}"
var COMPONENT_IMAGES = map[string]string{
	"curator":  "CURATOR_IMAGE",
	"fluentd":  "FLUENTD_IMAGE",
	"promtail": "PROMTAIL_IMAGE",
	"kibana":   "KIBANA_IMAGE",
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
func CalculateMD5Hash(text string) (string, error) {
	hasher := md5.New()
	_, err := hasher.Write([]byte(text))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
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
	if osType, ok := selectors[OsNodeLabel]; ok {
		if osType == LinuxValue {
			return selectors
		}
		// Selector is provided but is not "linux"
		logrus.Warnf("Overriding node selector value: %s=%s to %s", OsNodeLabel, osType, LinuxValue)
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
			tolerationSecondsBool = *lhs.TolerationSeconds == *rhs.TolerationSeconds
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

	envVarName, ok := COMPONENT_IMAGES[component]
	if !ok {
		logrus.Errorf("Environment variable name mapping missing for component: %s", component)
		return ""
	}
	imageTag := os.Getenv(envVarName)
	if imageTag == "" {
		logrus.Errorf("No image tag defined for component '%s' by environment variable '%s'", component, envVarName)
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
		logrus.Errorf("Operator unable to read local file to get contents: %v", err)
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

func PodVolumeEquivalent(lhs, rhs []v1.Volume) bool {

	if len(lhs) != len(rhs) {
		return false
	}

	lhsMap := make(map[string]v1.Volume)
	rhsMap := make(map[string]v1.Volume)

	for _, vol := range lhs {
		lhsMap[vol.Name] = vol
	}

	for _, vol := range rhs {
		rhsMap[vol.Name] = vol
	}

	for name, lhsVol := range lhsMap {
		if rhsVol, ok := rhsMap[name]; ok {
			if lhsVol.Secret != nil && rhsVol.Secret != nil {
				if lhsVol.Secret.SecretName != rhsVol.Secret.SecretName {
					return false
				}

				continue
			}
			if lhsVol.ConfigMap != nil && rhsVol.ConfigMap != nil {
				if lhsVol.ConfigMap.LocalObjectReference.Name != rhsVol.ConfigMap.LocalObjectReference.Name {
					return false
				}

				continue
			}
			if lhsVol.HostPath != nil && rhsVol.HostPath != nil {
				if lhsVol.HostPath.Path != rhsVol.HostPath.Path {
					return false
				}
				continue
			}

			return false
		} else {
			// if rhsMap doesn't have the same key has lhsMap
			return false
		}
	}

	return true
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
	var rval bool
	if esource1.FieldRef != nil {
		if rval = reflect.DeepEqual(*esource1.FieldRef, *esource2.FieldRef); !rval {
			return rval
		}
	}
	if esource1.ResourceFieldRef != nil {
		if rval = EnvVarResourceFieldSelectorEqual(*esource1.ResourceFieldRef, *esource2.ResourceFieldRef); !rval {
			return rval
		}
	}
	if esource1.ConfigMapKeyRef != nil {
		if rval = reflect.DeepEqual(*esource1.ConfigMapKeyRef, *esource2.ConfigMapKeyRef); !rval {
			return rval
		}
	}
	if esource1.SecretKeyRef != nil {
		if rval = reflect.DeepEqual(*esource1.SecretKeyRef, *esource2.SecretKeyRef); !rval {
			return rval
		}
	}
	return true
}

func EnvVarResourceFieldSelectorEqual(resource1, resource2 v1.ResourceFieldSelector) bool {
	if (resource1.ContainerName == resource2.ContainerName) &&
		(resource1.Resource == resource2.Resource) &&
		(resource1.Divisor.Cmp(resource2.Divisor) == 0) {
		return true
	}
	return false
}

func SetProxyEnvVars(proxyConfig *configv1.Proxy) []v1.EnvVar {
	envVars := []v1.EnvVar{}
	if proxyConfig == nil {
		return envVars
	}
	if len(proxyConfig.Status.HTTPSProxy) != 0 {
		envVars = append(envVars, v1.EnvVar{
			Name:  "HTTPS_PROXY",
			Value: proxyConfig.Status.HTTPSProxy,
		},
			v1.EnvVar{
				Name:  "https_proxy",
				Value: proxyConfig.Status.HTTPSProxy,
			})
	}
	if len(proxyConfig.Status.HTTPProxy) != 0 {
		envVars = append(envVars, v1.EnvVar{
			Name:  "HTTP_PROXY",
			Value: proxyConfig.Status.HTTPProxy,
		},
			v1.EnvVar{
				Name:  "http_proxy",
				Value: proxyConfig.Status.HTTPProxy,
			})
	}
	if len(proxyConfig.Status.NoProxy) != 0 {
		envVars = append(envVars, v1.EnvVar{
			Name:  "NO_PROXY",
			Value: proxyConfig.Status.NoProxy,
		},
			v1.EnvVar{
				Name:  "no_proxy",
				Value: proxyConfig.Status.NoProxy,
			})
	}
	return envVars
}
