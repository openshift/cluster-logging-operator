package k8shandler

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	certLocalPath = "/tmp/"
)

type esCurlStruct struct {
	Method       string // use net/http constants https://golang.org/pkg/net/http/#pkg-constants
	URI          string
	RequestBody  string
	StatusCode   int
	ResponseBody map[string]interface{}
	Error        error
}

func SetShardAllocation(clusterName, namespace string, state api.ShardAllocationState, client client.Client) (bool, error) {

	payload := &esCurlStruct{
		Method:      http.MethodPut,
		URI:         "_cluster/settings",
		RequestBody: fmt.Sprintf("{%q:{%q:%q}}", "transient", "cluster.routing.allocation.enable", state),
	}

	curlESService(clusterName, namespace, payload, client)

	acknowledged := false
	if acknowledgedBool, ok := payload.ResponseBody["acknowledged"].(bool); ok {
		acknowledged = acknowledgedBool
	}
	return (payload.StatusCode == 200 && acknowledged), payload.Error
}

func GetShardAllocation(clusterName, namespace string, client client.Client) (string, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "_cluster/settings",
	}

	curlESService(clusterName, namespace, payload, client)

	allocation := ""
	value := walkInterfaceMap("transient.cluster.routing.allocation.enable", payload.ResponseBody)

	if allocationString, ok := value.(string); ok {
		allocation = allocationString
	}

	return allocation, payload.Error
}

func GetNodeDiskUsage(clusterName, namespace, nodeName string, client client.Client) (string, float64, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "_cat/nodes?h=name,du,dup",
	}

	curlESService(clusterName, namespace, payload, client)

	usage := ""
	percentUsage := float64(-1)

	if payload, ok := payload.ResponseBody["results"].(string); ok {
		response := parseNodeDiskUsage(payload)
		if nodeResponse, ok := response[nodeName].(map[string]interface{}); ok {
			if usageString, ok := nodeResponse["used"].(string); ok {
				usage = usageString
			}

			if percentUsageFloat, ok := nodeResponse["used_percent"].(float64); ok {
				percentUsage = percentUsageFloat
			}
		}
	}

	return usage, percentUsage, payload.Error
}

func GetThresholdEnabled(clusterName, namespace string, client client.Client) (bool, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "_cluster/settings?include_defaults=true",
	}

	curlESService(clusterName, namespace, payload, client)

	var enabled interface{}

	if value := walkInterfaceMap(
		"defaults.cluster.routing.allocation.disk.threshold_enabled",
		payload.ResponseBody); value != nil {

		enabled = value
	}

	if value := walkInterfaceMap(
		"persistent.cluster.routing.allocation.disk.threshold_enabled",
		payload.ResponseBody); value != nil {

		enabled = value
	}

	if value := walkInterfaceMap(
		"transient.cluster.routing.allocation.disk.threshold_enabled",
		payload.ResponseBody); value != nil {

		enabled = value
	}

	enabledBool := false
	if enabledString, ok := enabled.(string); ok {
		if enabledTemp, err := strconv.ParseBool(enabledString); err == nil {
			enabledBool = enabledTemp
		}
	}

	return enabledBool, payload.Error
}

func GetDiskWatermarks(clusterName, namespace string, client client.Client) (interface{}, interface{}, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "_cluster/settings?include_defaults=true",
	}

	curlESService(clusterName, namespace, payload, client)

	var low interface{}
	var high interface{}

	if value := walkInterfaceMap(
		"defaults.cluster.routing.allocation.disk.watermark.low",
		payload.ResponseBody); value != nil {

		low = value
	}

	if value := walkInterfaceMap(
		"defaults.cluster.routing.allocation.disk.watermark.high",
		payload.ResponseBody); value != nil {

		high = value
	}

	if value := walkInterfaceMap(
		"persistent.cluster.routing.allocation.disk.watermark.low",
		payload.ResponseBody); value != nil {

		low = value
	}

	if value := walkInterfaceMap(
		"persistent.cluster.routing.allocation.disk.watermark.high",
		payload.ResponseBody); value != nil {

		high = value
	}

	if value := walkInterfaceMap(
		"transient.cluster.routing.allocation.disk.watermark.low",
		payload.ResponseBody); value != nil {

		low = value
	}

	if value := walkInterfaceMap(
		"transient.cluster.routing.allocation.disk.watermark.high",
		payload.ResponseBody); value != nil {

		high = value
	}

	if lowString, ok := low.(string); ok {
		if strings.HasSuffix(lowString, "%") {
			low, _ = strconv.ParseFloat(strings.TrimSuffix(lowString, "%"), 64)
		} else {
			if strings.HasSuffix(lowString, "b") {
				low = strings.TrimSuffix(lowString, "b")
			}
		}
	}

	if highString, ok := high.(string); ok {
		if strings.HasSuffix(highString, "%") {
			high, _ = strconv.ParseFloat(strings.TrimSuffix(highString, "%"), 64)
		} else {
			if strings.HasSuffix(highString, "b") {
				high = strings.TrimSuffix(highString, "b")
			}
		}
	}

	return low, high, payload.Error
}

func walkInterfaceMap(path string, interfaceMap map[string]interface{}) interface{} {

	current := interfaceMap
	keys := strings.Split(path, ".")
	keyCount := len(keys)

	for index, key := range keys {
		if current[key] != nil {
			if index+1 < keyCount {
				current = current[key].(map[string]interface{})
			} else {
				return current[key]
			}
		} else {
			return nil
		}
	}

	return nil
}

// ---
// method: GET
// uri: _cat/nodes?h=name,du,dup
// requestbody: ""
// statuscode: 200
// responsebody:
//   results: |
//     elasticsearch-cm-23bq83d3   6.3gb 5.36
//     elasticsearch-cd-ujt4y3n5-1 6.4gb 5.43
// error: null
func parseNodeDiskUsage(results string) map[string]interface{} {

	nodeDiskUsage := make(map[string]interface{})

	for _, result := range strings.Split(results, "\n") {

		fields := []string{}
		for _, val := range strings.Split(result, " ") {
			if len(val) > 0 {
				fields = append(fields, val)
			}
		}

		if len(fields) == 3 {
			percent, err := strconv.ParseFloat(fields[2], 64)
			if err != nil {
				percent = float64(-1)
			}

			nodeDiskUsage[fields[0]] = map[string]interface{}{
				"used":         strings.ToUpper(strings.TrimSuffix(fields[1], "b")),
				"used_percent": percent,
			}
		}
	}

	return nodeDiskUsage
}

func SetMinMasterNodes(clusterName, namespace string, numberMasters int32, client client.Client) (bool, error) {

	payload := &esCurlStruct{
		Method:      http.MethodPut,
		URI:         "_cluster/settings",
		RequestBody: fmt.Sprintf("{%q:{%q:%d}}", "persistent", "discovery.zen.minimum_master_nodes", numberMasters),
	}

	curlESService(clusterName, namespace, payload, client)

	acknowledged := false
	if acknowledgedBool, ok := payload.ResponseBody["acknowledged"].(bool); ok {
		acknowledged = acknowledgedBool
	}

	return (payload.StatusCode == 200 && acknowledged), payload.Error
}

func GetMinMasterNodes(clusterName, namespace string, client client.Client) (int32, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "_cluster/settings",
	}

	curlESService(clusterName, namespace, payload, client)

	masterCount := int32(0)
	if payload.ResponseBody["persistent"] != nil {
		persistentBody := payload.ResponseBody["persistent"].(map[string]interface{})
		if masterCountFloat, ok := persistentBody["discovery.zen.minimum_master_nodes"].(float64); ok {
			masterCount = int32(masterCountFloat)
		}
	}

	return masterCount, payload.Error
}

func GetClusterHealth(clusterName, namespace string, client client.Client) (string, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "_cluster/health",
	}

	curlESService(clusterName, namespace, payload, client)

	status := ""
	if payload.ResponseBody["status"] != nil {
		if statusString, ok := payload.ResponseBody["status"].(string); ok {
			status = statusString
		}
	}

	return status, payload.Error
}

func GetClusterNodeCount(clusterName, namespace string, client client.Client) (int32, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "_cluster/health",
	}

	curlESService(clusterName, namespace, payload, client)

	nodeCount := int32(0)
	if nodeCountFloat, ok := payload.ResponseBody["number_of_nodes"].(float64); ok {
		// we expect at most double digit numbers here, eg cluster with 15 nodes
		nodeCount = int32(nodeCountFloat)
	}

	return nodeCount, payload.Error
}

// TODO: also check that the number of shards in the response > 0?
func DoSynchronizedFlush(clusterName, namespace string, client client.Client) (bool, error) {

	payload := &esCurlStruct{
		Method: http.MethodPost,
		URI:    "_flush/synced",
	}

	curlESService(clusterName, namespace, payload, client)

	failed := 0
	if shards, ok := payload.ResponseBody["_shards"].(map[string]interface{}); ok {
		if failedFload, ok := shards["failed"].(float64); ok {
			failed = int(failedFload)
		}
	}

	if payload.Error == nil && failed != 0 {
		payload.Error = fmt.Errorf("Failed to flush %d shards in preparation for cluster restart", failed)
	}

	return (payload.StatusCode == 200), payload.Error
}

// This will curl the ES service and provide the certs required for doing so
//  it will also return the http and string response
func curlESService(clusterName, namespace string, payload *esCurlStruct, client client.Client) {

	urlString := fmt.Sprintf("https://%s.%s.svc:9200/%s", clusterName, namespace, payload.URI)
	urlURL, err := url.Parse(urlString)

	if err != nil {
		logrus.Warnf("Unable to parse URL %v: %v", urlString, err)
		return
	}

	request := &http.Request{
		Method: payload.Method,
		URL:    urlURL,
	}

	switch payload.Method {
	case http.MethodGet:
		// no more to do to request...
	case http.MethodPost:
		if payload.RequestBody != "" {
			// add to the request
			request.Header = map[string][]string{
				"Content-Type": []string{
					"application/json",
				},
			}
			request.Body = ioutil.NopCloser(bytes.NewReader([]byte(payload.RequestBody)))
		}

	case http.MethodPut:
		if payload.RequestBody != "" {
			// add to the request
			request.Header = map[string][]string{
				"Content-Type": []string{
					"application/json",
				},
			}
			request.Body = ioutil.NopCloser(bytes.NewReader([]byte(payload.RequestBody)))
		}

	default:
		// unsupported method -- do nothing
		return
	}

	httpClient := getClient(clusterName, namespace, client)
	resp, err := httpClient.Do(request)

	if resp != nil {
		payload.StatusCode = resp.StatusCode
		payload.ResponseBody = getMapFromBody(resp.Body)
	}
	payload.Error = err
}

func getRootCA(clusterName, namespace string) *x509.CertPool {
	certPool := x509.NewCertPool()

	// load cert into []byte
	caPem, err := ioutil.ReadFile(path.Join(certLocalPath, clusterName, "admin-ca"))
	if err != nil {
		logrus.Errorf("Unable to read file to get contents: %v", err)
		return nil
	}

	certPool.AppendCertsFromPEM(caPem)

	return certPool
}

func getMapFromBody(body io.ReadCloser) map[string]interface{} {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)

	var results map[string]interface{}
	err := json.Unmarshal([]byte(buf.String()), &results)
	if err != nil {
		results = make(map[string]interface{})
		results["results"] = buf.String()
	}

	return results
}

func getClientCertificates(clusterName, namespace string) []tls.Certificate {
	certificate, err := tls.LoadX509KeyPair(
		path.Join(certLocalPath, clusterName, "admin-cert"),
		path.Join(certLocalPath, clusterName, "admin-key"),
	)
	if err != nil {
		return []tls.Certificate{}
	}

	return []tls.Certificate{
		certificate,
	}
}

func getClient(clusterName, namespace string, client client.Client) *http.Client {

	// get the contents of the secret
	extractSecret(clusterName, namespace, client)

	// http.Transport sourced from go 1.10.7
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            getRootCA(clusterName, namespace),
				Certificates:       getClientCertificates(clusterName, namespace),
			},
		},
	}
}

func extractSecret(secretName, namespace string, client client.Client) {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
	}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, secret); err != nil {
		if errors.IsNotFound(err) {
			//return err
			logrus.Errorf("Unable to find secret %v: %v", secretName, err)
		}

		logrus.Errorf("Error reading secret %v: %v", secretName, err)
		//return fmt.Errorf("Unable to extract secret to file: %v", secretName, err)
	}

	// make sure that the dir === secretName exists
	if _, err := os.Stat(path.Join(certLocalPath, secretName)); os.IsNotExist(err) {
		err = os.MkdirAll(path.Join(certLocalPath, secretName), 0755)
		if err != nil {
			logrus.Errorf("Error creating dir %v: %v", path.Join(certLocalPath, secretName), err)
		}
	}

	for _, key := range []string{"admin-ca", "admin-cert", "admin-key"} {

		value, ok := secret.Data[key]

		// check to see if the map value exists
		if !ok {
			logrus.Errorf("Error secret key %v not found", key)
			//return fmt.Errorf("No secret data \"%s\" found", key)
		}

		if err := ioutil.WriteFile(path.Join(certLocalPath, secretName, key), value, 0644); err != nil {
			//return fmt.Errorf("Unable to write to working dir: %v", err)
			logrus.Errorf("Error writing %v to %v: %v", value, path.Join(certLocalPath, secretName, key), err)
		}
	}
}
