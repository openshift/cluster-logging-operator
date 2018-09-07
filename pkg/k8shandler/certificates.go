package k8shandler

import (
  "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
  "github.com/sirupsen/logrus"
  "k8s.io/api/core/v1"
  "os"
  "os/exec"
  "io/ioutil"

  sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
  k8sutil "github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func extractMasterCertificate(namespace string, secretName string)(error) {

  secret := &v1.Secret{
    TypeMeta: metav1.TypeMeta{
      Kind: "Secret",
      APIVersion: "v1",
    },
    ObjectMeta: metav1.ObjectMeta{
      Name: secretName,
      Namespace: namespace,
    },
  }
  err := sdk.Get(secret)

  // value []byte
  caValue, ok := secret.Data["masterca"]

  // check to see if the map value exists
  if ! ok {
    logrus.Fatalf("No secret data \"masterca\" found")
    return nil
  }

  err = ioutil.WriteFile("/tmp/_working_dir/ca.crt", caValue, 0644)
  if err != nil {
    logrus.Fatalf("Unable to write CA cert to working dir: %v", err)
  }

  keyValue, ok := secret.Data["masterkey"]

  // check to see if the map value exists
  if ! ok {
    logrus.Fatalf("No secret data \"masterkey\" found")
    return nil
  }

  err = ioutil.WriteFile("/tmp/_working_dir/ca.key", keyValue, 0644)
  if err != nil {
    logrus.Fatalf("Unable to write CA key to working dir: %v", err)
  }

  return nil
}

func CreateOrUpdateCertificates(logging *v1alpha1.ClusterLogging)(error) {

  // Pull master signing cert out from secret in logging.Spec.SecretName
  namespace, err := k8sutil.GetWatchNamespace()
  if err != nil {
    logrus.Fatalf("Failed to get watch namespace: %v", err)
  }

  err = extractMasterCertificate(namespace, "logging-master-ca")

  cmd := exec.Command("bash", "scripts/cert_generation.sh")
  cmd.Env = append(os.Environ(),
              "NAMESPACE="+namespace,
            )
  if err = cmd.Run(); err != nil {
    logrus.Fatalf("Error running script: %v", err)
  }

  return nil
}
