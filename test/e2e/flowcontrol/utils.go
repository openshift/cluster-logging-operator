package flowcontrol

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	SecretName = "oc get secret -n openshift-monitoring | grep  prometheus-k8s-token* | head -n 1 | awk '{print $1}'" //nolint:gosec
	Token      = "echo $(oc get secret %s -n openshift-monitoring -o json | jq -r '.data.token') | base64 -d"         //nolint:gosec
	ThanosHost = "oc get route thanos-querier -n openshift-monitoring -o json | jq -r '.spec.host'"

	VectorCompSentEvents = `rate(vector_component_sent_events_total{component_name="%s"}[30s])`
	SumMetric            = `sum(%s)`
	VectorUpTotal        = `vector_started_total`

	maxRetries = 100
)

func WaitForMetricsToShow() bool {
	for i := 0; i < maxRetries; i++ {
		prometheusResponse := GetCollectorMetrics(VectorUpTotal)
		prometheusResponse = prometheusResponse["data"].(map[string]interface{})

		if len(prometheusResponse["result"].([]interface{})) != 0 {
			return true
		}

	}
	return false
}

func ExpectMetricsNotFound(prometheusResponse map[string]interface{}) {
	prometheusResponse = prometheusResponse["data"].(map[string]interface{})
	Expect(len(prometheusResponse["result"].([]interface{})) == 0).To(BeTrue())
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
	secret := ExecuteCmd(SecretName)
	secret = secret[:len(secret)-1] // remove newline character
	token := ExecuteCmd(fmt.Sprintf(Token, secret))

	thanos_host := ExecuteCmd(ThanosHost) // remove newline character
	thanos_host = thanos_host[:len(thanos_host)-1]

	return QueryPrometheus(thanos_host, token, metric)
}

func ExecuteCmd(cmd string) string {
	result, err := exec.Command("bash", "-c", cmd).Output()

	if err != nil {
		Fail(fmt.Sprintf("Command Failed: %s with error %v", cmd, err))
	}
	return string(result)

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
