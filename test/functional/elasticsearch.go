package functional

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ViaQ/logerr/log"
	"k8s.io/apimachinery/pkg/util/wait"
)

var ElasticIndex = map[string]string{
	applicationLog: "app-write",
}

func (f *FluentdFunctionalFramework) GetLogsFromElasticSearch(outputName string, outputLogType string) (result string, err error) {
	index, ok := ElasticIndex[outputLogType]
	if !ok {
		return "", fmt.Errorf(fmt.Sprintf("can't find log of type %s", outputLogType))
	}
	return f.GetLogsFromElasticSearchIndex(outputName, index)
}

func (f *FluentdFunctionalFramework) GetLogsFromElasticSearchIndex(outputName string, index string) (result string, err error) {

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
					hits := resultHits["hits"].([]interface{})
					if int(total["value"].(float64)) == 1 {
						hit := hits[0].(map[string]interface{})
						jsonHit, err := json.Marshal(hit["_source"])
						if err == nil {
							result = string(jsonHit)
							return true, nil
						}
					}
				}
			}
		}
		log.V(4).Error(err, "Polling from ElasticSearch")
		return false, nil
	})
	if err == nil {
		result = fmt.Sprintf("[%s]", strings.Join(strings.Split(strings.TrimSpace(result), "\n"), ","))
	}
	log.V(4).Info("Returning", "logs", result)
	return result, err
}
