package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/test"
	testrt "github.com/openshift/cluster-logging-operator/test/runtime"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	debug = log.V(7) // Log at debug verbosity.
	trace = log.V(8) // Lower priority, more verbose.
)

// Client operates on any runtime.Object (core or custom) and has Watch to wait efficiently.
type Client struct {
	c       crclient.Client
	cfg     *rest.Config
	ctx     context.Context
	mapper  meta.RESTMapper
	rests   map[schema.GroupVersion]rest.Interface
	timeout time.Duration
	labels  map[string]string
}

// New client.
//
// Operations will fail if they take longer than timeout.
// The labels will be applied to each object created by the client.
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
		rests:   map[schema.GroupVersion]rest.Interface{},
		timeout: timeout,
		labels:  make(map[string]string, len(labels)),
	}
	for k, v := range labels {
		c.labels[k] = v
	}
	if c.mapper, err = apiutil.NewDynamicRESTMapper(c.cfg); err != nil {
		return nil, err
	}
	if c.c, err = crclient.New(cfg, crclient.Options{Mapper: c.mapper}); err != nil {
		return nil, err
	}
	return c, nil
}

func logKeyId(v interface{}) (key, id string) {
	switch v := v.(type) {
	case runtime.Object:
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
	kv = append([]interface{}{key, id}, kv...)
	trace.Info("Client."+op+" call", kv...)
	start := time.Now()
	return func() {
		kv := append(kv, "delay", time.Since(start))
		if errp != nil && *errp != nil {
			debug.Error(*errp, "Client."+op+" error", kv...)
		} else {
			debug.Info("Client."+op+" ok", kv...)
		}
	}
}

// Create the cluster resource described by the struct o.
func (c *Client) Create(o runtime.Object) (err error) {
	defer logBeginEnd("Create", o, &err)()
	for k, v := range c.labels {
		testrt.Labels(o)[k] = v
	}
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.Create(ctx, o)
}

// Get the resource with the name, namespace and type of o, copy it's state into o.
func (c *Client) Get(o runtime.Object) (err error) {
	defer logBeginEnd("Get", o, &err)()
	return c.get(o)
}

// get with no logging, for internal use.
func (c *Client) get(o runtime.Object) (err error) {
	nsName, err := crclient.ObjectKeyFromObject(o)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.Get(ctx, nsName, o)
}

// ListOption alias
type ListOption = crclient.ListOption

// InNamespace alias
type InNamespace = crclient.InNamespace

// List resources, return results in o, which must be an xxxList struct.
func (c *Client) List(o runtime.Object, opts ...ListOption) (err error) {
	defer logBeginEnd("List", o, &err)()
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.List(ctx, o, opts...)
}

// Update the state of resource with name, namespace and type of o using the state in o.
func (c *Client) Update(o runtime.Object) (err error) {
	defer logBeginEnd("Update", o, &err)()
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.Update(ctx, o)
}

// Delete the of resource with name, namespace and type of o.
func (c *Client) Delete(o runtime.Object) (err error) {
	defer logBeginEnd("Delete", o, &err)()
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.Delete(ctx, o)
}

// Remove is like Delete but ignores NotFound errors.
func (c *Client) Remove(o runtime.Object) error {
	// No logBeginEnd, Delete logs for itself.
	return crclient.IgnoreNotFound(c.Delete(o))
}

func Deleted(e watch.Event) (bool, error) {
	if e.Type == watch.Deleted {
		return true, nil
	}
	return false, nil
}

// Recreate creates o, removing any existing object of the same name and type.
func (c *Client) Recreate(o runtime.Object) (err error) {
	defer logBeginEnd("Recreate", o, &err)()
	if err := c.Remove(o); err != nil {
		return err
	}
	// Clear resource version so we can re-create, even if o was used before.
	testrt.Meta(o).SetResourceVersion("")
	err = c.Create(o)
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

func (c *Client) rest(gv schema.GroupVersion) (rest.Interface, error) {
	var err error
	if c.rests[gv] == nil {
		c.rests[gv], err = apiutil.RESTClientForGVK(gv.WithKind(""), c.cfg, testrt.Codecs)
	}
	return c.rests[gv], err
}

// GetGroupVersionResource uses the Client's RESTMapping to find the resource name for o.
func (c *Client) GetGroupVersionResource(o runtime.Object) (schema.GroupVersionResource, error) {
	m, err := c.mapper.RESTMapping(testrt.GroupVersionKind(o).GroupKind())
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return m.Resource, nil
}

// WithLabels returns a client that uses c but has a different set of Create labels.
func (c *Client) WithLabels(labels map[string]string) *Client {
	c2 := *c
	for k, v := range labels {
		c2.labels[k] = v
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
