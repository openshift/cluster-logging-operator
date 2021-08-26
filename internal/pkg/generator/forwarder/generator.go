package forwarder

import (
	"errors"
	"fmt"
	fluentd2 "github.com/openshift/cluster-logging-operator/internal/generator/fluentd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/ViaQ/logerr/log"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
)

const (
	//these are fixed at the moment
	logCollectorType         = logging.LogCollectionTypeFluentd
	includeLegacySyslog      = false
	useOldRemoteSyslogPlugin = false
)

func Generate(clfYaml string, includeDefaultLogStore, includeLegacyForward, debugOutput bool, client *client.Client) (string, error) {

	var err error
	g := generator.MakeGenerator()
	op := generator.Options{}
	if includeLegacyForward {
		op[generator.IncludeLegacyForwardConfig] = ""
	}
	if includeLegacySyslog {
		op[generator.IncludeLegacySyslogConfig] = ""
	}
	if useOldRemoteSyslogPlugin {
		op[generator.UseOldRemoteSyslogPlugin] = ""
	}
	if debugOutput {
		op["debug_output"] = ""
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
			ObjectMeta: metav1.ObjectMeta{
				Namespace: forwarder.GetNamespace(),
			},
			Spec: logging.ClusterLoggingSpec{},
		},
		FnIncludeLegacyForward: func() bool { return includeLegacyForward },
		FnIncludeLegacySyslog:  func() bool { return includeLegacySyslog },
		CLFVerifier: k8shandler.ClusterLogForwarderVerifier{
			VerifyOutputSecret: func(output *logging.OutputSpec, conds logging.NamedConditions) bool { return true },
		},
	}
	if client != nil {
		clRequest.Client = *client
	}
	if includeDefaultLogStore {
		clRequest.Cluster.Spec.LogStore = &logging.LogStoreSpec{
			Type: logging.LogStoreTypeElasticsearch,
		}
	}

	spec, status := clRequest.NormalizeForwarder()
	log.V(2).Info("Normalization", "spec", spec)
	log.V(2).Info("Normalization", "status", status)

	tunings := &logging.ForwarderSpec{}
	clspec := logging.ClusterLoggingSpec{
		Forwarder: tunings,
	}
	if logCollectorType == logging.LogCollectionTypeFluentd {

		sections := fluentd2.Conf(&clspec, nil, spec, op)
		es := generator.MergeSections(sections)

		generatedConfig, err := g.GenerateConf(es...)
		if err != nil {
			return "", fmt.Errorf("Unable to generate log configuration: %v", err)
		}
		return generatedConfig, nil
	} else {
		return "", errors.New("Only fluentd Log Collector supported")
	}
}
