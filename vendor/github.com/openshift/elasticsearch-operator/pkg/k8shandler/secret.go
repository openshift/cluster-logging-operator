package k8shandler

import (
	"crypto/sha256"
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getSecret(secretName, namespace string) *v1.Secret {
	secret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
	}

	err := sdk.Get(&secret)

	if err != nil {
		// check if doesn't exist
	}

	return &secret
}

func getSecretDataHash(secretName, namespace string) string {
	hash := ""

	secret := getSecret(secretName, namespace)

	dataHashes := make(map[string][32]byte)

	for key, data := range secret.Data {
		dataHashes[key] = sha256.Sum256([]byte(data))
	}

	sortedKeys := sortDataHashKeys(dataHashes)

	for _, key := range sortedKeys {
		hash = fmt.Sprintf("%s%s", hash, dataHashes[key])
	}

	return hash
}
