package k8shandler

import (
  "github.com/ViaQ/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
  "k8s.io/api/core/v1"
  "k8s.io/apimachinery/pkg/api/errors"
  "github.com/sirupsen/logrus"

  "bytes"
  "math/rand"
  "encoding/base64"
  "time"
  "io/ioutil"

  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
  logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
)

func CreateOrUpdateLogStore(logging *logging.ClusterLogging) error {
  CreateOrUpdateSecret(logging)
  return CreateOrUpdateCR(logging)
}

func getFileContents(filePath string) []byte {
  contents, err := ioutil.ReadFile(filePath)
  if err != nil {
    logrus.Fatalf("Unable to read file to get contents: %v", err)
  }

  return contents
}

func init() {
  rand.Seed(time.Now().UnixNano())
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func getRandomWord(wordSize int) string {
  b := make([]rune, wordSize)
  for i := range b {
    b[i] = letters[rand.Intn(len(letters))]
  }
  return string(b)
}

func createPasswdFile(user_name string, user_passwd string) []byte {
  var buffer bytes.Buffer

  buffer.WriteString(user_name)
  buffer.WriteString(":\n  passwd: ")
  buffer.WriteString(base64.StdEncoding.EncodeToString([]byte(user_passwd)))

  return buffer.Bytes()
}

func CreateOrUpdateSecret(logging *logging.ClusterLogging) error {

  esSecret := &v1.Secret{
    TypeMeta: metav1.TypeMeta{
      Kind: "Secret",
      APIVersion: "v1",
    },
    ObjectMeta: metav1.ObjectMeta{
      Name: "logging-elasticsearch",
      Namespace: logging.Namespace,
    },
    Type: "Opaque",
    Data: map[string][]byte{
      "elasticsearch.key": getFileContents("/tmp/_working_dir/elasticsearch.key"),
      "elasticsearch.crt": getFileContents("/tmp/_working_dir/elasticsearch.crt"),
      "logging-es.key": getFileContents("/tmp/_working_dir/logging-es.key"),
      "logging-es.crt": getFileContents("/tmp/_working_dir/logging-es.crt"),
      "admin-key": getFileContents("/tmp/_working_dir/system.admin.key"),
      "admin-cert": getFileContents("/tmp/_working_dir/system.admin.crt"),
      "admin-ca": getFileContents("/tmp/_working_dir/ca.crt"),
      "passwd.yml": createPasswdFile("prometheus", getRandomWord(16)),
    },
  }

  err := sdk.Create(esSecret)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure constructing Elasticsearch secret: %v", err)
  }

  return nil
}

func CreateOrUpdateCR(logging *logging.ClusterLogging) error {

  var esNodes []v1alpha1.ElasticsearchNode

  esNode := v1alpha1.ElasticsearchNode{
    Roles: []v1alpha1.ElasticsearchNodeRole{"client", "data", "master"},
    Replicas: logging.Spec.LogStore.Replicas,
    NodeSelector: logging.Spec.LogStore.NodeSelector,
    Spec: v1alpha1.ElasticsearchNodeSpec{
      Resources: logging.Spec.LogStore.Resources,
    },
    Storage: v1alpha1.ElasticsearchNodeStorageSource{
      VolumeClaimTemplate: &v1.PersistentVolumeClaim{},//TODO: fill out,
    },
  }

  // build Nodes
  esNodes = append(esNodes, esNode)

  esCR := &v1alpha1.Elasticsearch{
    ObjectMeta: metav1.ObjectMeta{
      Name: "logging-es",
      Namespace: logging.Namespace,
    },
    TypeMeta: metav1.TypeMeta{
      Kind: "Elasticsearch",
      APIVersion: "elasticsearch.redhat.com/v1alpha1",
    },
    Spec: v1alpha1.ElasticsearchSpec{
      Nodes: esNodes,
      Secure: v1alpha1.ElasticsearchSecure{
        Disabled: false,
        CertificatesSecret: "logging-elasticsearch",
      },
      ServiceAccountName: "aggregated-logging-elasticsearch",
      ConfigMapName: "logging-elasticsearch",
    },
  }

  logrus.Infof("Created Elasticsearch struct: %v", esCR)

  err := sdk.Create(esCR)
  if err != nil && !errors.IsAlreadyExists(err) {
    logrus.Fatalf("Failure constructing Elasticsearch CR: %v", err)
  } else if errors.IsAlreadyExists(err) {
    // Get existing configMap to check if it is same as what we want
    existingCR := &v1alpha1.Elasticsearch{
      ObjectMeta: metav1.ObjectMeta{
        Name: "logging-es",
        Namespace: logging.Namespace,
      },
      TypeMeta: metav1.TypeMeta{
        Kind: "Elasticsearch",
        APIVersion: "elasticsearch.redhat.com/v1alpha1",
      },
    }

    err = sdk.Get(existingCR)
    if err != nil {
      logrus.Fatalf("Unable to get Elasticsearch CR: %v", err)
    }

    logrus.Infof("Found existing CR: %v", existingCR)

    // TODO: Compare existing CR labels, selectors and port
  }

  return nil
}
