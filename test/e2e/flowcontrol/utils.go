//nolint:staticcheck
package flowcontrol

import (
	"context"
	"fmt"
	"strconv"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/helpers/prometheus"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	VectorCompSentEvents = `rate(vector_component_sent_events_total{component_name="%s"}[30s])`
	VectorUpTotal        = `vector_started_total`
)

func WaitForMetricsToShow() bool {
	if err := wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		response, err := prometheus.Query(VectorUpTotal)
		if err != nil {
			return false, err
		}
		return prometheus.HasResults(response), nil
	}); err != nil {
		log.V(0).Error(err, "Error waiting for metrics to be available")
		return false
	}
	return true
}

func ExpectMetricsWithinRange(prometheusResponse map[string]interface{}, lower, upper float64) {
	prometheusResponse = prometheusResponse["data"].(map[string]interface{})

	for _, result := range prometheusResponse["result"].([]interface{}) {
		r := result.(map[string]interface{})["value"].([]interface{})
		actual, err := strconv.ParseFloat(r[1].(string), 64)
		if err != nil {
			Fail(fmt.Sprintf("Failed to parse integer %v", err))
		}
		Expect(actual <= upper).To(BeTrue())
		Expect(actual >= lower).To(BeTrue())
	}
}

func GetCollectorMetrics(metric string) map[string]interface{} {
	response, err := prometheus.Query(metric)
	if err != nil {
		Fail(fmt.Sprintf("Failed to query metric %s: %v", metric, err))
	}
	return response
}
