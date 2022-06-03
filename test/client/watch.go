package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	gort "runtime"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/ViaQ/logerr/v2/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	ErrWatchClosed = errors.New("watch closed")
	ErrTimeout     = errors.New("timeout")
)

// watcher wraps a watch.Interface.
// Calls cancel() when the watch is stopped.
type watcher struct {
	watch.Interface
	cancel func()
}

func (w *watcher) Stop() { w.Interface.Stop(); w.cancel() }

// Watch for changes in namespace to objects with the given GroupVersionResource,
// apply the given list options to the watch.
func (c *Client) Watch(gvr schema.GroupVersionResource, opts ...ListOption) (w watch.Interface, err error) {
	restClient, err := c.rest(gvr.GroupVersion())
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	listOpts := ListOptions{Raw: &metav1.ListOptions{Watch: true}}
	listOpts.ApplyOptions(opts)
	w, err = restClient.Get().
		Timeout(c.timeout).
		NamespaceIfScoped(listOpts.Namespace, listOpts.Namespace != "").
		Resource(gvr.Resource).
		VersionedParams(listOpts.AsListOptions(), scheme.ParameterCodec).
		Watch(ctx)
	if err != nil {
		cancel()
		return nil, err
	}
	return &watcher{Interface: w, cancel: cancel}, nil
}

// WatchTypeOf returns a watch using the GroupVersionResource of object o.
func (c *Client) WatchTypeOf(o client.Object, opts ...ListOption) (w watch.Interface, err error) {
	gvr, err := c.GroupVersionResource(o)
	if err != nil {
		return nil, err
	}
	return c.Watch(gvr, opts...)
}

// WatchObject returns a watch for changes to a single named object.
// Note: it is not an error if no such object exists, the watcher will wait for creation.
func (c *Client) WatchObject(o client.Object) (w watch.Interface, err error) {
	mo := runtime.Meta(o)
	return c.WatchTypeOf(o,
		InNamespace(mo.GetNamespace()),
		MatchingFields{"metadata.name": mo.GetName()},
	)
}

// Condition returns true if an event meets the condition, error if something goes wrong.
type Condition func(watch.Event) (bool, error)

// funcName extracts the name of a function object for debug logs.
func funcName(f interface{}) string {
	name := gort.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	i := strings.LastIndex(name, ".")
	return name[i+1:]
}

func (c *Client) waitFor(o runtime.Object, condition Condition, w watch.Interface, msg string) (err error) {
	defer logBeginEnd(msg, o, &err)()
	start := time.Now()
	defer w.Stop()
	for {
		select {
		case e, ok := <-w.ResultChan():
			if !ok {
				return fmt.Errorf("%w: %v: %v", ErrWatchClosed, msg, runtime.ID(o))
			}
			log.NewLogger("test-watch").V(3).Info("event: "+msg,
				"object", runtime.ID(e.Object),
				"type", e.Type,
				"elapsed", time.Since(start).String(),
			)
			start = time.Now()
			// Copy payload of event to the original object so it is up-to-date.
			rhs := reflect.Indirect(reflect.ValueOf(e.Object))
			lhs := reflect.Indirect(reflect.ValueOf(o))
			if rhs.IsValid() && lhs.IsValid() && rhs.Type().AssignableTo(lhs.Type()) {
				lhs.Set(rhs)
			}
			if ok, err := condition(e); ok {
				return nil
			} else if err != nil {
				return err
			}
		case <-time.After(c.Timeout()):
			return ErrTimeout
		}
	}
}

// WaitFor watches o until condition() returns true or error, or c.Timeout expires.
// It is not an error if o does not exist, it will be waited for.
// o is updated from the last object seen by the watch.
func (c *Client) WaitFor(o client.Object, condition Condition) (err error) {
	w, err := c.WatchObject(o)
	if err != nil {
		return err
	}
	defer w.Stop()
	msg := fmt.Sprintf("WaitFor(%v)", funcName(condition))
	return c.waitFor(o, condition, w, msg)
}

// WaitForType watches for events involving objects of the same type as o,
// until condition() returns true or error, or c.Timeout expires.
// o is updated from the last object seen by the watch.
func (c *Client) WaitForType(o client.Object, condition Condition, opts ...ListOption) (err error) {
	w, err := c.WatchTypeOf(o, opts...)
	if err != nil {
		return err
	}
	defer w.Stop()
	msg := fmt.Sprintf("WaitForTypeOf(%v, %#v)", funcName(condition), opts)
	return c.waitFor(o, condition, w, msg)
}
