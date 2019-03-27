package k8shandler

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	certLocalPath = "/tmp/"
)

type esCurlStruct struct {
	Method       string // use net/http constants https://golang.org/pkg/net/http/#pkg-constants
	Uri          string
	RequestBody  string
	StatusCode   int
	ResponseBody map[string]interface{}
	Error        error
}

func SetShardAllocation(clusterName, namespace string, state api.ShardAllocationState) (bool, error) {

	payload := &esCurlStruct{
		Method:      http.MethodPut,
		Uri:         "_cluster/settings",
		RequestBody: fmt.Sprintf("{%q:{%q:%q}}", "transient", "cluster.routing.allocation.enable", state),
	}

	curlESService(clusterName, namespace, payload)

	acknowledged := false
	if payload.ResponseBody["acknowledged"] != nil {
		acknowledged = payload.ResponseBody["acknowledged"].(bool)
	}
	return (payload.StatusCode == 200 && acknowledged), payload.Error
}

func GetShardAllocation(clusterName, namespace string) (string, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		Uri:    "_cluster/settings",
	}

	curlESService(clusterName, namespace, payload)

	allocation := ""
	value := walkInterfaceMap("transient.cluster.routing.allocation.enable", payload.ResponseBody)

	if value != nil {
		allocation = value.(string)
	}

	return allocation, payload.Error
}

func GetNodeDiskUsage(clusterName, namespace, nodeName string) (string, float64, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		Uri:    "_cat/nodes?h=name,du,dup",
	}

	curlESService(clusterName, namespace, payload)

	usage := ""
	percentUsage := float64(-1)

	if payload.ResponseBody["results"] != nil {
		response := parseNodeDiskUsage(payload.ResponseBody["results"].(string))
		if nodeResponse, ok := response[nodeName]; ok {
			nodeResponseBody := nodeResponse.(map[string]interface{})

			if nodeResponseBody["used"] != nil {
				usage = nodeResponseBody["used"].(string)
			}

			if nodeResponseBody["used_percent"] != nil {
				percentUsage = nodeResponseBody["used_percent"].(float64)
			}
		}
	}

	return usage, percentUsage, payload.Error
}

func GetThresholdEnabled(clusterName, namespace string) (bool, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		Uri:    "_cluster/settings?include_defaults=true&filter_path=defaults.cluster.routing.allocation.disk",
	}

	curlESService(clusterName, namespace, payload)

	var enabled interface{}

	if value := walkInterfaceMap(
		"defaults.cluster.routing.allocation.disk.threshold_enabled",
		payload.ResponseBody); enabled != nil {

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

	if enabled != nil {
		enabled, _ = strconv.ParseBool(enabled.(string))
	} else {
		enabled = false
	}

	return enabled.(bool), payload.Error
}

func GetDiskWatermarks(clusterName, namespace string) (interface{}, interface{}, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		Uri:    "_cluster/settings?include_defaults=true&filter_path=defaults.cluster.routing.allocation.disk",
	}

	curlESService(clusterName, namespace, payload)

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

	if low != nil {
		if strings.HasSuffix(low.(string), "%") {
			low, _ = strconv.ParseFloat(strings.TrimSuffix(low.(string), "%"), 64)
		} else {
			if strings.HasSuffix(low.(string), "b") {
				low = strings.TrimSuffix(low.(string), "b")
			}
		}
	}

	if high != nil {
		if strings.HasSuffix(high.(string), "%") {
			high, _ = strconv.ParseFloat(strings.TrimSuffix(high.(string), "%"), 64)
		} else {
			if strings.HasSuffix(high.(string), "b") {
				high = strings.TrimSuffix(high.(string), "b")
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

func parseNodeDiskUsage(results string) map[string]interface{} {

	nodeDiskUsage := make(map[string]interface{})

	for _, result := range strings.Split(results, "\n") {
		fields := strings.Split(result, " ")

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

func SetMinMasterNodes(clusterName, namespace string, numberMasters int32) (bool, error) {

	payload := &esCurlStruct{
		Method:      http.MethodPut,
		Uri:         "_cluster/settings",
		RequestBody: fmt.Sprintf("{%q:{%q:%d}}", "persistent", "discovery.zen.minimum_master_nodes", numberMasters),
	}

	curlESService(clusterName, namespace, payload)

	acknowledged := false
	if payload.ResponseBody["acknowledged"] != nil {
		acknowledged = payload.ResponseBody["acknowledged"].(bool)
	}

	return (payload.StatusCode == 200 && acknowledged), payload.Error
}

func GetMinMasterNodes(clusterName, namespace string) (int32, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		Uri:    "_cluster/settings",
	}

	curlESService(clusterName, namespace, payload)

	masterCount := int32(0)
	if payload.ResponseBody["persistent"] != nil {
		persistentBody := payload.ResponseBody["persistent"].(map[string]interface{})
		if persistentBody["discovery.zen.minimum_master_nodes"] != nil {
			masterCount = int32(persistentBody["discovery.zen.minimum_master_nodes"].(float64))
		}
	}

	return masterCount, payload.Error
}

func GetClusterHealth(clusterName, namespace string) (string, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		Uri:    "_cluster/health",
	}

	curlESService(clusterName, namespace, payload)

	status := ""
	if payload.ResponseBody["status"] != nil {
		status = payload.ResponseBody["status"].(string)
	}

	return status, payload.Error
}

func GetClusterNodeCount(clusterName, namespace string) (int32, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		Uri:    "_cluster/health",
	}

	curlESService(clusterName, namespace, payload)

	nodeCount := int32(0)
	if payload.ResponseBody["number_of_nodes"] != nil {
		nodeCount = int32(payload.ResponseBody["number_of_nodes"].(float64))
	}

	return nodeCount, payload.Error
}

// TODO: also check that the number of shards in the response > 0?
func DoSynchronizedFlush(clusterName, namespace string) (bool, error) {

	payload := &esCurlStruct{
		Method: http.MethodPost,
		Uri:    "_flush/synced",
	}

	curlESService(clusterName, namespace, payload)

	return (payload.StatusCode == 200), payload.Error
}

// This will curl the ES service and provide the certs required for doing so
//  it will also return the http and string response
func curlESService(clusterName, namespace string, payload *esCurlStruct) {

	urlString := fmt.Sprintf("https://%s.%s.svc:9200/%s", clusterName, namespace, payload.Uri)
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

	client := getClient(clusterName, namespace)
	resp, err := client.Do(request)

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

func getClient(clusterName, namespace string) *http.Client {

	// get the contents of the secret
	extractSecret(clusterName, namespace)

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            getRootCA(clusterName, namespace),
				Certificates:       getClientCertificates(clusterName, namespace),
			},
		},
	}
}

func extractSecret(secretName, namespace string) {
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
	if err := sdk.Get(secret); err != nil {
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
