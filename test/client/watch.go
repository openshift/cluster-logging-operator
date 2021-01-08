package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	gort "runtime"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/test/runtime"
	testrt "github.com/openshift/cluster-logging-operator/test/runtime"
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
func (c *Client) Watch(namespace string, gvr schema.GroupVersionResource, opts metav1.ListOptions) (watch.Interface, error) {
	restClient, err := c.rest(gvr.GroupVersion())
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	opts.Watch = true
	w, err := restClient.Get().
		Timeout(c.timeout).
		Namespace(namespace).
		Resource(gvr.Resource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
	if err != nil {
		cancel()
		return nil, err
	}
	return &watcher{Interface: w, cancel: cancel}, nil
}

// WatchObject returns a watch for changes to a single named object.
// Note: it is not an error if no such object exists, the watcher will wait for creation.
func (c *Client) WatchObject(o runtime.Object) (w watch.Interface, err error) {
	gvr, err := c.GetGroupVersionResource(o)
	if err != nil {
		return nil, err
	}
	m := runtime.Meta(o)
	opts := metav1.ListOptions{FieldSelector: "metadata.name=" + m.GetName()}
	w, err = c.Watch(m.GetNamespace(), gvr, opts)
	return w, err
}

// Condition returns true if an event meets the condition, error if something goes wrong.
type Condition func(watch.Event) (bool, error)

// funcName extracts the name of a function object for debug logs.
func funcName(f interface{}) string {
	name := gort.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	i := strings.LastIndex(name, ".")
	return name[i+1:]
}

// WaitFor watches o until condition() returns true or error, or c.Timeout expires.
// o is updated (replaced) by the latest state.
func (c *Client) WaitFor(o runtime.Object, condition Condition) (err error) {
	defer logBeginEnd(fmt.Sprintf("WaitFor(%v)", funcName(condition)), o, &err,
		"resourceVersion", testrt.Meta(o).GetResourceVersion())()
	start := time.Now()
	w, err := c.WatchObject(o)
	if err != nil {
		return err
	}
	defer w.Stop()
	// Watch won't fail if the object doesn't exist, make sure it does or we'll wait forever.
	if err = c.get(o); err != nil {
		return err
	}
	for {
		select {
		case e, ok := <-w.ResultChan():
			if !ok {
				return ErrWatchClosed
			}
			trace.Info("Client.WaitFor event",
				"object", runtime.ID(o),
				"resourceVersion", testrt.Meta(o).GetResourceVersion(),
				"event", e.Type,
				"delay", time.Since(start),
			)
			start = time.Now()
			// Copy payload of event to the original object so it is up-to-date.
			if rhs := reflect.Indirect(reflect.ValueOf(e.Object)); rhs.IsValid() {
				if lhs := reflect.Indirect(reflect.ValueOf(o)); lhs.IsValid() {
					lhs.Set(rhs)
				}
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
