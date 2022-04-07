package functional

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	elastichelper "github.com/openshift/cluster-logging-operator/test/helpers/elasticsearch"
)

var ElasticIndex = map[string]string{
	logging.InputNameApplication:    "app-write",
	logging.InputNameInfrastructure: "app-infra",
}

func (f *CollectorFunctionalFramework) GetLogsFromElasticSearch(outputName string, outputLogType string) (result string, err error) {
	index, ok := ElasticIndex[outputLogType]
	if !ok {
		return "", fmt.Errorf(fmt.Sprintf("can't find log of type %s", outputLogType))
	}
	return f.GetLogsFromElasticSearchIndex(outputName, index)
}

func (f *CollectorFunctionalFramework) GetLogsFromElasticSearchIndex(outputName string, index string) (result string, err error) {

	buffer := []string{}
	hits, err := elastichelper.GetDocsFromElasticSearch(f.Test.NS.Name, f.Pod, outputName, index, false)
	if err != nil {
		return "", err
	}
	for i := 0; i < len(hits); i++ {
		hit := hits[i].(map[string]interface{})
		jsonHit, err := json.Marshal(hit["_source"])
		if err == nil {
			buffer = append(buffer, string(jsonHit))
		} else {
			log.V(4).Info("Marshall failed", "err", err)
		}
	}
	result = fmt.Sprintf("[%s]", strings.Join(buffer, ","))
	log.V(4).Info("Returning", "logs", result)
	return result, err
}
