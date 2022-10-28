package functional

import (
	"strconv"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/url"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

const (
	ElasticSearchTag   = "7.10.1"
	ElasticSearchImage = "elasticsearch:" + ElasticSearchTag
)

func (f *CollectorFunctionalFramework) AddES7Output(b *runtime.PodBuilder, output logging.OutputSpec) error {
	log.V(2).Info("Adding elasticsearch7 output", "name", output.Name)
	name := strings.ToLower(output.Name)

	esURL, err := url.Parse(output.URL)
	if err != nil {
		return err
	}
	transportPort, portError := determineTransportPort(esURL.Port())
	if portError != nil {
		return portError
	}

	log.V(2).Info("Adding container", "name", name)
	log.V(2).Info("Adding ElasticSearch output container", "name", logging.OutputTypeElasticsearch)
	b.AddContainer(name, ElasticSearchImage).
		AddEnvVar("discovery.type", "single-node").
		AddEnvVar("http.port", esURL.Port()).
		AddEnvVar("transport.port", transportPort).
		AddRunAsUser(2000).
		End()
	return nil
}

func determineTransportPort(port string) (string, error) {
	iPort, err := strconv.Atoi(port)
	if err != nil {
		return "", err
	}
	iPort = iPort + 100
	return strconv.Itoa(iPort), nil
}
