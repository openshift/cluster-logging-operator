package functional

import (
	"encoding/json"
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"

	"github.com/ViaQ/logerr/v2/log"
	"k8s.io/apimachinery/pkg/util/wait"
)

var ElasticIndex = map[string]string{
	logging.InputNameApplication:    "app-write",
	logging.InputNameAudit:          "audit-write",
	logging.InputNameInfrastructure: "infra-write",
}

func (f *CollectorFunctionalFramework) GetLogsFromElasticSearch(outputName string, outputLogType string) (result string, err error) {
	index, ok := ElasticIndex[outputLogType]
	if !ok {
		return "", fmt.Errorf(fmt.Sprintf("can't find log of type %s", outputLogType))
	}
	return f.GetLogsFromElasticSearchIndex(outputName, index)
}

func (f *CollectorFunctionalFramework) GetLogsFromElasticSearchIndex(outputName string, index string) (result string, err error) {

	logger := log.NewLogger("elasticsearch")
	buffer := []string{}
	err = wait.PollImmediate(defaultRetryInterval, maxDuration, func() (done bool, err error) {
		cmd := `curl -X GET "localhost:9200/` + index + `/_search?pretty" -H 'Content-Type: application/json' -d'
{
	"query": {
	"match_all": {}
}
}
'`
		result, err = f.RunCommand(outputName, "bash", "-c", cmd)
		if result != "" && err == nil {
			//var elasticResult ElasticSearchResult
			var elasticResult map[string]interface{}
			err = json.Unmarshal([]byte(result), &elasticResult)
			if err == nil {
				if elasticResult["timed_out"] == false {
					resultHits := elasticResult["hits"].(map[string]interface{})
					total := resultHits["total"].(map[string]interface{})
					if int(total["value"].(float64)) == 0 {
						return false, nil
					}
					hits := resultHits["hits"].([]interface{})
					for i := 0; i < len(hits); i++ {
						hit := hits[i].(map[string]interface{})
						jsonHit, err := json.Marshal(hit["_source"])
						if err == nil {
							buffer = append(buffer, string(jsonHit))
						} else {
							logger.V(4).Info("Marshall failed", "err", err)
						}
					}
					return true, nil
				}
			} else {
				logger.V(4).Info("Unmarshall failed", "err", err)
			}
		}
		logger.V(4).Info("Polling from ElasticSearch", "err", err, "result", result)
		return false, nil
	})
	if err == nil {
		result = fmt.Sprintf("[%s]", strings.Join(buffer, ","))
	}
	logger.V(4).Info("Returning", "logs", result)
	return result, err
}
