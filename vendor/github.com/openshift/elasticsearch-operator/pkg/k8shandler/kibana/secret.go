package kibana

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/openshift/elasticsearch-operator/pkg/utils"
	core "k8s.io/api/core/v1"
)

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
