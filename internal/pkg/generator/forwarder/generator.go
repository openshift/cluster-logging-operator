package forwarder

import (
	"errors"
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/migrations"

	forwardergenerator "github.com/openshift/cluster-logging-operator/internal/generator/forwarder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	log "github.com/ViaQ/logerr/v2/log/static"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
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
		ResourceNames: factory.GenerateResourceNames(*forwarder),
	}

	if client != nil {
		clRequest.Client = client
	}

	if includeDefaultLogStore {
		clRequest.Cluster.Spec.LogStore = &logging.LogStoreSpec{
			Type: logging.LogStoreTypeElasticsearch,
		}
	}

	mSpec, extras, condition := migrations.MigrateClusterLogForwarder(forwarder.Namespace, forwarder.Name, forwarder.Spec, clRequest.Cluster.Spec.LogStore, map[string]bool{}, "", "")
	log.V(0).Info("Migrated ClusterLogForwarder", "spec", mSpec, "extras", extras, "condition", condition)
	// Set the output secrets if any
	clRequest.SetOutputSecrets()
	forwarder.Spec = mSpec
	tunings := &logging.FluentdForwarderSpec{}
	clspec := logging.CollectionSpec{
		Fluentd: tunings,
	}
	op := framework.Options{}
	k8shandler.EvaluateAnnotationsForEnabledCapabilities(forwarder, op)
	op[framework.ClusterTLSProfileSpec] = tls.GetClusterTLSProfileSpec(nil)

	configGenerator := forwardergenerator.New(collectionType)
	if configGenerator == nil {
		return "", errors.New("unsupported collector implementation")
	}
	return configGenerator.GenerateConf(&clspec, clRequest.OutputSecrets, &forwarder.Spec, clRequest.Cluster.Namespace, forwarder.Name, clRequest.ResourceNames, op)
}
