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

func QueryPrometheus(host, token, query string) (map[string]interface{}, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		},
	}
	client := &http.Client{Transport: tr, Timeout: 30 * time.Second}
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/api/v1/query", host), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for host %s: %w", host, err)
	}
	request.Header.Add("Authorization", "Bearer "+token)
	request.Header.Add("Accept", "application/json")

	q := request.URL.Query()
	q.Add("query", query)
	request.URL.RawQuery = q.Encode()

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error sending request to Thanos: %w", err)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.V(3).Error(err, "Failed to close response body")
		}
	}()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from Thanos for query %q", response.StatusCode, query)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Thanos response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Thanos response: %w", err)
	}

	return result, nil
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
	return QueryPrometheus(host, token, query)
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
