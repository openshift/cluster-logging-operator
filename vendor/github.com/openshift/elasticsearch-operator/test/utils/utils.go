package utils

import (
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	"github.com/openshift/elasticsearch-operator/pkg/k8shandler"
	"github.com/openshift/elasticsearch-operator/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	goctx "context"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetFileContents(filePath string) []byte {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		logrus.Errorf("Unable to read file to get contents: %v", err)
		return nil
	}

	return contents
}

func Secret(secretName string, namespace string, data map[string][]byte) *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Type: "Opaque",
		Data: data,
	}
}

func WaitForNodeStatusCondition(t *testing.T, f *framework.Framework, namespace, name string, condition api.ElasticsearchNodeUpgradeStatus, retryInterval, timeout time.Duration) error {
	elasticsearchCR := &api.Elasticsearch{}
	elasticsearchName := types.NamespacedName{Name: name, Namespace: namespace}

	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		err = f.Client.Get(goctx.TODO(), elasticsearchName, elasticsearchCR)
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s elasticsearch\n", name)
				return false, nil
			}
			return false, err
		}

		allMatch := true

		for _, node := range elasticsearchCR.Status.Nodes {
			if !reflect.DeepEqual(node.UpgradeStatus, condition) {
				allMatch = false
			}
		}

		if allMatch {
			return true, nil
		}
		t.Logf("Waiting for full condition match of %s elasticsearch\n", name)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Full condition matches\n")
	return nil
}

func WaitForClusterStatusCondition(t *testing.T, f *framework.Framework, namespace, name string, condition api.ClusterCondition, retryInterval, timeout time.Duration) error {
	elasticsearchCR := &api.Elasticsearch{}
	elasticsearchName := types.NamespacedName{Name: name, Namespace: namespace}

	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		err = f.Client.Get(goctx.TODO(), elasticsearchName, elasticsearchCR)
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s elasticsearch\n", name)
				return false, nil
			}
			return false, err
		}

		contained := false

		for _, clusterCondition := range elasticsearchCR.Status.Conditions {
			if reflect.DeepEqual(clusterCondition, condition) {
				contained = true
			}
		}

		if contained {
			return true, nil
		}
		t.Logf("Waiting for full condition match of %s elasticsearch\n", name)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Full condition matches\n")
	return nil
}

func WaitForStatefulset(t *testing.T, kubeclient kubernetes.Interface, namespace, name string, replicas int, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		statefulset, err := kubeclient.AppsV1().StatefulSets(namespace).Get(name, metav1.GetOptions{IncludeUninitialized: true})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s statefulset\n", name)
				return false, nil
			}
			return false, err
		}

		if int(statefulset.Status.ReadyReplicas) == replicas {
			return true, nil
		}
		t.Logf("Waiting for full availability of %s statefulset (%d/%d)\n", name, statefulset.Status.ReadyReplicas, replicas)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Statefulset available (%d/%d)\n", replicas, replicas)
	return nil
}

func GenerateUUID() string {

	uuid, err := utils.RandStringBytes(8)
	if err != nil {
		return ""
	}

	return uuid
}

func WaitForIndexTemplateReplicas(t *testing.T, kubeclient kubernetes.Interface, namespace, clusterName string, replicas int32, retryInterval, timeout time.Duration) error {
	// mock out Secret response from client
	mockClient := fake.NewFakeClient(getMockedSecret(clusterName, namespace))

	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		// get all index replica count
		indexTemplates, err := k8shandler.GetIndexTemplates(clusterName, namespace, mockClient)
		if err != nil {
			t.Logf("Received error: %v", err)
			return false, nil
		}

		// for each index -- check replica count
		for templateName, template := range indexTemplates {
			if numberOfReplicas := parseString("settings.index.number_of_replicas", template.(map[string]interface{})); numberOfReplicas != "" {
				currentReplicas, err := strconv.ParseInt(numberOfReplicas, 10, 32)
				if err != nil {
					return false, err
				}

				if int32(currentReplicas) == replicas {
					continue
				}

				t.Logf("Index template %s did not have correct replica count (%d/%d)", templateName, currentReplicas, replicas)
				return false, nil
			} else {
				return false, nil
			}
		}

		return true, nil
	})
	if err != nil {
		return err
	}
	t.Logf("All index templates have correct replica count of %d\n", replicas)
	return nil
}

func WaitForIndexReplicas(t *testing.T, kubeclient kubernetes.Interface, namespace, clusterName string, replicas int32, retryInterval, timeout time.Duration) error {
	// mock out Secret response from client
	mockClient := fake.NewFakeClient(getMockedSecret(clusterName, namespace))

	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {

		// get all index replica count
		indexHealth, err := k8shandler.GetIndexReplicaCounts(clusterName, namespace, mockClient)
		if err != nil {
			return false, nil
		}

		// for each index -- check replica count
		for index, health := range indexHealth {
			if numberOfReplicas := parseString("settings.index.number_of_replicas", health.(map[string]interface{})); numberOfReplicas != "" {
				currentReplicas, err := strconv.ParseInt(numberOfReplicas, 10, 32)
				if err != nil {
					return false, err
				}

				if int32(currentReplicas) == replicas {
					continue
				}

				t.Logf("Index %s did not have correct replica count (%d/%d)", index, currentReplicas, replicas)
				return false, nil
			} else {
				return false, nil
			}
		}

		return true, nil
	})
	if err != nil {
		return err
	}
	t.Logf("All indices have correct replica count of %d\n", replicas)
	return nil
}

func getMockedSecret(clusterName, namespace string) *v1.Secret {
	return Secret(
		clusterName,
		namespace,
		map[string][]byte{
			"elasticsearch.key": GetFileContents("test/files/elasticsearch.key"),
			"elasticsearch.crt": GetFileContents("test/files/elasticsearch.crt"),
			"logging-es.key":    GetFileContents("test/files/logging-es.key"),
			"logging-es.crt":    GetFileContents("test/files/logging-es.crt"),
			"admin-key":         GetFileContents("test/files/system.admin.key"),
			"admin-cert":        GetFileContents("test/files/system.admin.crt"),
			"admin-ca":          GetFileContents("test/files/ca.crt"),
		},
	)
}

func parseString(path string, interfaceMap map[string]interface{}) string {
	value := walkInterfaceMap(path, interfaceMap)

	if parsedString, ok := value.(string); ok {
		return parsedString
	} else {
		return ""
	}
}

func walkInterfaceMap(path string, interfaceMap map[string]interface{}) interface{} {

	current := interfaceMap
	keys := strings.Split(path, ".")
	keyCount := len(keys)

	for index, key := range keys {
		if current[key] != nil {
			if index+1 < keyCount {
				current = current[key].(map[string]interface{})
			} else {
				return current[key]
			}
		} else {
			return nil
		}
	}

	return nil
}
