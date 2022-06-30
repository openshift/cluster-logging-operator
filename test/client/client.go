package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	testrt "github.com/openshift/cluster-logging-operator/internal/runtime"
	"golang.org/x/time/rate"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/test"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Object is any k8s object, it combines runtime.Object and metav1.Object
type Object = crclient.Object

// ObjectList is any k8s objct list, it combines runtime.Object and metav1.ListInterface
type ObjectList = crclient.ObjectList

// Client operates on any runtime.Object (core or custom) and has Watch to wait efficiently.
type Client struct {
	c       crclient.Client
	cfg     *rest.Config
	ctx     context.Context
	mapper  meta.RESTMapper
	rests   *sync.Map // map[schema.GroupVersion]rest.Interface
	timeout time.Duration
	Labels  map[string]string
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
	// Don't rate-limit the test client. Tests should be short and fast.
	cfg.RateLimiter = flowcontrol.NewFakeAlwaysRateLimiter()
	c := &Client{
		cfg:     cfg,
		ctx:     context.Background(),
		rests:   &sync.Map{},
		timeout: timeout,
		Labels:  make(map[string]string, len(labels)),
	}
	for k, v := range labels {
		c.Labels[k] = v
	}
	if c.mapper, err = apiutil.NewDynamicRESTMapper(c.cfg, apiutil.WithLimiter(rate.NewLimiter(1000, 10000))); err != nil {
		return nil, err
	}
	if c.c, err = crclient.New(cfg, crclient.Options{Mapper: c.mapper}); err != nil {
		return nil, err
	}
	return c, nil
}

func logKeyId(v interface{}) (key, id string) {
	switch v := v.(type) {
	case Object:
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
func (c *Client) Create(o Object) (err error) {
	defer logBeginEnd("Create", o, &err)()
	return c.create(o)
}

func (c *Client) create(o Object) (err error) {
	// Clear resource version and UID so we can re-create, even if o was used before.
	testrt.Meta(o).SetResourceVersion("")
	testrt.Meta(o).SetUID("")
	for k, v := range c.Labels {
		testrt.Labels(o)[k] = v
	}
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.Create(ctx, o)
}

// Get the resource with the name, namespace and type of o, copy it's state into o.
func (c *Client) Get(o Object) (err error) {
	defer logBeginEnd("Get", o, &err)()
	return c.get(o)
}

// get with no logging, for internal use.
func (c *Client) get(o Object) (err error) {
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
func (c *Client) List(list ObjectList, opts ...ListOption) (err error) {
	defer logBeginEnd("List", list, &err)()
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.List(ctx, list, opts...)
}

// Update the state of resource with name, namespace and type of o using the state in o.
func (c *Client) Update(o Object) (err error) {
	defer logBeginEnd("Update", o, &err)()
	return c.update(o)
}

func (c *Client) update(o Object) (err error) {
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.Update(ctx, o)
}

// Delete the of resource with name, namespace and type of o.
func (c *Client) Delete(o Object) (err error) {
	defer logBeginEnd("Delete", o, &err)()
	return c.delete(o)
}

func (c *Client) delete(o Object, opts ...crclient.DeleteOption) (err error) {
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.Delete(ctx, o, opts...)
}

// Remove is like Delete but ignores NotFound errors.
func (c *Client) Remove(o Object) (err error) {
	defer logBeginEnd("Remove", o, &err)()
	return crclient.IgnoreNotFound(c.delete(o))
}

// RemoveSync is like Remove but waits for object to be finalized and deleted.
func (c *Client) RemoveSync(o Object) (err error) {
	defer func() { err = IgnoreNotFound(err) }()
	defer logBeginEnd("RemoveSync", o, &err)()
	if err := c.delete(o, crclient.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
		return err
	}
	return c.WaitFor(o, Deleted)
}

// Deleted is a condition for Client.WaitFor true when the event is of type Deleted.
func Deleted(e watch.Event) (bool, error) {
	if e.Type == watch.Deleted {
		return true, nil
	}
	return false, nil
}

// Recreate creates o after removing any existing object of the same name and type.
func (c *Client) Recreate(o Object) (err error) {
	defer logBeginEnd("Recreate", o, &err)()
	if err := crclient.IgnoreNotFound(c.delete(o)); err != nil {
		return err
	}
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
		if r, err = apiutil.RESTClientForGVK(gv.WithKind(""), false, c.cfg, testrt.Codecs); err == nil {
			c.rests.Store(gv, r)
		}
	}
	return r, err
}

// GroupVersionResource uses the Client's RESTMapping to find the resource name for o.
func (c *Client) GroupVersionResource(o Object) (schema.GroupVersionResource, error) {
	m, err := c.mapper.RESTMapping(testrt.GroupVersionKind(o).GroupKind())
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return m.Resource, nil
}

// WithLabels returns a derived client that uses c but has a different set of Create labels.
func (c *Client) WithLabels(labels map[string]string) *Client {
	c2 := *c
	for k, v := range labels {
		c2.Labels[k] = v
	}
	return &c2
}

// WithTimeout returns a derived client that uses c but has a different timeout.
func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c2 := *c
	c2.timeout = timeout
	return &c2
}

// WithContext returns a derived client that uses c but has a different context.
// All client operations are interrupted when the context is cancelled.
func (c *Client) WithContext(ctx context.Context) *Client {
	c2 := *c
	c2.ctx = ctx
	return &c2
}

// Timeout returns the client timeout.
func (c *Client) Timeout() time.Duration { return c.timeout }

// Host returns the API host or URL used by the client's rest.Config.
func (c *Client) Host() string { return c.cfg.Host }

// ControllerRuntimeClient returns the underlying controller runtime Client
func (c *Client) ControllerRuntimeClient() crclient.Client { return c.c }

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

// ForEachNoError calls f for each object in turn.
// If any call returns an error, stop and return the error.
func ForEachNoError(f func(Object) error, objs ...Object) (err error) {
	for _, o := range objs {
		if err := f(o); err != nil {
			return err
		}
	}
	return nil
}

// ForEach calls f for each object in turn.
// If a call returns an error, continue and return the last error seen.
func ForEach(f func(Object) error, objs ...Object) (err error) {
	for _, o := range objs {
		if e := f(o); e != nil {
			err = e
		}
	}
	return err
}
