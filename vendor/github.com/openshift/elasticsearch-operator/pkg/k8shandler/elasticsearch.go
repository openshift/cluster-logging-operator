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

	"github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
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

func SetShardAllocation(clusterName, namespace string, state v1alpha1.ShardAllocationState) (bool, error) {

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
	if payload.ResponseBody["transient"] != nil {
		transientBody := payload.ResponseBody["transient"].(map[string]interface{})
		if transientBody["cluster"] != nil {
			clusterBody := transientBody["cluster"].(map[string]interface{})
			if clusterBody["routing"] != nil {
				routingBody := clusterBody["routing"].(map[string]interface{})
				if routingBody["allocation"] != nil {
					allocationBody := routingBody["allocation"].(map[string]interface{})
					if allocationBody["enable"] != nil {
						allocation = allocationBody["enable"].(string)
					}
				}
			}
		}
	}

	return allocation, payload.Error
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
