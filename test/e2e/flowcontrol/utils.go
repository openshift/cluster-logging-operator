package flowcontrol

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"k8s.io/apimachinery/pkg/util/wait"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	SecretName = `oc get secret -n openshift-monitoring -o name | grep  prometheus-k8s-token* | head -n 1 | grep -o prometheus-k8s.*` //nolint:gosec
	Token      = `oc get secret %s -n openshift-monitoring -o jsonpath={.data.token} | base64 -d`                                     //nolint:gosec
	ThanosHost = `oc get route thanos-querier -n openshift-monitoring -o jsonpath={.spec.host}`

	VectorCompSentEvents = `rate(vector_component_sent_events_total{component_name="%s"}[30s])`
	SumMetric            = `sum(%s)`
	VectorUpTotal        = `vector_started_total`

	maxRetries = 100
)

func WaitForMetricsToShow() bool {

	if err := wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		prometheusResponse := GetCollectorMetrics(VectorUpTotal)
		prometheusResponse = prometheusResponse["data"].(map[string]interface{})
		return len(prometheusResponse["result"].([]interface{})) != 0, nil
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
	var err error
	var secret string
	secret, err = ExecuteCmd(SecretName)
	Expect(err).To(Succeed(), secret)
	secret = strings.TrimSpace(secret)

	secret, err = ExecuteCmd(fmt.Sprintf(Token, secret))
	Expect(err).To(Succeed(), secret)
	secret = strings.TrimSpace(secret)

	var thanosHost string
	thanosHost, err = ExecuteCmd(ThanosHost)
	Expect(err).To(Succeed(), thanosHost)
	thanosHost = strings.TrimSpace(thanosHost)

	return QueryPrometheus(thanosHost, secret, metric)
}

func ExecuteCmd(cmd string) (_ string, err error) {
	var result []byte
	result, err = exec.Command("bash", "-c", cmd).Output()
	return string(result), err
}

func QueryPrometheus(host, token, query string) map[string]interface{} {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			//nolint:gosec
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/api/v1/query", host), nil)
	if err != nil {
		Fail(fmt.Sprintf("Failed to create a new request for %s. Err: %v", host, err))
	}
	request.Header.Add("Authorization", "Bearer "+token)
	request.Header.Add("Accept", "application/json")

	query_body := request.URL.Query()
	query_body.Add("query", query)
	request.URL.RawQuery = query_body.Encode()

	response, err := client.Do(request)
	if err != nil {
		Fail(fmt.Sprintf("Error when sending request to the server %v", err))
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		Fail(fmt.Sprintf("%v", err))
	}

	var result_json map[string]interface{}
	if err := json.Unmarshal(body, &result_json); err != nil {
		Fail(fmt.Sprintf("Failed to unmarshal json %v", err))
	}

	return result_json

}
