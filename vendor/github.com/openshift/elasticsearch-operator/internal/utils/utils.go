package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/ViaQ/logerr/kverrors"

	configv1 "github.com/openshift/api/config/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ViaQ/logerr/log"
)

const (
	DefaultWorkingDir = "/tmp/ocp-eo"
	OsNodeLabel       = "kubernetes.io/os"
	LinuxValue        = "linux"
)

// EnsureLinuxNodeSelector takes given selector map and returns a selector map with linux node selector added into it.
// If there is already a node type selector and is different from "linux" then it is overridden and warning is logged.
// See https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#interlude-built-in-node-labels
func EnsureLinuxNodeSelector(selectors map[string]string) map[string]string {
	if selectors == nil {
		return map[string]string{OsNodeLabel: LinuxValue}
	}
	if name, ok := selectors[OsNodeLabel]; ok {
		if name == LinuxValue {
			return selectors
		}
		// Selector is provided but is not "linux"
		log.Info("Overriding node selector value",
			"from", fmt.Sprintf("%s=%s", OsNodeLabel, name),
			"to", LinuxValue)
	}
	selectors[OsNodeLabel] = LinuxValue
	return selectors
}

func ToJSON(obj interface{}) (string, error) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return "", kverrors.Wrap(err, "failed to marshal JSON")
	}
	return string(bytes), nil
}

func LookupEnvWithDefault(envName, defaultValue string) string {
	if value, ok := os.LookupEnv(envName); ok {
		return value
	}
	return defaultValue
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func RandStringBytes(length uint) (string, error) {
	randString := make([]byte, length)
	_, err := rand.Read(randString)
	if err != nil {
		return "", kverrors.Wrap(err, "failed to generate random string")
	}

	for i, b := range randString {
		randString[i] = byte(letters[b%byte(len(letters))])
	}
	return string(randString), nil
}

// CalculateMD5Hash returns a MD5 hash of the give text
func CalculateMD5Hash(text string) (string, error) {
	hasher := md5.New()
	_, err := hasher.Write([]byte(text))
	if err != nil {
		return "", kverrors.Wrap(err, "failed to calculate hash")
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func AreMapsSame(lhs, rhs map[string]string) bool {
	return reflect.DeepEqual(lhs, rhs)
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

func GetWorkingDirFilePath(toFile string) string {
	workingDir := os.Getenv("WORKING_DIR")
	if workingDir == "" {
		workingDir = DefaultWorkingDir
	}
	return path.Join(workingDir, toFile)
}

func WriteToWorkingDirFile(toFile string, value []byte) error {
	if err := ioutil.WriteFile(GetWorkingDirFilePath(toFile), value, 0o644); err != nil {
		return kverrors.Wrap(err, "Unable to write to working dir")
	}

	return nil
}

func GetInt32(value int32) *int32 {
	i := value
	return &i
}

func GetInt64(value int64) *int64 {
	i := value
	return &i
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

func Contains(list []string, s string) bool {
	for _, item := range list {
		if s == item {
			return true
		}
	}

	return false
}

func GetMajorVersion(v string) string {
	ver := strings.Split(v, ".")
	return ver[0]
}
