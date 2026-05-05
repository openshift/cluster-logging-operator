package prometheus

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
)

const (
	tokenCmd      = `oc create token prometheus-k8s -n openshift-monitoring` //nolint:gosec
	thanosHostCmd = `oc get route thanos-querier -n openshift-monitoring -o jsonpath={.spec.host}`
)

func executeCmd(cmd string) (string, error) {
	result, err := exec.Command("bash", "-c", cmd).Output()
	return string(result), err
}

func GetToken() (string, error) {
	token, err := executeCmd(tokenCmd)
	return strings.TrimSpace(token), err
}

func GetThanosHost() (string, error) {
	host, err := executeCmd(thanosHostCmd)
	return strings.TrimSpace(host), err
}

func QueryPrometheus(host, token, query string) map[string]interface{} {
	empty := map[string]any{}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		},
	}
	client := &http.Client{Transport: tr, Timeout: 30 * time.Second}
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/api/v1/query", host), nil)
	if err != nil {
		log.V(0).Error(err, "Failed to create request", "host", host)
		return empty
	}
	request.Header.Add("Authorization", "Bearer "+token)
	request.Header.Add("Accept", "application/json")

	q := request.URL.Query()
	q.Add("query", query)
	request.URL.RawQuery = q.Encode()

	response, err := client.Do(request)
	if err != nil {
		log.V(0).Error(err, "Error sending request to Thanos")
		return empty
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.V(3).Error(err, "Failed to close response body")
		}
	}()

	if response.StatusCode != http.StatusOK {
		log.V(0).Info("Unexpected status from Thanos", "status", response.StatusCode, "query", query)
		return empty
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.V(0).Error(err, "Failed to read response body")
		return empty
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.V(0).Error(err, "Failed to unmarshal Thanos response")
		return empty
	}

	return result
}

func Query(query string) (map[string]interface{}, error) {
	token, err := GetToken()
	if err != nil {
		return nil, err
	}
	host, err := GetThanosHost()
	if err != nil {
		return nil, err
	}
	return QueryPrometheus(host, token, query), nil
}

func HasResults(response map[string]interface{}) bool {
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return false
	}
	results, ok := data["result"].([]interface{})
	if !ok {
		return false
	}
	return len(results) > 0
}
