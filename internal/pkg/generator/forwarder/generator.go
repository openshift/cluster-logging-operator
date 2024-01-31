package forwarder

import (
	"errors"
	"fmt"

	forwardergenerator "github.com/openshift/cluster-logging-operator/internal/generator/forwarder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	log "github.com/ViaQ/logerr/v2/log/static"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	"github.com/openshift/cluster-logging-operator/internal/tls"
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

func Generate(collectionType logging.LogCollectionType, clfYaml string, includeDefaultLogStore, debugOutput bool, client client.Client) (string, error) {
	var err error
	forwarder, err := UnMarshalClusterLogForwarder(clfYaml)
	if err != nil {
		return "", fmt.Errorf("Error Unmarshalling %q: %v", clfYaml, err)
	}
	log.V(2).Info("Unmarshalled", "forwarder", forwarder)

	clRequest := &k8shandler.ClusterLoggingRequest{
		Forwarder: forwarder,
		Cluster: &logging.ClusterLogging{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: forwarder.GetNamespace(),
			},
			Spec: logging.ClusterLoggingSpec{},
		},
	}

	if client != nil {
		clRequest.Client = client
	}

	if includeDefaultLogStore {
		clRequest.Cluster.Spec.LogStore = &logging.LogStoreSpec{
			Type: logging.LogStoreTypeElasticsearch,
		}
	}

	// Added because this generator is used for tests and the tests assume a correct
	// CLF spec.
	// Originally, this generator normalized the forwarder spec only which dropped
	// input specs with names that were reserved (application, audit, infrastructure).
	// With the change in validation, CLFs are rejected outright with input specs name equal to reserved names.
	sanitizedInputSpec := []logging.InputSpec{}
	for _, inputSpec := range forwarder.Spec.Inputs {
		if inputSpec.Name != logging.InputNameApplication &&
			inputSpec.Name != logging.InputNameAudit &&
			inputSpec.Name != logging.InputNameInfrastructure {
			sanitizedInputSpec = append(sanitizedInputSpec, inputSpec)
		}
	}

	forwarder.Spec.Inputs = sanitizedInputSpec

	// Set the output secrets if any
	clRequest.SetOutputSecrets()

	tunings := &logging.FluentdForwarderSpec{}
	clspec := logging.CollectionSpec{
		Fluentd: tunings,
	}
	op := generator.Options{}
	k8shandler.EvaluateAnnotationsForEnabledCapabilities(forwarder, op)
	op[generator.ClusterTLSProfileSpec] = tls.GetClusterTLSProfileSpec(nil)

	configGenerator := forwardergenerator.New(collectionType)
	if configGenerator == nil {
		return "", errors.New("unsupported collector implementation")
	}
	return configGenerator.GenerateConf(&clspec, clRequest.OutputSecrets, &forwarder.Spec, clRequest.Cluster.Namespace, forwarder.Name, op)
}
