package utils

import (
	// #nosec G501
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultWorkingDir = "/tmp/ocp-clo"
	DefaultScriptsDir = "./scripts"
	OsNodeLabel       = "kubernetes.io/os"
	LinuxValue        = "linux"
)

// COMPONENT_IMAGES are thee keys are based on the "container name" + "-{image,version}"
var COMPONENT_IMAGES = map[string]string{
	"curator":             "CURATOR_IMAGE",
	constants.FluentdName: constants.FluentdImageEnvVar,
	"kibana":              "KIBANA_IMAGE",
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
	// #nosec G401
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
		log.Info("Overriding node selector value", "OsNodeLabel", OsNodeLabel, "osType", osType, "LinuxValue", LinuxValue)
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
		log.Error(errors.New("Can't get component image"), "Environment variable name mapping missing for component", "component", component)
		return ""
	}
	imageTag := os.Getenv(envVarName)
	if imageTag == "" {
		log.Error(errors.New("Can't get component image"), "No image tag defined for component by environment variable", "component", component, "envVarName", envVarName)
	}
	log.V(3).Info("Setting component image for to", "component", component, "imageTag", imageTag)
	return imageTag
}

func GetFileContents(filePath string) []byte {

	if filePath == "" {
		log.V(2).Info("Empty file path provided for retrieving file contents")
		return nil
	}

	contents, err := ioutil.ReadFile(filepath.Clean(filePath))
	if err != nil {
		log.Error(err, "Operator unable to read local file to get contents")
		return nil
	}

	return contents
}

const defaultShareDir = "/usr/share/logging"

func GetShareDir() string {
	// shareDir is <repo-root>/files - try to find from working dir.
	const sep = string(os.PathSeparator)
	const repoRoot = sep + "cluster-logging-operator" + sep
	wd, err := os.Getwd()
	if err != nil {
		return defaultShareDir
	}
	i := strings.LastIndex(wd, repoRoot)
	if i >= 0 {
		return filepath.Join(wd[0:i+len(repoRoot)] + "files")
	}
	return defaultShareDir
}

func GetScriptsDir() string {
	scriptsDir := os.Getenv("SCRIPTS_DIR")
	if scriptsDir == "" {
		return DefaultScriptsDir
	}
	return scriptsDir
}

func GetWorkingDirFileContents(filePath string) []byte {
	return GetFileContents(GetWorkingDirFilePath(filePath))
}
func GetWorkingDir() string {
	workingDir := os.Getenv("WORKING_DIR")
	if workingDir == "" {
		workingDir = DefaultWorkingDir
	}
	return workingDir
}
func GetWorkingDirFilePath(toFile string) string {
	return path.Join(GetWorkingDir(), toFile)
}

func WriteToWorkingDirFile(toFile string, value []byte) error {

	if err := ioutil.WriteFile(GetWorkingDirFilePath(toFile), value, 0600); err != nil {
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
	return resource1.ContainerName == resource2.ContainerName &&
		resource1.Resource == resource2.Resource &&
		resource1.Divisor.Cmp(resource2.Divisor) == 0
}

func GetProxyEnvVars() []v1.EnvVar {
	envVars := []v1.EnvVar{}
	for _, envvar := range []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy", "NO_PROXY", "no_proxy"} {
		if value := os.Getenv(envvar); value != "" {
			envVars = append(envVars, v1.EnvVar{
				Name:  envvar,
				Value: value,
			})
		}
	}
	return envVars
}

// WrapError wraps some types of error to provide more informative Error() message.
// If err is exec.ExitError and has Stderr text, include it in Error()
// Otherwise return err unchanged.
func WrapError(err error) error {
	exitErr := &exec.ExitError{}
	if errors.As(err, &exitErr) && len(exitErr.Stderr) != 0 {
		return fmt.Errorf("%w: %v", err, string(exitErr.Stderr))
	}
	return err
}

func ToJsonLogs(logs []string) string {
	return fmt.Sprintf("[%s]", strings.Join(logs, ","))
}
