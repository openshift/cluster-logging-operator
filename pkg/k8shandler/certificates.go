package k8shandler

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"io/ioutil"
	"k8s.io/api/core/v1"
	"os"
	"os/exec"
	"path"

	"github.com/openshift/cluster-logging-operator/pkg/utils"

	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	k8sutil "github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func extractSecretToFile(namespace string, secretName string, key string, toFile string) (err error) {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
	}
	if err = sdk.Get(secret); err != nil {
		return fmt.Errorf("Unable to extract secret to file: %v", secretName, err)
	}

	value, ok := secret.Data[key]

	// check to see if the map value exists
	if !ok {
		return fmt.Errorf("No secret data \"%s\" found", key)
	}

	if err = ioutil.WriteFile(path.Join(utils.WORKING_DIR, toFile), value, 0644); err != nil {
		return fmt.Errorf("Unable to write to working dir: %v", err)
	}

	return nil
}

func extractMasterCertificate(namespace string, secretName string) (err error) {

	if err = extractSecretToFile(namespace, secretName, "masterca", "ca.crt"); err != nil {
		return
	}

	if err = extractSecretToFile(namespace, secretName, "masterkey", "ca.key"); err != nil {
		return
	}

	return nil
}

func extractKibanaInternalCertificate(namespace string, secretName string) (err error) {

	if err = extractSecretToFile(namespace, secretName, "kibanacert", "kibana-internal.crt"); err != nil {
		return
	}

	if err = extractSecretToFile(namespace, secretName, "kibanakey", "kibana-internal.key"); err != nil {
		return
	}

	return nil
}

func CreateOrUpdateCertificates(logging *v1alpha1.ClusterLogging) (err error) {

	// Pull master signing cert out from secret in logging.Spec.SecretName
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		return fmt.Errorf("Failed to get watch namespace: %v", err)
	}

	if err = extractMasterCertificate(namespace, "logging-master-ca"); err != nil {
		return
	}

	if err = extractKibanaInternalCertificate(namespace, "logging-master-ca"); err != nil {
		return
	}

	cmd := exec.Command("bash", "scripts/cert_generation.sh")
	cmd.Env = append(os.Environ(),
		"NAMESPACE="+namespace,
	)
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("Error running script: %v", err)
	}

	return nil
}
