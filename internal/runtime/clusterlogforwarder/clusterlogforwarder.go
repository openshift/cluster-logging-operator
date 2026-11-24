package clusterlogforwarder

import (
	"context"
	"fmt"
	"strings"

	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceAccountRef is a namespace + name pair identifying a collector ServiceAccount.
type ServiceAccountRef struct {
	Namespace string
	Name      string
}

// String returns the ConfigMap data key for this ServiceAccount reference.
// The format "sa_<namespace>_<name>" is collision-free because namespaces
// (DNS-1123 label) and ServiceAccount names (DNS-1123 subdomain) forbid '_',
// and ConfigMap keys may not contain '/'.
func (r ServiceAccountRef) String() string {
	return fmt.Sprintf("sa_%s_%s", r.Namespace, r.Name)
}

// ListServiceAccounts lists all ClusterLogForwarders and returns the
// unique set of ServiceAccount references they declare.
func ListServiceAccounts(ctx context.Context, k8sClient client.Reader) ([]ServiceAccountRef, error) {
	clfList := &obsv1.ClusterLogForwarderList{}
	if err := k8sClient.List(ctx, clfList); err != nil {
		return nil, fmt.Errorf("list ClusterLogForwarders: %w", err)
	}
	seen := map[ServiceAccountRef]struct{}{}
	var refs []ServiceAccountRef
	for i := range clfList.Items {
		sa := strings.TrimSpace(clfList.Items[i].Spec.ServiceAccount.Name)
		if sa == "" {
			continue
		}
		ref := ServiceAccountRef{Namespace: clfList.Items[i].Namespace, Name: sa}
		if _, ok := seen[ref]; !ok {
			seen[ref] = struct{}{}
			refs = append(refs, ref)
		}
	}
	return refs, nil
}

// OtelCollectorRef is a namespace + name pair identifying the
// OpenTelemetryCollector custom resource a ClusterLogForwarder deploys when
// using the otelc collector implementation. The CR is named after the
// forwarder.
type OtelCollectorRef struct {
	Namespace string
	Name      string
}

// String returns the ConfigMap data key for this OpenTelemetryCollector
// reference. The format "otelc_<namespace>_<name>" is collision-free because
// namespaces and ClusterLogForwarder names forbid '_'.
func (r OtelCollectorRef) String() string {
	return fmt.Sprintf("otelc_%s_%s", r.Namespace, r.Name)
}

// ListOtelCollectors lists the OpenTelemetryCollector custom resources
// deployed by ClusterLogForwarders using the otelc collector implementation.
func ListOtelCollectors(ctx context.Context, k8sClient client.Reader) ([]OtelCollectorRef, error) {
	clfList := &obsv1.ClusterLogForwarderList{}
	if err := k8sClient.List(ctx, clfList); err != nil {
		return nil, fmt.Errorf("list ClusterLogForwarders: %w", err)
	}
	var refs []OtelCollectorRef
	for i := range clfList.Items {
		clf := clfList.Items[i]
		if observability.IsOtelc(clf) {
			refs = append(refs, OtelCollectorRef{Namespace: clf.Namespace, Name: clf.Name})
		}
	}
	return refs, nil
}
