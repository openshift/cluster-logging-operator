package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/openshift/cluster-logging-operator/test/runtime"
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
// Reports on long delays between watch events.
type watcher struct {
	watch  watch.Interface
	result chan watch.Event
	cancel func()
}

func newWatcher(wi watch.Interface, cancel func(), msg string) watch.Interface {
	w := &watcher{watch: wi, result: make(chan watch.Event), cancel: cancel}
	go func() {
		for e := range wi.ResultChan() {
			w.result <- e
		}
	}()
	return w
}
func (w *watcher) Stop()                          { w.watch.Stop(); w.cancel() }
func (w *watcher) ResultChan() <-chan watch.Event { return w.result }

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
	return newWatcher(w, cancel, fmt.Sprintf("client.Watch [%v, Namespace=%v, Fields=%q, Labels=%q]", gvr.Resource, namespace, opts.FieldSelector, opts.LabelSelector)), nil
}

// WatchObject returns a watch for changes to a single object.
func (c *Client) WatchObject(o runtime.Object) (watch.Interface, error) {
	gvr, err := c.GetGroupVersionResource(o)
	if err != nil {
		return nil, err
	}
	m := runtime.Meta(o)
	opts := metav1.ListOptions{FieldSelector: "metadata.name=" + m.GetName()}
	w, err := c.Watch(m.GetNamespace(), gvr, opts)
	return w, err
}

// Condition returns true if an event meets the condition, error if something goes wrong.
type Condition func(watch.Event) (bool, error)

// WaitFor watches o until condition() returns true or error, or c.Timeout expires.
// o is updated (replaced) by the latest state.
func (c *Client) WaitFor(o runtime.Object, condition Condition) error {
	w, err := c.WatchObject(o)
	if err != nil {
		return err
	}
	for {
		select {
		case e, ok := <-w.ResultChan():
			if !ok {
				return ErrWatchClosed
			}
			// Copy payload of event to update the original object.
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
