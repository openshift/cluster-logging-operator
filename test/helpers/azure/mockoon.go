package azure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	Mockoon     = "mockoon"
	Port        = 3000
	Image       = "quay.io/openshift-logging/mockoon-cli:6.2.0"
	AzureDomain = "acme.com"
)

type MockoonLog struct {
	App             string      `json:"app,omitempty"`
	EnvironmentName string      `json:"environmentName,omitempty"`
	EnvironmentUUID string      `json:"environmentUUID,omitempty"`
	Level           string      `json:"level,omitempty"`
	Message         string      `json:"message,omitempty"`
	RequestMethod   string      `json:"requestMethod,omitempty"`
	RequestPath     string      `json:"requestPath,omitempty"`
	RequestProxied  bool        `json:"requestProxied,omitempty"`
	ResponseStatus  int         `json:"responseStatus,omitempty"`
	Timestamp       time.Time   `json:"timestamp"`
	Transaction     Transaction `json:"transaction"`
}
type Headers struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
type QueryParams struct {
	APIVersion string `json:"api-version,omitempty"`
}
type Request struct {
	Body        string      `json:"body,omitempty"`
	Headers     []Headers   `json:"headers,omitempty"`
	Method      string      `json:"method,omitempty"`
	Params      []any       `json:"params,omitempty"`
	Query       string      `json:"query,omitempty"`
	QueryParams QueryParams `json:"queryParams,omitempty"`
	Route       string      `json:"route,omitempty"`
	URLPath     string      `json:"urlPath,omitempty"`
}
type Response struct {
	Body          string    `json:"body,omitempty"`
	Headers       []Headers `json:"headers,omitempty"`
	StatusCode    int       `json:"statusCode,omitempty"`
	StatusMessage string    `json:"statusMessage,omitempty"`
}
type Transaction struct {
	Proxied           bool     `json:"proxied,omitempty"`
	Request           Request  `json:"request,omitempty"`
	Response          Response `json:"response,omitempty"`
	RouteResponseUUID string   `json:"routeResponseUUID,omitempty"`
	RouteUUID         string   `json:"routeUUID,omitempty"`
}

// DecodeMockoonLogs parses Mockoon's NDJSON transaction log output into structured log entries.
func DecodeMockoonLogs(output string) ([]MockoonLog, error) {
	output = "[" + strings.ReplaceAll(output, "}\n{", "},{") + "]"
	var logs []MockoonLog
	if err := json.NewDecoder(bytes.NewBufferString(output)).Decode(&logs); err != nil {
		return nil, fmt.Errorf("error decoding logs: %v", err)
	}
	return logs, nil
}

// MockoonLine builds a single Mockoon transaction log line as JSON. Useful for tests.
func MockoonLine(method, path string, status int, body string) string {
	entry := MockoonLog{
		App:            "mockoon-server",
		Level:          "info",
		Message:        "Transaction recorded",
		RequestMethod:  method,
		RequestPath:    path,
		ResponseStatus: status,
		Transaction: Transaction{
			Request: Request{
				Body:   body,
				Method: method,
				Route:  path,
			},
			Response: Response{
				StatusCode: status,
			},
		},
	}
	b, err := json.Marshal(entry)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal mockoon line: %v", err))
	}
	return string(b)
}
