package k8shandler

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

//createOrGetTrustedCABundleConfigMap creates or returns an existing Trusted CA Bundle ConfigMap.
//By setting label "config.openshift.io/inject-trusted-cabundle: true", the cert is automatically filled/updated.
func (clusterRequest *ClusterLoggingRequest) createOrGetTrustedCABundleConfigMap(name string) (*corev1.ConfigMap, error) {
	logrus.Debug("createOrGetTrustedCABundleConfigMap...")
	configMap := NewConfigMap(
		name,
		clusterRequest.Cluster.Namespace,
		map[string]string{
			constants.TrustedCABundleKey: "",
		},
	)
	configMap.ObjectMeta.Labels = make(map[string]string)
	configMap.ObjectMeta.Labels[constants.InjectTrustedCABundleLabel] = "true"

	utils.AddOwnerRefToObject(configMap, utils.AsOwner(clusterRequest.Cluster))

	err := clusterRequest.Create(configMap)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("failed to create trusted CA bundle config map %q for %q: %s", name, clusterRequest.Cluster.Name, err)
		}

		// Get the existing config map which may include an injected CA bundle
		if err = clusterRequest.Get(configMap.Name, configMap); err != nil {
			if errors.IsNotFound(err) {
				// the object doesn't exist -- it was likely culled
				// recreate it on the next time through if necessary
				return nil, fmt.Errorf("failed to find trusted CA bundle config map %q for %q: %s", name, clusterRequest.Cluster.Name, err)
			}
			return nil, fmt.Errorf("failed to get trusted CA bundle config map %q for %q: %s", name, clusterRequest.Cluster.Name, err)
		}
	}
	return configMap, err
}

func hasTrustedCABundle(configMap *corev1.ConfigMap) bool {
	if configMap == nil {
		return false
	}
	caBundle, ok := configMap.Data[constants.TrustedCABundleKey]
	return ok && caBundle != ""
}

func calcTrustedCAHashValue(configMap *corev1.ConfigMap) (string, error) {
	hashValue := ""
	var err error

	if configMap == nil {
		return hashValue, nil
	}
	caBundle, ok := configMap.Data[constants.TrustedCABundleKey]
	if ok && caBundle != "" {
		hashValue, err = utils.CalculateMD5Hash(caBundle)
		if err != nil {
			return "", err
		}
	}

	if !ok {
		return "", fmt.Errorf("Expected key %v does not exist in %v", constants.TrustedCABundleKey, configMap.Name)
	}

	return hashValue, nil
}
