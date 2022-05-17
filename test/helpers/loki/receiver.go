package loki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/runtime"

	"github.com/ViaQ/logerr/v2/log"
	openshiftv1 "github.com/openshift/api/route/v1"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Image        = "grafana/loki:2.2.1"
	Port         = int32(3100)
	lokiReceiver = "loki-receiver"
)

// Receiver is a service running loki in single-process mode.
type Receiver struct {
	Name    string
	Pod     *corev1.Pod
	service *corev1.Service
	route   *openshiftv1.Route // TODO  use k8s Ingress instead?
	ready   chan struct{}
	timeout time.Duration
}

// NewReceiver creates a Receiver to run Loki in single-process mode.
func NewReceiver(ns, name string) *Receiver {
	r := &Receiver{
		Name:    name,
		Pod:     runtime.NewPod(ns, name),
		service: runtime.NewService(ns, name),
		route:   runtime.NewRoute(ns, name, name, "loki"),
		ready:   make(chan struct{}),
	}
	runtime.Labels(r.Pod)[lokiReceiver] = name
	r.Pod.Spec.Containers = []corev1.Container{{
		Name:  name,
		Image: Image,
		Ports: []corev1.ContainerPort{{Name: name, ContainerPort: Port}},
	}}
	r.service.Spec = corev1.ServiceSpec{
		Selector: map[string]string{lokiReceiver: name},
		Ports:    []corev1.ServicePort{{Name: "loki", Port: Port}},
	}
	return r
}

// Create the receiver's resources. Blocks till created.
func (r *Receiver) Create(c *client.Client) error {
	r.timeout = c.Timeout()
	g := errgroup.Group{}
	for _, o := range []crclient.Object{r.Pod, r.service, r.route} {
		if err := c.Create(o); err != nil {
			return err
		}
	}
	g.Go(func() error { return c.WaitFor(r.Pod, client.PodRunning) })
	g.Go(func() error { return c.WaitFor(r.route, client.RouteReady) })
	if err := g.Wait(); err != nil {
		return err
	}
	// Wait till we can get the metrics page, means server is up.
	return wait.PollImmediate(time.Second, r.timeout, func() (bool, error) {
		resp, err := http.Get(r.ExternalURL("/metrics").String())
		if err == nil {
			defer resp.Body.Close()
			err = test.HTTPError(resp)
		}
		return err == nil, nil
	})
}

// ExternalURL returns the URL of the external route. Only valid after Create()
func (r *Receiver) ExternalURL(path string) *url.URL {
	return &url.URL{Scheme: "http", Host: r.route.Spec.Host, Path: path}
}

// InternalURL returns the internal svc.cluster.local URL
func (r *Receiver) InternalURL(path string) *url.URL {
	host := runtime.SvcClusterLocal(r.service.Namespace, r.service.Name)
	return &url.URL{Scheme: "http", Host: fmt.Sprintf("%v:%v", host, Port), Path: path}
}

// Query from outside cluster for logs matching logQL query expression.
// Returns up to limit values.
func (r *Receiver) Query(logQL string, orgID string, limit int) ([]StreamValues, error) {
	u := r.ExternalURL("/loki/api/v1/query_range")
	q := url.Values{}
	q.Add("query", logQL)
	q.Add("limit", strconv.Itoa(limit))
	q.Add("direction", "FORWARD")
	u.RawQuery = q.Encode()
	newLogger := log.NewLogger("loki-testing")
	newLogger.V(3).Info("Loki Query", "url", u.String(), "org-id", orgID)
	header := http.Header{}
	if orgID != "" {
		header.Add("X-Scope-OrgID", orgID)
	}
	req := &http.Request{
		Method: "GET",
		URL:    u,
		Header: header,
	}
	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		err = test.HTTPError(resp)
	}
	if err != nil {
		newLogger.V(3).Error(err, "Loki Query", "url", u.String())
		return nil, fmt.Errorf("%w\nURL: %v", err, u)
	}
	defer resp.Body.Close()
	qr := QueryResponse{}
	if err = json.NewDecoder(resp.Body).Decode(&qr); err != nil {
		return nil, err
	}
	if qr.Status != "success" {
		return nil, fmt.Errorf("expected 'status: success' in %v", qr)
	}
	if qr.Data.ResultType != "streams" {
		return nil, fmt.Errorf("expected 'resultType: streams' in %v", qr)
	}
	newLogger.V(3).Info("Loki Query done", "result", test.JSONString(qr.Data.Result))
	return qr.Data.Result, nil
}

// QueryUntil repeats the query until at least n lines are received.
func (r *Receiver) QueryUntil(logQL string, orgID string, n int) (values []StreamValues, err error) {
	newLogger := log.NewLogger("loki-testing")
	newLogger.V(2).Info("Loki QueryUntil", "query", logQL, "n", n)
	err = wait.PollImmediate(time.Second, r.timeout, func() (bool, error) {
		var err error
		values, err = r.Query(logQL, orgID, n)
		if err != nil {
			return false, err
		}
		got := 0
		for _, v := range values {
			got += len(v.Values)
		}
		return got >= n, nil
	})
	return values, errors.Wrap(err, fmt.Sprintf("waiting for loki query %q, orgID %q", logQL, orgID))
}

// QueryResponse is the response to a loki query.
type QueryResponse struct {
	Status string    `json:"status"`
	Data   QueryData `json:"data"`
}

// QueryData holds the data for a query
type QueryData struct {
	ResultType string         `json:"resultType"`
	Result     []StreamValues `json:"result"`
}

// StreamValues is a set of log values ["time", "line"] for a log stream.
type StreamValues struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// Lines extracts all the log lines from a QueryResult
func (sv StreamValues) Lines() (lines []string) {
	for _, l := range sv.Values { // Values are ["time", "line"]
		lines = append(lines, l[1])
	}
	return lines
}

// Records extracts log lines and parses as JSON maps.
// Lines that are not valid JSON are are returned as: {"INVALID <error-message>": "original line"}
func (sv StreamValues) Records() (records []map[string]interface{}) {
	for _, l := range sv.Lines() {
		m := map[string]interface{}{}
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			m["INVALID "+err.Error()] = l
		}
		records = append(records, m)
	}
	return records
}

// MakeValue returns a [timestamp, line] pair.
func MakeValue(t time.Time, line string) []string {
	return []string{fmt.Sprintf("%v", t.UnixNano()), line}
}

// MakeValues takes a slice of entries and returns a slice of [timestamp,line] values.
func MakeValues(lines []string) (values [][]string) {
	t := time.Now()
	for _, e := range lines {
		values = append(values, MakeValue(t, e))
	}
	return values
}

type labelResponse struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
}

func (r *Receiver) Labels() ([]string, error) {
	u := r.ExternalURL("/loki/api/v1/labels")
	resp, err := http.Get(u.String())
	if err == nil {
		err = test.HTTPError(resp)
	}
	if err != nil {
		return nil, fmt.Errorf("get %q: %w", u, err)
	}
	defer resp.Body.Close()
	lr := labelResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, err
	}
	return lr.Data, nil
}

func (r *Receiver) Push(sv ...StreamValues) error {
	u := r.ExternalURL("/loki/api/v1/push")
	b, err := json.Marshal(map[string][]StreamValues{"streams": sv})
	if err != nil {
		return err
	}
	resp, err := http.Post(u.String(), "application/json", bytes.NewReader(b))
	if err == nil {
		err = test.HTTPError(resp)
	}
	if err != nil {
		return fmt.Errorf("post %q: %w", u, err)
	}
	resp.Body.Close()
	return nil
}
