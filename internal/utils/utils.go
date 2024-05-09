package utils

import (
	// #nosec G501
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	OsNodeLabel = "kubernetes.io/os"
	LinuxValue  = "linux"
)

var (
	DefaultNodeSelector = map[string]string{OsNodeLabel: LinuxValue}
)

// COMPONENT_IMAGES are thee keys are based on the "container name" + "-{image,version}"
var COMPONENT_IMAGES = map[string]string{
	constants.VectorName:                 constants.VectorImageEnvVar,
	constants.LogfilesmetricexporterName: constants.LogfilesmetricImageEnvVar,
}

func AsOwner(o runtime.Object) metav1.OwnerReference {
	m, err := meta.Accessor(o)
	if err != nil {
		panic(err)
	}
	return metav1.OwnerReference{
		APIVersion: fmt.Sprintf("%s/%s", o.GetObjectKind().GroupVersionKind().Group, o.GetObjectKind().GroupVersionKind().Version),
		Kind:       o.GetObjectKind().GroupVersionKind().Kind,
		Name:       m.GetName(),
		UID:        m.GetUID(),
		Controller: GetPtr(true),
	}
}

func HasSameOwner(curr []metav1.OwnerReference, desired []metav1.OwnerReference) bool {
	return reflect.DeepEqual(curr, desired)
}

// CalculateMD5Hash returns a MD5 hash of the give text
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

// AddOwnerRefToObject adds the parent as an owner to the child
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

	contents, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		log.Error(err, "Operator unable to read local file to get contents")
		return nil
	}

	return contents
}

func GetPtr[T any](value T) *T {
	return &value
}

// GetEnvVar returns EnvVar entry that matches the given name
func GetEnvVar(name string, envVars []v1.EnvVar) (v1.EnvVar, bool) {
	for _, env := range envVars {
		if env.Name == name {
			return env, true
		}
	}
	return v1.EnvVar{}, false
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
		log.V(9).Info("PodVolumeEquivalent unequal lengths", "left", lhs, "right", rhs)
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
					log.V(9).Info("PodVolumeEquivalent unmatched secretName", "name", name, "left", lhsVol, "right", rhsVol)
					return false
				}

				continue
			}
			if lhsVol.ConfigMap != nil && rhsVol.ConfigMap != nil {
				if lhsVol.ConfigMap.LocalObjectReference.Name != rhsVol.ConfigMap.LocalObjectReference.Name {
					log.V(9).Info("PodVolumeEquivalent unmatched configMap.LocalObjectReference", "name", name, "left", lhsVol.ConfigMap, "right", rhsVol.ConfigMap)
					return false
				}

				continue
			}
			if lhsVol.HostPath != nil && rhsVol.HostPath != nil {
				if lhsVol.HostPath.Path != rhsVol.HostPath.Path {
					log.V(9).Info("PodVolumeEquivalent unmatched hostpath", "name", name, "left", lhsVol.HostPath, "right", rhsVol.HostPath)
					return false
				}
				continue
			}
			if lhsVol.EmptyDir != nil && rhsVol.EmptyDir != nil {
				if lhsVol.EmptyDir.Medium != rhsVol.EmptyDir.Medium {
					log.V(9).Info("PodVolumeEquivalent unmatched emptyDir", "name", name, "left", lhsVol.EmptyDir, "right", rhsVol.EmptyDir)
					return false
				}
				continue
			}

			return false
		} else {
			// if rhsMap doesn't have the same key has lhsMap
			log.V(9).Info("PodVolumeEquivalent missing volume", "name", name, "left", lhs, "right", rhs)
			return false
		}
	}

	return true
}

/*
*
EnvValueEqual - check if 2 EnvValues are equal or not
Notes:
- reflect.DeepEqual does not return expected results if the to-be-compared value is a pointer.
- needs to adjust with k8s.io/api/core/v#/types.go when the types are updated.
*
*/
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
			if envvar == "NO_PROXY" || envvar == "no_proxy" {
				if len(constants.ExtraNoProxyList) > 0 {
					value = strings.Join(constants.ExtraNoProxyList, ",") + "," + value
				}
			}
			envVars = append(envVars, v1.EnvVar{
				Name:  strings.ToLower(envvar),
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
	if len(logs) == 1 && strings.HasPrefix(logs[0], "[") && strings.HasSuffix(logs[0], "]") {
		return logs[0]
	}
	return fmt.Sprintf("[%s]", strings.Join(logs, ","))
}

func AddLabels(object metav1.Object, labels map[string]string) {
	existed := object.GetLabels()
	for key, val := range existed {
		labels[key] = val
	}
	object.SetLabels(labels)
}
