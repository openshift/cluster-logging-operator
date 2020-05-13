package kibana

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/openshift/elasticsearch-operator/pkg/logger"
	"github.com/openshift/elasticsearch-operator/pkg/utils"
	core "k8s.io/api/core/v1"
)

var secretCertificates = map[string]map[string]string{
	"kibana": {
		"ca":   "ca.crt",
		"key":  "system.logging.kibana.key",
		"cert": "system.logging.kibana.crt",
	},
	"kibana-proxy": {
		"server-key":     "kibana-internal.key",
		"server-cert":    "kibana-internal.crt",
		"session-secret": "kibana-session-secret",
	},
}

func (clusterRequest *KibanaRequest) GetSecret(secretName string) (*core.Secret, error) {
	secret := &core.Secret{}
	if err := clusterRequest.Get(secretName, secret); err != nil {
		if errors.IsNotFound(err) {
			return nil, err
		}
		return nil, fmt.Errorf("Failed to get %v secret: %v", secret.Name, err)
	}

	return secret, nil
}

func (clusterRequest *KibanaRequest) readSecrets() (err error) {

	for secretName, certMap := range secretCertificates {
		if err = clusterRequest.extractCertificates(secretName, certMap); err != nil {
			return
		}
	}

	return nil
}

func (clusterRequest *KibanaRequest) extractCertificates(secretName string, certs map[string]string) (err error) {

	for secretKey, certPath := range certs {
		if err = clusterRequest.extractSecretToFile(secretName, secretKey, certPath); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return
		}
	}

	return nil
}

func (clusterRequest *KibanaRequest) extractSecretToFile(secretName string, key string, toFile string) (err error) {
	secret, err := clusterRequest.GetSecret(secretName)
	if err != nil {
		if errors.IsNotFound(err) {
			return err
		}
		return fmt.Errorf("Unable to extract secret %s to file: %v", secretName, err)
	}

	value, ok := secret.Data[key]

	// check to see if the map value exists
	if !ok {
		logger.Warnf("No secret data %q found", key)
		return nil
	}

	return utils.WriteToWorkingDirFile(toFile, value)
}

func calcSecretHashValue(secret *core.Secret) (string, error) {
	hashValue := ""
	var err error

	if secret == nil {
		return hashValue, nil
	}

	var hashKeys []string
	var rawbytes []byte

	// we just want the keys here to sort them for consistently calculated hashes
	for key := range secret.Data {
		hashKeys = append(hashKeys, key)
	}

	sort.Strings(hashKeys)

	for _, key := range hashKeys {
		rawbytes = append(rawbytes, secret.Data[key]...)
	}

	hashValue, err = utils.CalculateMD5Hash(string(rawbytes))
	if err != nil {
		return "", err
	}

	return hashValue, nil
}
