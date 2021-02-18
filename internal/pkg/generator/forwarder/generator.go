package forwarder

import (
	"fmt"

	yaml "sigs.k8s.io/yaml"

	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
)

const (
	//these are fixed at the moment
	logCollectorType         = logging.LogCollectionTypeFluentd
	includeLegacySyslog      = false
	useOldRemoteSyslogPlugin = false
)

func Generate(clfYaml string, includeDefaultLogStore, includeLegacyForward bool) (string, error) {

	generator, err := forwarding.NewConfigGenerator(
		logCollectorType,
		includeLegacyForward,
		includeLegacySyslog,
		useOldRemoteSyslogPlugin)
	if err != nil {
		return "", fmt.Errorf("Unable to create collector config generator: %v", err)
	}

	forwarder := &logging.ClusterLogForwarder{}
	if clfYaml != "" {
		err = yaml.Unmarshal([]byte(clfYaml), forwarder)
		if err != nil {
			return "", fmt.Errorf("Error Unmarshalling %q: %v", clfYaml, err)
		}
	}
	log.V(2).Info("Unmarshalled", "forwarder", forwarder)
	clRequest := &k8shandler.ClusterLoggingRequest{
		ForwarderSpec: forwarder.Spec,
		Cluster: &logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{},
		},
		CLFVerifier: k8shandler.ClusterLogForwarderVerifier{
			VerifyOutputSecret: func(output *logging.OutputSpec, conds logging.NamedConditions) bool { return true },
		},
	}
	if includeDefaultLogStore {
		clRequest.Cluster.Spec.LogStore = &logging.LogStoreSpec{
			Type: logging.LogStoreTypeElasticsearch,
		}
	}

	spec, status := clRequest.NormalizeForwarder()
	log.V(2).Info("Normalization", "status", status)

	tunings := &logging.ForwarderSpec{}

	generatedConfig, err := generator.Generate(spec, nil, tunings)
	if err != nil {
		return "", fmt.Errorf("Unable to generate log configuration: %v", err)
	}
	return generatedConfig, nil
}
