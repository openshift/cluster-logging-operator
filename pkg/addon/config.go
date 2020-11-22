package addon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/pkg/apis"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const addonSecretName = "addon-cluster-logging-operator-parameters"

var ErrNoAddonConfig = errors.New("No addon configiration found")

type AddonConfiguration struct {
	EsMaxRetention string `data-field:"es-max-retention" default:"7d"`
	EsMemory       string `data-field:"es-memory" default:"2Gi"`
	EsNodeCount    string `data-field:"es-node-count" default:"3"`
	EsStorageClass string `data-field:"es-storage-class" default:"gp2"`
	EsStorageSize  string `data-field:"es-storage-size" default:"200G"`
}

func ProcessAddonConfiguration(namespace string) {
	// Verify if CLO's CR already present in the target namespace
	cl, err := newClient()
	if err != nil {
		log.Error(err, "failed to create client")
		return
	}

	present, err := verifyIfCrPresent(cl, namespace)
	if err != nil {
		log.Error(err, "unable to verify if clo's cr is present")
		return
	}

	if present {
		log.Info("clo's cr detected. Skipping addon configuration")
		return
	}

	// Read addon secret from the target namespace
	addonConfig, err := readAddonConfiguration(cl, namespace, addonSecretName)
	if err != nil {
		if err == ErrNoAddonConfig {
			log.Info("no addon config found. skipping cr generation")
			return
		} else {
			log.Error(err, "unable to read addon configuration")
			return
		}
	}

	// Generate new CLO's CR with parameters from the addon secret
	cloCr, err := generateCrFromAddonConfig(addonConfig, namespace)
	if err != nil {
		log.Error(err, "unable to generate cr from addon config")
		return
	}

	// Deploy the generated CR
	err = cl.Create(context.Background(), cloCr)
	if err != nil {
		log.Error(err, "unable to create cr for addon config")
	}
	fmt.Println("addon configuration processed")
}

func verifyIfCrPresent(cl client.Client, namespace string) (bool, error) {
	lg := &logging.ClusterLogging{}
	err := cl.Get(context.Background(), client.ObjectKey{Name: "instance", Namespace: namespace}, lg)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func generateCrFromAddonConfig(addonConfig *AddonConfiguration, namespace string) (*logging.ClusterLogging, error) {
	nodeCount, err := strconv.Atoi(addonConfig.EsNodeCount)
	if err != nil {
		return nil, fmt.Errorf("incorrect es-node-count parameter. %v", err)
	}
	mpEsStorage := resource.MustParse(addonConfig.EsStorageSize)

	result := &logging.ClusterLogging{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance",
			Namespace: namespace,
		},
		Spec: logging.ClusterLoggingSpec{
			ManagementState: "Managed",
			Visualization: &logging.VisualizationSpec{
				Type: "kibana",
				KibanaSpec: logging.KibanaSpec{
					Replicas: 1,
				},
			},
			LogStore: &logging.LogStoreSpec{
				Type: "elasticsearch",
				ElasticsearchSpec: logging.ElasticsearchSpec{
					Resources: &v1.ResourceRequirements{
						Requests: map[v1.ResourceName]resource.Quantity{
							"memory": resource.MustParse(addonConfig.EsMemory),
						},
					},
					NodeCount: int32(nodeCount),
					Storage: elasticsearch.ElasticsearchStorageSpec{
						StorageClassName: &addonConfig.EsStorageClass,
						Size:             &mpEsStorage,
					},
					RedundancyPolicy: "SingleRedundancy",
				},
				RetentionPolicy: &logging.RetentionPoliciesSpec{
					App: &logging.RetentionPolicySpec{
						MaxAge: elasticsearch.TimeUnit(addonConfig.EsMaxRetention),
					},
					Infra: &logging.RetentionPolicySpec{
						MaxAge: elasticsearch.TimeUnit(addonConfig.EsMaxRetention),
					},
					Audit: &logging.RetentionPolicySpec{
						MaxAge: elasticsearch.TimeUnit(addonConfig.EsMaxRetention),
					},
				},
			},
			Collection: &logging.CollectionSpec{
				Logs: logging.LogCollectionSpec{
					Type: "fluentd",
				},
			},
			Curation: &logging.CurationSpec{
				Type: "curator",
				CuratorSpec: logging.CuratorSpec{
					Schedule: "30 3 * * *",
				},
			},
		},
	}
	return result, nil
}

func readAddonConfiguration(cl client.Client, namespace, secretName string) (*AddonConfiguration, error) {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{},
	}
	err := cl.Get(context.Background(), client.ObjectKey{
		Name:      secretName,
		Namespace: namespace,
	}, secret)

	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, ErrNoAddonConfig
		} else {
			return nil, fmt.Errorf("unable to get addon-secret. %v", err)
		}
	}

	result := &AddonConfiguration{}
	sData := secret.Data

	v := reflect.ValueOf(result).Elem()
	t := reflect.TypeOf(*result)
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := t.Field(i)
		tTag := fieldType.Tag
		if iv, ok := sData[tTag.Get("data-field")]; !ok || len(iv) <= 0 {
			defaultValue := tTag.Get("default")
			fieldValue.SetString(defaultValue)
		} else {
			fieldValue.SetString(string(iv))
		}
	}
	return result, nil
}

func newClient() (client.Client, error) {
	config, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	shm := scheme.Scheme

	if err := apis.AddToScheme(shm); err != nil {
		log.Error(err, "failed to add resources to scheme", "resource", "apis")
		os.Exit(1)
	}

	cl, err := client.New(config, client.Options{
		Scheme: shm,
	})

	if err != nil {
		return nil, err
	}
	return cl, nil
}
