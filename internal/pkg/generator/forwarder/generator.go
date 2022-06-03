package forwarder

import (
	"errors"
	"fmt"

	forwardergenerator "github.com/openshift/cluster-logging-operator/internal/generator/forwarder"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/ViaQ/logerr/v2/log"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
)

const (
	useOldRemoteSyslogPlugin = false
)

func UnMarshalClusterLogForwarder(clfYaml string) (forwarder *logging.ClusterLogForwarder, err error) {
	forwarder = &logging.ClusterLogForwarder{}
	if clfYaml != "" {
		err = yaml.Unmarshal([]byte(clfYaml), forwarder)
		if err != nil {
			return nil, fmt.Errorf("Error Unmarshalling %q: %v", clfYaml, err)
		}
	}
	return forwarder, err
}

func Generate(collectionType logging.LogCollectionType, clfYaml string, includeDefaultLogStore, debugOutput bool, client *client.Client) (string, error) {

	logger := log.NewLogger("k8sHandler")
	var err error
	forwarder, err := UnMarshalClusterLogForwarder(clfYaml)
	if err != nil {
		return "", fmt.Errorf("Error Unmarshalling %q: %v", clfYaml, err)
	}
	logger.V(2).Info("Unmarshalled", "forwarder", forwarder)

	clRequest := &k8shandler.ClusterLoggingRequest{
		ForwarderSpec: forwarder.Spec,
		Cluster: &logging.ClusterLogging{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: forwarder.GetNamespace(),
			},
			Spec: logging.ClusterLoggingSpec{},
		},
	}

	if client != nil {
		clRequest.Client = *client
		clRequest.Log = log.NewLogger("k8sHandler")
	}

	if includeDefaultLogStore {
		clRequest.Cluster.Spec.LogStore = &logging.LogStoreSpec{
			Type: logging.LogStoreTypeElasticsearch,
		}
	}

	spec, status := clRequest.NormalizeForwarder()
	logger.V(2).Info("Normalization", "spec", spec)
	logger.V(2).Info("Normalization", "status", status)

	tunings := &logging.ForwarderSpec{}
	clspec := logging.ClusterLoggingSpec{
		Forwarder: tunings,
	}
	op := generator.Options{}
	if useOldRemoteSyslogPlugin {
		op[generator.UseOldRemoteSyslogPlugin] = ""
	}
	if debugOutput {
		op[helpers.EnableDebugOutput] = ""
	}
	configGenerator := forwardergenerator.New(collectionType)
	if configGenerator == nil {
		return "", errors.New("unsupported collector implementation")
	}
	return configGenerator.GenerateConf(&clspec, clRequest.OutputSecrets, spec, op)
}
