package forwarder

import (
	"errors"
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	forwardergenerator "github.com/openshift/cluster-logging-operator/internal/generator/forwarder"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	log "github.com/ViaQ/logerr/v2/log/static"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/internal/tls"
)

func UnMarshalClusterLogForwarder(clfYaml string) (forwarder *obs.ClusterLogForwarder, err error) {
	forwarder = &obs.ClusterLogForwarder{}
	if clfYaml != "" {
		err = yaml.Unmarshal([]byte(clfYaml), forwarder)
		if err != nil {
			return nil, fmt.Errorf("Error Unmarshalling %q: %v", clfYaml, err)
		}
	}
	return forwarder, err
}

func Generate(clfYaml string, debugOutput bool, client client.Client) (string, error) {
	var err error
	forwarder, err := UnMarshalClusterLogForwarder(clfYaml)
	if err != nil {
		return "", fmt.Errorf("Error Unmarshalling %q: %v", clfYaml, err)
	}
	log.V(2).Info("Unmarshalled", "forwarder", forwarder)

	//clRequest := &k8shandler.ClusterLoggingRequest{
	//	Forwarder: forwarder,
	//	Cluster: &logging.ClusterLogging{
	//		ObjectMeta: metav1.ObjectMeta{
	//			Namespace: forwarder.GetNamespace(),
	//		},
	//		Spec: logging.ClusterLoggingSpec{},
	//	},
	//	ResourceNames: factory.ResourceNames(*forwarder),
	//}

	//if client != nil {
	//	clRequest.Client = client
	//}
	//TODO: Enable migrations?
	//mSpec, extras, condition := migrations.MigrateClusterLogForwarder(forwarder.Namespace, forwarder.Name, forwarder.Spec, clRequest.Cluster.Spec.LogStore, map[string]bool{}, "", "")
	//log.V(0).Info("Migrated ClusterLogForwarder", "spec", mSpec, "extras", extras, "condition", condition)
	// forwarder.Spec = mSpec
	// Set the output secrets if any
	// TODO: enable secrets
	//secrets := internalobs.FetchSecrets(forwarder.Spec.Outputs, client)
	secrets := map[string]*corev1.Secret{}

	op := framework.Options{}
	//k8shandler.EvaluateAnnotationsForEnabledCapabilities(forwarder.Annotations, op)
	op[framework.ClusterTLSProfileSpec] = tls.GetClusterTLSProfileSpec(nil)

	configGenerator := forwardergenerator.New()
	if configGenerator == nil {
		return "", errors.New("unsupported collector implementation")
	}
	return configGenerator.GenerateConf(secrets, forwarder.Spec, forwarder.Namespace, forwarder.Name, *factory.ResourceNames(*forwarder), op)
}
