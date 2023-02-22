package functional

import (
	"encoding/json"
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"

	log "github.com/ViaQ/logerr/v2/log/static"
	"k8s.io/apimachinery/pkg/util/wait"
)

var ElasticIndex = map[string]string{
	logging.InputNameApplication:    "app-write",
	logging.InputNameAudit:          "audit-write",
	logging.InputNameInfrastructure: "infra-write",
}

func (f *CollectorFunctionalFramework) GetLogsFromElasticSearch(outputName string, outputLogType string, options ...Option) (results []string, err error) {
	index, ok := ElasticIndex[outputLogType]
	if !ok {
		return []string{}, fmt.Errorf(fmt.Sprintf("can't find log of type %s", outputLogType))
	}
	return f.GetLogsFromElasticSearchIndex(outputName, index, options...)
}

func (f *CollectorFunctionalFramework) GetLogsFromElasticSearchIndex(outputName string, index string, options ...Option) (results []string, err error) {
	port := "9200"
	if found, o := OptionsInclude("port", options); found {
		port = o.Value
	}
	err = wait.PollImmediate(defaultRetryInterval, maxDuration, func() (done bool, err error) {
		cmd := `curl -X GET "localhost:` + port + `/` + index + `/_search?pretty" -H 'Content-Type: application/json' -d'
{
	"query": {
	"match_all": {}
}
}
'`
		var result string
		result, err = f.RunCommand(outputName, "bash", "-c", cmd)
		if result != "" && err == nil {
			//var elasticResult ElasticSearchResult
			var elasticResult map[string]interface{}
			log.V(2).Info("results", "response", result)
			err = json.Unmarshal([]byte(result), &elasticResult)
			if err == nil {
				if elasticResult["timed_out"] == false {
					rawHits, ok := elasticResult["hits"]
					if !ok {
						return false, fmt.Errorf("No hits found")
					}
					resultHits := rawHits.(map[string]interface{})
					total, ok := resultHits["total"].(map[string]interface{})
					if ok {
						if int(total["value"].(float64)) == 0 {
							return false, nil
						}
					} else {
						if resultHits["total"].(float64) == 0 {
							return false, nil
						}
					}
					hits := resultHits["hits"].([]interface{})
					for i := 0; i < len(hits); i++ {
						hit := hits[i].(map[string]interface{})
						jsonHit, err := json.Marshal(hit["_source"])
						if err == nil {
							results = append(results, string(jsonHit))
						} else {
							log.V(4).Info("Marshall failed", "err", err)
						}
					}
					return true, nil
				}
			} else {
				log.V(4).Info("Unmarshall failed", "err", err)
			}
		}
		log.V(4).Info("Polling from ElasticSearch", "err", err, "result", result)
		return false, nil
	})
	log.V(4).Info("Returning", "logs", results)
	return results, err
}
