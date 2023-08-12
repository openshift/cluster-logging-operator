package main

import (
	"context"
	"fmt"
	esv1 "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	security "github.com/openshift/api/security/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

type ObjectWriterClient struct {
	client.WithWatch
	cache map[string]runtime.Object
}

func NewObjectWriterClient(initRuntimeObjs []runtime.Object) *ObjectWriterClient {
	watch := fake.NewClientBuilder().
		WithRuntimeObjects(initRuntimeObjs...).
		Build()
	watch.Scheme().AddKnownTypes(esv1.GroupVersion, &esv1.Elasticsearch{}, &esv1.Kibana{})
	watch.Scheme().AddKnownTypes(security.GroupVersion, &security.SecurityContextConstraints{})
	watch.Scheme().AddKnownTypes(monitoringv1.SchemeGroupVersion, &monitoringv1.ServiceMonitor{}, &monitoringv1.PrometheusRule{})
	return &ObjectWriterClient{
		watch,
		map[string]runtime.Object{},
	}
}

func (w *ObjectWriterClient) put(obj client.Object) {
	key := fmt.Sprintf("%v/%v/%v", obj.GetObjectKind(), obj.GetNamespace(), obj.GetName())
	w.cache[key] = obj
}

func (w *ObjectWriterClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	w.put(obj)
	return w.WithWatch.Create(ctx, obj, opts...)
}

func (w *ObjectWriterClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	w.put(obj)
	return w.WithWatch.Update(ctx, obj, opts...)
}

func (w *ObjectWriterClient) Values() []runtime.Object {
	results := []runtime.Object{}
	for _, value := range w.cache {
		results = append(results, value)
	}
	return results
}
