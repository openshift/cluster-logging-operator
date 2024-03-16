package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	testrt "github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Client operates on any runtime.Object (core or custom) and has Watch to wait efficiently.
type Client struct {
	c          crclient.Client
	httpClient *http.Client
	cfg        *rest.Config
	ctx        context.Context
	mapper     meta.RESTMapper
	rests      *sync.Map // map[schema.GroupVersion]rest.Interface
	timeout    time.Duration
	Labels     map[string]string
}

// New client.
//
// Operations will fail if they take longer than timeout.
// The Labels will be applied to each object created by the client.
func New(cfg *rest.Config, timeout time.Duration, labels map[string]string) (*Client, error) {
	var err error
	if cfg == nil {
		if cfg, err = config.GetConfig(); err != nil {
			return nil, err
		}
	}
	c := &Client{
		cfg:     cfg,
		ctx:     context.Background(),
		rests:   &sync.Map{},
		timeout: timeout,
		Labels:  make(map[string]string, len(labels)),
	}

	if c.httpClient, err = rest.HTTPClientFor(cfg); err != nil {
		return nil, err
	}

	for k, v := range labels {
		c.Labels[k] = v
	}

	if c.mapper, err = apiutil.NewDynamicRESTMapper(c.cfg, c.httpClient); err != nil {
		return nil, err
	}
	if c.c, err = crclient.New(cfg, crclient.Options{Mapper: c.mapper, WarningHandler: crclient.WarningHandlerOptions{SuppressWarnings: true}}); err != nil {
		return nil, err
	}
	return c, nil
}

func logKeyId(v interface{}) (key, id string) {
	switch v := v.(type) {
	case crclient.Object:
		return "object", testrt.ID(v)
	case schema.GroupVersionKind:
		return "type", v.String()
	case schema.GroupVersionResource:
		return "type", v.String()
	default:
		return fmt.Sprintf("unexpected-%T", v), fmt.Sprintf("%v", v)
	}
}

func logBeginEnd(op string, v interface{}, errp *error, kv ...interface{}) func() {
	key, id := logKeyId(v)
	return test.LogBeginEnd(log.WithName(""), op, errp, append(kv, key, id)...)
}

// Create the cluster resource described by the struct o.
func (c *Client) Create(o crclient.Object) (err error) {
	defer logBeginEnd("Create", o, &err)()
	return c.create(o)
}

func (c *Client) create(o crclient.Object) (err error) {
	for k, v := range c.Labels {
		testrt.Labels(o)[k] = v
	}
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.Create(ctx, o)
}

// Get the resource with the name, namespace and type of o, copy it's state into o.
func (c *Client) Get(o crclient.Object) (err error) {
	defer logBeginEnd("Get", o, &err)()
	return c.get(o)
}

// get with no logging, for internal use.
func (c *Client) get(o crclient.Object) (err error) {
	nsName := crclient.ObjectKeyFromObject(o)
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.Get(ctx, nsName, o)
}

// List options aliased from crclient
type ListOption = crclient.ListOption
type ListOptions = crclient.ListOptions
type MatchingLabels = crclient.MatchingLabels
type HasLabels = crclient.HasLabels
type MatchingLabelsSelector = crclient.MatchingLabelsSelector
type MatchingFields = crclient.MatchingFields
type MatchingFieldsSelector = crclient.MatchingFieldsSelector
type InNamespace = crclient.InNamespace
type Limit = crclient.Limit
type Continue = crclient.Continue

// List resources, return results in o, which must be an xxxList struct.
func (c *Client) List(list crclient.ObjectList, opts ...ListOption) (err error) {
	defer logBeginEnd("List", list, &err)()
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.List(ctx, list, opts...)
}

// Update the state of resource with name, namespace and type of o using the state in o.
func (c *Client) Update(o crclient.Object) (err error) {
	defer logBeginEnd("Update", o, &err)()
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.Update(ctx, o)
}

// Delete the of resource with name, namespace and type of o.
func (c *Client) Delete(o crclient.Object) (err error) {
	defer logBeginEnd("Delete", o, &err)()
	return c.delete(o)
}

func (c *Client) delete(o crclient.Object, opts ...crclient.DeleteOption) (err error) {
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.Delete(ctx, o, opts...)
}

// Remove is like Delete but ignores NotFound errors.
func (c *Client) Remove(o crclient.Object) (err error) {
	defer logBeginEnd("Remove", o, &err)()
	return crclient.IgnoreNotFound(c.delete(o))
}

// RemoveSync is like Delete but ignores NotFound errors and deletes in the foreground
func (c *Client) RemoveSync(o crclient.Object) (err error) {
	defer logBeginEnd("Remove", o, &err)()
	return crclient.IgnoreNotFound(c.delete(o, crclient.PropagationPolicy(metav1.DeletePropagationForeground)))
}

// Deleted is a condition for Client.WaitFor true when the event is of type Deleted.
func Deleted(e watch.Event) (bool, error) {
	if e.Type == watch.Deleted {
		return true, nil
	}
	return false, nil
}

// Recreate creates o after removing any existing object of the same name and type.
func (c *Client) Recreate(o crclient.Object) (err error) {
	defer logBeginEnd("Recreate", o, &err)()
	if err := crclient.IgnoreNotFound(c.delete(o)); err != nil {
		return err
	}
	// Clear resource version so we can re-create, even if o was used before.
	testrt.Meta(o).SetResourceVersion("")
	err = c.create(o)
	switch {
	case err == nil:
		return nil
	case errors.IsAlreadyExists(err): // Delete is incomplete.
		if err = c.WaitFor(o, Deleted); err != nil {
			return err
		}
		testrt.Meta(o).SetResourceVersion("")
		return c.Create(o)
	default:
		return err
	}
}

// IgnoreNotFound returns nil on NotFound errors, other values returned unmodified.
func IgnoreNotFound(err error) error {
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}

// IgnoreAlreadyExists returns nil on AlreadyExists errors, other values returned unmodified.
func IgnoreAlreadyExists(err error) error {
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (c *Client) rest(gv schema.GroupVersion) (rest.Interface, error) {
	var err error
	i, _ := c.rests.Load(gv)
	r, _ := i.(rest.Interface)
	if r == nil {
		if r, err = apiutil.RESTClientForGVK(gv.WithKind(""), false, c.cfg, testrt.Codecs, c.httpClient); err == nil {
			c.rests.Store(gv, r)
		}
	}
	return r, err
}

// GroupVersionResource uses the Client's RESTMapping to find the resource name for o.
func (c *Client) GroupVersionResource(o crclient.Object) (schema.GroupVersionResource, error) {
	m, err := c.mapper.RESTMapping(testrt.GroupVersionKind(o).GroupKind())
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return m.Resource, nil
}

// WithLabels returns a client that uses c but has a different set of Create Labels.
func (c *Client) WithLabels(labels map[string]string) *Client {
	c2 := *c
	for k, v := range labels {
		c2.Labels[k] = v
	}
	return &c2
}

// WithTimeout returns a client that uses c but has a different timeout.
func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c2 := *c
	c2.timeout = timeout
	return &c2
}

func (c *Client) Timeout() time.Duration { return c.timeout }

// Host returns the API host or URL used by the client's rest.Config.
func (c *Client) Host() string { return c.cfg.Host }

// ControllerRuntimeClient returns the underlying controller runtime Client
func (c *Client) ControllerRuntimeClient() crclient.Client { return c.c }

// Cfg returns the the client's rest.Config
func (c *Client) Cfg() *rest.Config { return c.cfg }

// ContainerLogStream returns the log stream for a container.
// Empty containerName means default container.
func (c *Client) ContainerLogStream(namespace, podName, containerName string, follow bool) (io.ReadCloser, error) {
	clientset, err := kubernetes.NewForConfig(c.cfg) // Need to use the go-client for logs
	if err != nil {
		return nil, err
	}
	return clientset.CoreV1().Pods(namespace).GetLogs(podName,
		&corev1.PodLogOptions{Container: containerName, Follow: follow}).Stream(c.ctx)
}

// ContainerLogs returns the pod logs, or an error string if there is a problem reading them.
// Empty containerName means default container.
func (c *Client) ContainerLogs(namespace, podName, containerName string) string {
	r, err := c.ContainerLogStream(namespace, podName, containerName, false)
	if err != nil {
		return err.Error()
	}
	defer r.Close()
	b, err := io.ReadAll(r)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// Singleton
var singleton struct {
	c    *Client
	err  error
	once sync.Once
}

const (
	// LabelKey is the create label key for the Get() client.
	LabelKey = "test-client"
	// LabelValue is the create label value for the Get() client.
	LabelValue = "true"
)

// Get returns a process-wide singleton client, created on demand.
// Uses defaults: DefaultTimeout, LabelKey, LabelValue.
func Get() *Client {
	s := &singleton
	s.once.Do(func() {
		s.c, s.err = New(nil, test.SuccessTimeout(), map[string]string{LabelKey: LabelValue})
	})
	if s.err != nil {
		panic(s.err)
	}
	return s.c
}
