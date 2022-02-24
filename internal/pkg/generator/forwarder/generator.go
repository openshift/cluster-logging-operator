package forwarder

import (
	"errors"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
	"strings"

	"github.com/ViaQ/logerr/log"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
)

const (
	//these are fixed at the moment
	logCollectorType         = logging.LogCollectionTypeFluentd
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

func Generate(clfYaml string, includeDefaultLogStore, debugOutput bool, client client.Client) (string, error) {

	g := generator.MakeGenerator()
	op := generator.Options{}
	if useOldRemoteSyslogPlugin {
		op[generator.UseOldRemoteSyslogPlugin] = ""
	}
	if debugOutput {
		op["debug_output"] = ""
	}

	forwarder, err := UnMarshalClusterLogForwarder(clfYaml)
	if err != nil {
		return "", fmt.Errorf("Error Unmarshalling %q: %v", clfYaml, err)
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
	}

	if client != nil {
		clRequest.Client = client
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

		sections := fluentd.Conf(&clspec, clRequest.OutputSecrets, spec, op)
		es := generator.MergeSections(sections)
		generatedConfig, err := g.GenerateConf(es...)
		if err != nil {
			return "", fmt.Errorf("Unable to generate log configuration: %v", err)
		}
		return formatFluentConf(generatedConfig), nil

	} else {
		return "", errors.New("Only fluentd Log collector supported")
	}
}
func formatFluentConf(conf string) string {
	indent := 0
	lines := strings.Split(conf, "\n")
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(trimmed, "</") && strings.HasSuffix(trimmed, ">"):
			indent--
			trimmed = pad(trimmed, indent)
		case strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">"):
			trimmed = pad(trimmed, indent)
			indent++
		default:
			trimmed = pad(trimmed, indent)
		}
		if len(strings.TrimSpace(trimmed)) == 0 {
			trimmed = ""
		}
		lines[i] = trimmed
	}
	return strings.Join(lines, "\n")
}

func pad(line string, indent int) string {
	prefix := ""
	if indent >= 0 {
		prefix = strings.Repeat("  ", indent)
	}
	return prefix + line
}
