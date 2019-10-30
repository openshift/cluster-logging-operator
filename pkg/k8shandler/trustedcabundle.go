package k8shandler

import (
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	core "k8s.io/api/core/v1"
)

/*
 * Create or update Trusted CA Bundle ConfigMap
 * By setting label "config.openshift.io/inject-trusted-cabundle: true", the cert is automatically filled/updated.
 */
func (clusterRequest *ClusterLoggingRequest) createOrUpdateTrustedCABundleConfigMap(configMapName string) error {
	logrus.Debug("createOrUpdateTrustedCABundleConfigMap...")
	configMap := NewConfigMap(
		configMapName,
		clusterRequest.cluster.Namespace,
		map[string]string{
			constants.TrustedCABundleKey: "",
		},
	)
	configMap.ObjectMeta.Labels = make(map[string]string)
	configMap.ObjectMeta.Labels[constants.InjectTrustedCABundleLabel] = "true"

	utils.AddOwnerRefToObject(configMap, utils.AsOwner(clusterRequest.cluster))

	err := clusterRequest.CreateOrUpdateConfigMap(configMap)
	return err
}

func hasTrustedCABundle(configMap *core.ConfigMap) bool {
	if configMap == nil {
		return false
	}
	caBundle, ok := configMap.Data[constants.TrustedCABundleKey]
	if ok && caBundle != "" {
		return true
	} else {
		return false
	}
}

func calcTrustedCAHashValue(configMap *core.ConfigMap) string {
	hashValue := "0"
	if configMap == nil {
		return hashValue
	}
	caBundle, ok := configMap.Data[constants.TrustedCABundleKey]
	if ok && caBundle != "" {
		hashValue = utils.CalculateMD5Hash(caBundle)
	}
	return hashValue
}
