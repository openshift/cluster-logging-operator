package esclient

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
	"time"

	"github.com/ViaQ/logerr/kverrors"
	"github.com/ViaQ/logerr/log"
	api "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	"github.com/openshift/elasticsearch-operator/internal/manifests/secret"
	estypes "github.com/openshift/elasticsearch-operator/internal/types/elasticsearch"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	certLocalPath = "/tmp/"
	k8sTokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

type Client interface {
	ClusterName() string

	// Cluster Settings API
	GetClusterNodeVersions() ([]string, error)
	GetThresholdEnabled() (bool, error)
	GetDiskWatermarks() (interface{}, interface{}, interface{}, error)
	GetMinMasterNodes() (int32, error)
	SetMinMasterNodes(numberMasters int32) (bool, error)
	DoSynchronizedFlush() (bool, error)

	// Cluster State API
	GetLowestClusterVersion() (string, error)
	IsNodeInCluster(nodeName string) (bool, error)

	// Health API
	GetClusterHealth() (api.ClusterHealth, error)
	GetClusterHealthStatus() (string, error)
	GetClusterNodeCount() (int32, error)

	// Index API
	GetIndex(name string) (*estypes.Index, error)
	CreateIndex(name string, index *estypes.Index) error
	ReIndex(src, dst, script, lang string) error
	GetAllIndices(name string) (estypes.CatIndicesResponses, error)

	// Index Alias API
	ListIndicesForAlias(aliasPattern string) ([]string, error)
	UpdateAlias(actions estypes.AliasActions) error
	AddAliasForOldIndices() bool

	// Index Settings API
	GetIndexSettings(name string) (*estypes.Index, error)
	UpdateIndexSettings(name string, settings *estypes.IndexSettings) error

	// Nodes API
	GetNodeDiskUsage(nodeName string) (string, float64, error)

	// Replicas
	UpdateReplicaCount(replicaCount int32) error
	GetIndexReplicaCounts() (map[string]interface{}, error)
	GetLowestReplicaValue() (int32, error)

	// Shards API
	ClearTransientShardAllocation() (bool, error)
	GetShardAllocation() (string, error)
	SetShardAllocation(state api.ShardAllocationState) (bool, error)

	// Index Templates API
	CreateIndexTemplate(name string, template *estypes.IndexTemplate) error
	DeleteIndexTemplate(name string) error
	ListTemplates() (sets.String, error)
	GetIndexTemplates() (map[string]estypes.GetIndexTemplate, error)
	UpdateTemplatePrimaryShards(shardCount int32) error

	SetSendRequestFn(fn FnEsSendRequest)
}

type FnEsSendRequest func(cluster, namespace string, payload *EsRequest, client k8sclient.Client)

type esClient struct {
	cluster         string
	namespace       string
	k8sClient       k8sclient.Client
	fnSendEsRequest FnEsSendRequest
}

type EsRequest struct {
	Method          string // use net/http constants https://golang.org/pkg/net/http/#pkg-constants
	URI             string
	RequestBody     string
	StatusCode      int
	RawResponseBody string
	ResponseBody    map[string]interface{}
	Error           error
}

func NewClient(cluster, namespace string, client k8sclient.Client) Client {
	return &esClient{
		cluster:         cluster,
		namespace:       namespace,
		k8sClient:       client,
		fnSendEsRequest: sendEsRequest,
	}
}

func (ec *esClient) SetSendRequestFn(fn FnEsSendRequest) {
	ec.fnSendEsRequest = fn
}

func (ec *esClient) ClusterName() string {
	return ec.cluster
}

func (ec *esClient) errorCtx() kverrors.Context {
	return kverrors.NewContext(
		"namespace", ec.namespace,
		"cluster", ec.ClusterName(),
	)
}

// FIXME: this needs to return an error instead of swallowing
func sendEsRequest(cluster, namespace string, payload *EsRequest, client k8sclient.Client) {
	u := fmt.Sprintf("https://%s.%s.svc:9200/%s", cluster, namespace, payload.URI)
	urlURL, err := url.Parse(u)
	if err != nil {
		log.Error(err, "failed to parse URL", "url", u)
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
	// we use the insecure TLS client here because we are providing the SA token.
	httpClient := getTLSClient(cluster, namespace, client)
	resp, err := httpClient.Do(request)
	if err != nil {
		if resp == nil {
			payload.Error = err
			return
		}
	}
	defer resp.Body.Close()

	if resp != nil {
		// TODO: eventually remove after all ES images have been updated to use SA token auth for EO?
		if resp.StatusCode == http.StatusForbidden ||
			resp.StatusCode == http.StatusUnauthorized {
			log.Info("failed sending payload using bearer token", "method", payload.Method, "url", payload.URI)
			// if we get a 401 that means that we couldn't read from the token and provided
			// no header.
			// if we get a 403 that means the ES cluster doesn't allow us to use
			// our SA token.
			// in both cases, try the old way.

			// Not sure why, but just trying to reuse the request with the old client
			// resulted in a 400 every time. Doing it this way got a 200 response as expected.
			sendRequestWithMTlsClient(cluster, namespace, payload, client)
			return
		}

		payload.StatusCode = resp.StatusCode
		if payload.RawResponseBody, err = getRawBody(resp.Body); err != nil {
			log.Error(err, "failed to get raw response body")
		}
		if payload.ResponseBody, err = getMapFromBody(payload.RawResponseBody); err != nil {
			log.Error(err, "getMapFromBody failed")
		}
	}

	payload.Error = err
}

func sendRequestWithMTlsClient(clusterName, namespace string, payload *EsRequest, client k8sclient.Client) {
	u := fmt.Sprintf("https://%s.%s.svc:9200/%s", clusterName, namespace, payload.URI)
	urlURL, err := url.Parse(u)
	if err != nil {
		log.Error(err, "unable to parse URL", "url", u)
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

	httpClient := getMTlsClient(clusterName, namespace, client)
	resp, err := httpClient.Do(request)
	if err != nil {
		if resp == nil {
			payload.Error = err
			return
		}
	}
	defer resp.Body.Close()

	if resp != nil {
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
			log.Info("failed sending payload using mTLS PKI", "method", payload.Method, "url", payload.URI)
		}

		payload.StatusCode = resp.StatusCode
		if payload.RawResponseBody, err = getRawBody(resp.Body); err != nil {
			log.Error(err, "failed to get raw response body")
		}
		if payload.ResponseBody, err = getMapFromBody(payload.RawResponseBody); err != nil {
			log.Error(err, "getMapFrombody failed")
		}
	}

	payload.Error = err
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
		log.Error(err, "Unable to read auth token from file", "file", tokenFile)
		return "", false
	}

	if len(token) == 0 {
		log.Error(nil, "Unable to read auth token from file", "file", tokenFile)
		return "", false
	}

	return string(token), true
}

// this client is used with the SA token, it does not present any client certs
func getTLSClient(clusterName, namespace string, client k8sclient.Client) *http.Client {
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
			},
		},
	}
}

// this client is used in the case where the SA token is not honored. it presents client certs
// and validates the ES cluster CA cert
func getMTlsClient(clusterName, namespace string, client k8sclient.Client) *http.Client {
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

func getRootCA(clusterName, namespace string) *x509.CertPool {
	certPool := x509.NewCertPool()

	// load cert into []byte
	f := path.Join(certLocalPath, namespace, clusterName, "admin-ca")
	caPem, err := ioutil.ReadFile(f)
	if err != nil {
		log.Error(err, "Unable to read file to get contents", "file", f)
		return nil
	}

	certPool.AppendCertsFromPEM(caPem)

	return certPool
}

func getClientCertificates(clusterName, namespace string) []tls.Certificate {
	certificate, err := tls.LoadX509KeyPair(
		path.Join(certLocalPath, namespace, clusterName, "admin-cert"),
		path.Join(certLocalPath, namespace, clusterName, "admin-key"),
	)
	if err != nil {
		return []tls.Certificate{}
	}

	return []tls.Certificate{
		certificate,
	}
}

func getRawBody(body io.ReadCloser) (string, error) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func getMapFromBody(rawBody string) (map[string]interface{}, error) {
	if rawBody == "" {
		return make(map[string]interface{}), nil
	}
	var results map[string]interface{}
	err := json.Unmarshal([]byte(rawBody), &results)
	if err != nil {
		results = make(map[string]interface{})
		results["results"] = rawBody
	}

	return results, nil
}

func extractSecret(secretName, namespace string, client k8sclient.Client) {
	key := types.NamespacedName{Name: secretName, Namespace: namespace}
	s, err := secret.Get(context.TODO(), client, key)
	if err != nil {
		log.Error(err, "Error reading secret", "secret", secretName)
	}

	// make sure that the dir === secretName exists
	secretDir := path.Join(certLocalPath, namespace, secretName)
	if _, err := os.Stat(secretDir); os.IsNotExist(err) {
		if err = os.MkdirAll(secretDir, 0o755); err != nil {
			log.Error(err, "Error creating dir", "dir", secretDir)
		}
	}

	for _, key := range []string{"admin-ca", "admin-cert", "admin-key"} {

		value, ok := s.Data[key]

		// check to see if the map value exists
		if !ok {
			log.Error(nil, "secret key not found", "key", key)
		}

		secretFile := path.Join(certLocalPath, namespace, secretName, key)
		if err := ioutil.WriteFile(secretFile, value, 0o644); err != nil {
			log.Error(err, "failed to write value to file", "value", value, "file", secretFile)
		}
	}
}
