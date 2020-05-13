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

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
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

func parseBool(path string, interfaceMap map[string]interface{}) bool {
	value := walkInterfaceMap(path, interfaceMap)

	if parsedBool, ok := value.(bool); ok {
		return parsedBool
	} else {
		return false
	}
}

func parseString(path string, interfaceMap map[string]interface{}) string {
	value := walkInterfaceMap(path, interfaceMap)

	if parsedString, ok := value.(string); ok {
		return parsedString
	} else {
		return ""
	}
}

func parseInt32(path string, interfaceMap map[string]interface{}) int32 {
	return int32(parseFloat64(path, interfaceMap))
}

func parseFloat64(path string, interfaceMap map[string]interface{}) float64 {
	value := walkInterfaceMap(path, interfaceMap)

	if parsedFloat, ok := value.(float64); ok {
		return parsedFloat
	} else {
		return float64(-1)
	}
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

func updateAllIndexReplicas(clusterName, namespace string, client client.Client, replicaCount int32) (bool, error) {

	indexHealth, _ := GetIndexReplicaCounts(clusterName, namespace, client)

	// get list of indices and call updateIndexReplicas for each one
	for index, health := range indexHealth {
		if healthMap, ok := health.(map[string]interface{}); ok {
			// only update replicas for indices that don't have same replica count
			if numberOfReplicas := parseString("settings.index.number_of_replicas", healthMap); numberOfReplicas != "" {
				currentReplicas, err := strconv.ParseInt(numberOfReplicas, 10, 32)
				if err != nil {
					return false, err
				}

				if int32(currentReplicas) != replicaCount {
					// best effort initially?
					logrus.Debugf("Updating %v from %d replicas to %d", index, currentReplicas, replicaCount)
					if ack, err := updateIndexReplicas(clusterName, namespace, client, index, replicaCount); err != nil {
						return ack, err
					}
				}
			}
		} else {
			logrus.Warnf("Unable to evaluate the number of replicas for index %q: %v. cluster: %s, namespace: %s ", index, health, clusterName, namespace)
			return false, fmt.Errorf("Unable to evaluate number of replicas for index")
		}
	}

	return true, nil
}

func GetIndexReplicaCounts(clusterName, namespace string, client client.Client) (map[string]interface{}, error) {
	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "*/_settings/index.number_of_replicas",
	}

	curlESService(clusterName, namespace, payload, client)

	return payload.ResponseBody, payload.Error
}

func GetIndexTemplates(clusterName, namespace string, client client.Client) (map[string]interface{}, error) {
	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "_template/common.*",
	}

	curlESService(clusterName, namespace, payload, client)

	return payload.ResponseBody, payload.Error
}

func updateAllIndexTemplateReplicas(clusterName, namespace string, client client.Client, replicaCount int32) (bool, error) {

	// get the index template and then update the replica and put it
	indexTemplates, _ := GetIndexTemplates(clusterName, namespace, client)

	for templateName := range indexTemplates {

		if template, ok := indexTemplates[templateName].(map[string]interface{}); ok {
			if settings, ok := template["settings"].(map[string]interface{}); ok {
				if index, ok := settings["index"].(map[string]interface{}); ok {
					currentReplicas, ok := index["number_of_replicas"].(string)

					if ok && currentReplicas != fmt.Sprintf("%d", replicaCount) {
						template["settings"].(map[string]interface{})["index"].(map[string]interface{})["number_of_replicas"] = fmt.Sprintf("%d", replicaCount)

						templateJson, _ := json.Marshal(template)

						logrus.Debugf("Updating template %v from %v replicas to %d", templateName, currentReplicas, replicaCount)

						payload := &esCurlStruct{
							Method:      http.MethodPut,
							URI:         fmt.Sprintf("_template/%s", templateName),
							RequestBody: string(templateJson),
						}

						curlESService(clusterName, namespace, payload, client)

						acknowledged := false
						if acknowledgedBool, ok := payload.ResponseBody["acknowledged"].(bool); ok {
							acknowledged = acknowledgedBool
						}

						if !(payload.StatusCode == 200 && acknowledged) {
							logrus.Warnf("Unable to update template %q: %v", templateName, payload.Error)
						}
					}
				}
			}
		}

	}

	return true, nil
}

func updateIndexReplicas(clusterName, namespace string, client client.Client, index string, replicaCount int32) (bool, error) {
	payload := &esCurlStruct{
		Method:      http.MethodPut,
		URI:         fmt.Sprintf("%s/_settings", index),
		RequestBody: fmt.Sprintf("{%q:\"%d\"}}", "index.number_of_replicas", replicaCount),
	}

	curlESService(clusterName, namespace, payload, client)

	acknowledged := false
	if acknowledgedBool, ok := payload.ResponseBody["acknowledged"].(bool); ok {
		acknowledged = acknowledgedBool
	}
	return payload.StatusCode == 200 && acknowledged, payload.Error
}

func ensureTokenHeader(header http.Header) http.Header {
	if header == nil {
		header = map[string][]string{}
	}

	if saToken, ok := readSAToken(k8sTokenFile); ok {
		header.Set("Authorization", fmt.Sprintf("Bearer %s", saToken))
	}

	return header
}

// we want to read each time so that we can be sure to have the most up to date
// token in the case where our perms change and a new token is mounted
func readSAToken(tokenFile string) (string, bool) {
	// read from /var/run/secrets/kubernetes.io/serviceaccount/token
	token, err := ioutil.ReadFile(tokenFile)

	if err != nil {
		logrus.Errorf("Unable to read auth token from file [%s]: %v", tokenFile, err)
		return "", false
	}

	if len(token) == 0 {
		logrus.Errorf("Unable to read auth token from file [%s]: empty token", tokenFile)
		return "", false
	}

	return string(token), true
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
				"Content-Type": {
					"application/json",
				},
			}
			request.Body = ioutil.NopCloser(bytes.NewReader([]byte(payload.RequestBody)))
		}

	case http.MethodPut:
		if payload.RequestBody != "" {
			// add to the request
			request.Header = map[string][]string{
				"Content-Type": {
					"application/json",
				},
			}
			request.Body = ioutil.NopCloser(bytes.NewReader([]byte(payload.RequestBody)))
		}

	default:
		// unsupported method -- do nothing
		return
	}

	request.Header = ensureTokenHeader(request.Header)
	httpClient := getClient(clusterName, namespace, client)
	resp, err := httpClient.Do(request)

	if resp != nil {
		// TODO: eventually remove after all ES images have been updated to use SA token auth for EO?
		if resp.StatusCode == http.StatusForbidden ||
			resp.StatusCode == http.StatusUnauthorized {
			// if we get a 401 that means that we couldn't read from the token and provided
			// no header.
			// if we get a 403 that means the ES cluster doesn't allow us to use
			// our SA token.
			// in both cases, try the old way.

			// Not sure why, but just trying to reuse the request with the old client
			// resulted in a 400 every time. Doing it this way got a 200 response as expected.
			curlESServiceOldClient(clusterName, namespace, payload, client)
			return
		}

		payload.StatusCode = resp.StatusCode
		if payload.ResponseBody, err = getMapFromBody(resp.Body); err != nil {
			logrus.Warnf("getMapFromBody failed. E: %s\r\n", err.Error())
		}
	}

	payload.Error = err
}

func curlESServiceOldClient(clusterName, namespace string, payload *esCurlStruct, client client.Client) {

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
				"Content-Type": {
					"application/json",
				},
			}
			request.Body = ioutil.NopCloser(bytes.NewReader([]byte(payload.RequestBody)))
		}

	case http.MethodPut:
		if payload.RequestBody != "" {
			// add to the request
			request.Header = map[string][]string{
				"Content-Type": {
					"application/json",
				},
			}
			request.Body = ioutil.NopCloser(bytes.NewReader([]byte(payload.RequestBody)))
		}

	default:
		// unsupported method -- do nothing
		return
	}

	httpClient := getOldClient(clusterName, namespace, client)
	resp, err := httpClient.Do(request)

	if resp != nil {
		payload.StatusCode = resp.StatusCode
		if payload.ResponseBody, err = getMapFromBody(resp.Body); err != nil {
			logrus.Warnf("getMapFromBody failed. E: %s\r\n", err.Error())
		}
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

func getMapFromBody(body io.ReadCloser) (map[string]interface{}, error) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		return make(map[string]interface{}), err
	}

	var results map[string]interface{}
	err := json.Unmarshal([]byte(buf.String()), &results)
	if err != nil {
		results = make(map[string]interface{})
		results["results"] = buf.String()
	}

	return results, nil
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

func getOldClient(clusterName, namespace string, client client.Client) *http.Client {

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
