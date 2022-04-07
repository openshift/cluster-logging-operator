package elasticsearch

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ViaQ/logerr/log"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	maxDuration          time.Duration
	defaultRetryInterval time.Duration
)

// Get documents array from index
func GetDocsFromElasticSearch(namespaceName string, pod *v1.Pod, container string, index string, secure bool) (result []interface{}, err error) {
	var hits []interface{}
	maxDuration, _ = time.ParseDuration("5m")
	defaultRetryInterval, _ = time.ParseDuration("10s")
	err = wait.PollImmediate(defaultRetryInterval, maxDuration, func() (done bool, err error) {
		command := []string{}
		if secure {
			cmd := fmt.Sprintf("--query=%s/_search?pretty=true", index)
			command = append(command, "es_util", cmd)
		} else {
			cmd := "curl -s -k http://localhost:9200/" + index + "/_search?pretty"
			command = append(command, "bash", "-c", cmd)
		}
		log.V(2).Info("Running", "container", container, "cmd", command)
		result, err := testruntime.ExecOc(pod, container, command[0], command[1:]...)
		log.V(2).Info("Exec'd", "out", result, "err", err)

		if result != "" && err == nil {
			var elasticResult map[string]interface{}
			err = json.Unmarshal([]byte(result), &elasticResult)
			if err == nil {
				if elasticResult["timed_out"] == false {
					shards := elasticResult["_shards"].(map[string]interface{})
					value := shards["total"].(interface{})
					if int(value.(float64)) == 0 {
						return false, nil
					}
					resultHits := elasticResult["hits"].(map[string]interface{})
					hits = resultHits["hits"].([]interface{})
					return true, nil
				}
			} else {
				log.V(4).Info("Unmarshall failed", "err", err)
			}
		}
		log.V(4).Info("Polling from ElasticSearch", "err", err, "result", result)
		return false, nil
	})
	return hits, err
}
