// Licensed to Red Hat, Inc under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Red Hat, Inc licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/openshift/cluster-logging-operator/test"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
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

// Create the cluster resource described by the struct o.
func (c *Client) Create(o runtime.Object) error {
	for k, v := range c.labels {
		testruntime.Labels(o)[k] = v
	}
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.Create(ctx, o)
}

// Get the resource with the name, namespace and type of o, copy it's state into o.
func (c *Client) Get(o runtime.Object) error {
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
func (c *Client) List(o runtime.Object, opts ...ListOption) error {
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.List(ctx, o, opts...)
}

// Update the state of resource with name, namespace and type of o using the state in o.
func (c *Client) Update(o runtime.Object) error {
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.Update(ctx, o)
}

// Delete the of resource with name, namespace and type of o.
func (c *Client) Delete(o runtime.Object) error {
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return c.c.Delete(ctx, o)
}

// Remove is like Delete but ignores NotFound errors.
func (c *Client) Remove(o runtime.Object) error {
	return crclient.IgnoreNotFound(c.Delete(o))
}

func Deleted(e watch.Event) (bool, error) {
	if e.Type == watch.Deleted {
		return true, nil
	}
	return false, nil
}

// Recreate calls Remove, waits for the object to be gone, then calls Create.
func (c *Client) Recreate(o runtime.Object) error {
	if err := c.Remove(o); err != nil {
		return err
	}
	// Clear resource version so we can re-create, even if o was used before.
	testruntime.Meta(o).SetResourceVersion("")
	err := c.Create(o)
	switch {
	case err == nil:
		return nil
	case errors.IsAlreadyExists(err):
		if err := c.WaitFor(o, Deleted); err != nil {
			return err
		}
		testruntime.Meta(o).SetResourceVersion("")
		return c.Create(o)
	default:
		return err
	}
}

func (c *Client) rest(gv schema.GroupVersion) (rest.Interface, error) {
	var err error
	if c.rests[gv] == nil {
		c.rests[gv], err = apiutil.RESTClientForGVK(gv.WithKind(""), c.cfg, testruntime.Codecs)
	}
	return c.rests[gv], err
}

// GetGroupVersionResource uses the Client's RESTMapping to find the resource name for o.
func (c *Client) GetGroupVersionResource(o runtime.Object) (schema.GroupVersionResource, error) {
	m, err := c.mapper.RESTMapping(testruntime.GroupVersionKind(o).GroupKind())
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

// ID returns a human-readable identifier for the object.
func (c *Client) ID(o runtime.Object) string {
	gvr, err := c.GetGroupVersionResource(o)
	if err != nil {
		return testruntime.ID(o)
	}
	m := testruntime.Meta(o)
	return fmt.Sprintf("[%v/%v, Namespace=%v]", gvr.Resource, m.GetName(), m.GetNamespace())
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
	// DefaultSuccessTimeout is the timeout for the Get() client.
	DefaultSuccessTimeout = test.DefaultSuccessTimeout
	// DefaultFailureTimeout is a timeout for tests that _expect_ a timeout.
	DefaultFailureTimeout = test.DefaultFailureTimeout
)

// Get returns a process-wide singleton client, created on demand.
// Uses defaults: DefaultTimeout, LabelKey, LabelValue.
func Get() *Client {
	s := &singleton
	s.once.Do(func() {
		s.c, s.err = New(nil, DefaultSuccessTimeout, map[string]string{LabelKey: LabelValue})
	})
	if s.err != nil {
		panic(s.err)
	}
	return s.c
}

// All applies f to each object in turn, returns on the first error.
func All(f func(runtime.Object) error, objs ...runtime.Object) error {
	for _, o := range objs {
		if err := f(o); err != nil {
			return nil
		}
	}
	return nil
}
