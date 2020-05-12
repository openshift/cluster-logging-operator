package k8shandler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"html/template"
	"io"
	"strconv"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	esConfig            = "elasticsearch.yml"
	log4jConfig         = "log4j2.properties"
	indexSettingsConfig = "index_settings"
)

// esYmlStruct is used to render esYmlTmpl to a proper elasticsearch.yml format
type esYmlStruct struct {
	KibanaIndexMode       string
	EsUnicastHost         string
	NodeQuorum            string
	RecoverExpectedShards string
}

type log4j2PropertiesStruct struct {
	RootLogger string
}

type indexSettingsStruct struct {
	PrimaryShards string
	ReplicaShards string
}

// CreateOrUpdateConfigMaps ensures the existence of ConfigMaps with Elasticsearch configuration
func (elasticsearchRequest *ElasticsearchRequest) CreateOrUpdateConfigMaps() (err error) {

	dpl := elasticsearchRequest.cluster

	kibanaIndexMode, err := kibanaIndexMode("")
	if err != nil {
		return err
	}
	dataNodeCount := int((getDataCount(dpl)))
	masterNodeCount := int((getMasterCount(dpl)))

	configmap := newConfigMap(
		dpl.Name,
		dpl.Namespace,
		dpl.Labels,
		kibanaIndexMode,
		esUnicastHost(dpl.Name, dpl.Namespace),
		rootLogger(elasticsearchRequest.cluster),
		strconv.Itoa(masterNodeCount/2+1),
		strconv.Itoa(dataNodeCount),
		strconv.Itoa(dataNodeCount),
		strconv.Itoa(calculateReplicaCount(dpl)),
	)

	addOwnerRefToObject(configmap, getOwnerRef(dpl))

	err = elasticsearchRequest.client.Create(context.TODO(), configmap)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing Elasticsearch ConfigMap: %v", err)
		}

		if errors.IsAlreadyExists(err) {
			// Get existing configMap to check if it is same as what we want
			current := configmap.DeepCopy()
			err = elasticsearchRequest.client.Get(context.TODO(), types.NamespacedName{Name: current.Name, Namespace: current.Namespace}, current)
			if err != nil {
				return fmt.Errorf("Unable to get Elasticsearch cluster configMap: %v", err)
			}

			if configMapContentChanged(current, configmap) {
				// Cluster settings has changed, make sure it doesnt go unnoticed
				if err := updateConditionWithRetry(dpl, v1.ConditionTrue, updateUpdatingSettingsCondition, elasticsearchRequest.client); err != nil {
					return err
				}

				return retry.RetryOnConflict(retry.DefaultRetry, func() error {
					if getErr := elasticsearchRequest.client.Get(context.TODO(), types.NamespacedName{Name: current.Name, Namespace: current.Namespace}, current); getErr != nil {
						logrus.Debugf("Could not get Elasticsearch configmap %v: %v", configmap.Name, getErr)
						return getErr
					}

					current.Data = configmap.Data
					if updateErr := elasticsearchRequest.client.Update(context.TODO(), current); updateErr != nil {
						logrus.Debugf("Failed to update Elasticsearch configmap %v: %v", configmap.Name, updateErr)
						return updateErr
					}
					return nil
				})
			} else {
				if err := updateConditionWithRetry(dpl, v1.ConditionFalse, updateUpdatingSettingsCondition, elasticsearchRequest.client); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func renderData(kibanaIndexMode, esUnicastHost, nodeQuorum, recoverExpectedShards, primaryShardsCount, replicaShardsCount, rootLogger string) (error, map[string]string) {

	data := map[string]string{}
	buf := &bytes.Buffer{}
	if err := renderEsYml(buf, kibanaIndexMode, esUnicastHost, nodeQuorum, recoverExpectedShards); err != nil {
		return err, data
	}
	data[esConfig] = buf.String()

	buf = &bytes.Buffer{}
	if err := renderLog4j2Properties(buf, rootLogger); err != nil {
		return err, data
	}
	data[log4jConfig] = buf.String()

	buf = &bytes.Buffer{}
	if err := renderIndexSettings(buf, primaryShardsCount, replicaShardsCount); err != nil {
		return err, data
	}
	data[indexSettingsConfig] = buf.String()

	return nil, data
}

// newConfigMap returns a v1.ConfigMap object
func newConfigMap(configMapName, namespace string, labels map[string]string,
	kibanaIndexMode, esUnicastHost, rootLogger, nodeQuorum, recoverExpectedShards, primaryShardsCount, replicaShardsCount string) *v1.ConfigMap {

	err, data := renderData(kibanaIndexMode, esUnicastHost, nodeQuorum, recoverExpectedShards, primaryShardsCount, replicaShardsCount, rootLogger)
	if err != nil {
		return nil
	}

	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
			Labels:    labels,
		},
		Data: data,
	}
}

func configMapContentChanged(old, new *v1.ConfigMap) bool {
	oldEsConfigSum := sha256.Sum256([]byte(old.Data[esConfig]))
	newEsConfigSum := sha256.Sum256([]byte(new.Data[esConfig]))

	if oldEsConfigSum != newEsConfigSum {
		return true
	}

	oldLog4jConfig := sha256.Sum256([]byte(old.Data[log4jConfig]))
	newLog4jConfig := sha256.Sum256([]byte(new.Data[log4jConfig]))

	if oldLog4jConfig != newLog4jConfig {
		return true
	}

	oldIndexSettingsConfig := sha256.Sum256([]byte(old.Data[indexSettingsConfig]))
	newIndexSettingsConfig := sha256.Sum256([]byte(new.Data[indexSettingsConfig]))

	if oldIndexSettingsConfig != newIndexSettingsConfig {
		return true
	}

	return false
}

func renderEsYml(w io.Writer, kibanaIndexMode, esUnicastHost, nodeQuorum, recoverExpectedShards string) error {
	t := template.New("elasticsearch.yml")
	config := esYmlTmpl
	t, err := t.Parse(config)
	if err != nil {
		return err
	}
	esy := esYmlStruct{
		KibanaIndexMode:       kibanaIndexMode,
		EsUnicastHost:         esUnicastHost,
		NodeQuorum:            nodeQuorum,
		RecoverExpectedShards: recoverExpectedShards,
	}

	return t.Execute(w, esy)
}

func renderLog4j2Properties(w io.Writer, rootLogger string) error {
	t := template.New("log4j2.properties")
	t, err := t.Parse(log4j2PropertiesTmpl)
	if err != nil {
		return err
	}

	log4jProp := log4j2PropertiesStruct{
		RootLogger: rootLogger,
	}

	return t.Execute(w, log4jProp)
}

func renderIndexSettings(w io.Writer, primaryShardsCount, replicaShardsCount string) error {
	t := template.New("index_settings")
	t, err := t.Parse(indexSettingsTmpl)
	if err != nil {
		return err
	}

	indexSettings := indexSettingsStruct{
		PrimaryShards: primaryShardsCount,
		ReplicaShards: replicaShardsCount,
	}

	return t.Execute(w, indexSettings)
}

func getConfigmap(configmapName, namespace string, client client.Client) *v1.ConfigMap {

	configMap := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmapName,
			Namespace: namespace,
		},
	}

	err := client.Get(context.TODO(), types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, &configMap)

	if err != nil {
		// check if doesn't exist
	}

	return &configMap
}

func getConfigmapDataHash(configmapName, namespace string, client client.Client) string {

	hash := ""

	configMap := getConfigmap(configmapName, namespace, client)

	dataHashes := make(map[string][32]byte)

	for key, data := range configMap.Data {
		if key != "index_settings" {
			dataHashes[key] = sha256.Sum256([]byte(data))
		}
	}

	sortedKeys := sortDataHashKeys(dataHashes)

	for _, key := range sortedKeys {
		hash = fmt.Sprintf("%s%s", hash, dataHashes[key])
	}

	return hash
}
