package runtime

import (
	"fmt"

	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	observabilityv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/version"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// Object is an alias for this central type.
type Object = runtime.Object

// Codecs is a codec factory for the default scheme, including core and our custom types.
var Codecs = serializer.NewCodecFactory(scheme.Scheme)

func init() {
	must(loggingv1alpha1.AddToScheme(scheme.Scheme))
	must(observabilityv1.AddToScheme(scheme.Scheme))
}

// Decode JSON or YAML resource manifest to a new typed struct.
func Decode(manifest string) runtime.Object {
	d := Codecs.UniversalDeserializer()
	o, _, err := d.Decode([]byte(manifest), nil, nil)
	must(err)
	return o
}

// Meta interface to get/set object metadata.
// Panics if o is not a metav1.Object, e.g. if it is a List type.
func Meta(o runtime.Object) metav1.Object {
	m, err := meta.Accessor(o)
	must(err)
	return m
}

// NamespacedName returns the namespaced name of an object.
func NamespacedName(o client.Object) types.NamespacedName {
	nn := client.ObjectKeyFromObject(o)
	return nn
}

// ID returns a human-readable identifier for the object, for debugging and tests.
func ID(o runtime.Object) string {
	gvk, err := apiutil.GVKForObject(o, scheme.Scheme)
	if err != nil {
		return fmt.Sprintf("%v", o)
	}
	m, err := meta.Accessor(o)
	if err != nil {
		return gvk.String()
	}
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	if m.GetNamespace() != "" {
		group := gvr.Group
		if group == "" {
			group = "core" // Core APIs are in group "" which doesn't read well.
		}
		return fmt.Sprintf("%v/%v/namespaces/%v/%v/%v", group, gvr.Version, m.GetNamespace(), gvr.Resource, m.GetName())
	}
	return fmt.Sprintf("%v/%v/%v/%v", gvr.Group, gvr.Version, gvr.Resource, m.GetName())
}

// GroupVersionKind deduces the Kind from the Go type.
func GroupVersionKind(o runtime.Object) schema.GroupVersionKind {
	gvk, err := apiutil.GVKForObject(o, scheme.Scheme)
	must(err)
	return gvk
}

type ObjectLabels map[string]string

// Includes compares an object labels to those given and returns true if
// the object set includes the keys and also matches the values
func (objLabels ObjectLabels) Includes(other ObjectLabels) bool {
	for includeKey, includeValue := range other {
		if value, found := objLabels[includeKey]; found {
			if value != includeValue {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

// Labels returns the labels map for object, guaranteed to be non-nil.
func Labels(o runtime.Object) ObjectLabels {
	m := Meta(o)
	l := m.GetLabels()
	if l == nil {
		l = map[string]string{}
		m.SetLabels(l)
	}
	return l
}

// SetCommonLabels initialize given object labels with K8s Common labels
// These are recommended labels. They make it easier to manage applications but aren't required for any core tooling.
// https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels
func SetCommonLabels(object runtime.Object, name, instanceName, component string) {
	common := map[string]string{
		constants.LabelK8sName:      name,
		constants.LabelK8sInstance:  instanceName,
		constants.LabelK8sComponent: component,
		constants.LabelK8sPartOf:    constants.ClusterLogging,
		constants.LabelK8sManagedBy: constants.ClusterLoggingOperator,
		constants.LabelK8sVersion:   version.Version,
	}
	utils.AddLabels(Meta(object), common)
}

func Selectors(instanceName, component, name string) map[string]string {
	return map[string]string{
		constants.LabelK8sName:      name,
		constants.LabelK8sInstance:  instanceName,
		constants.LabelK8sComponent: component,
		constants.LabelK8sPartOf:    constants.ClusterLogging,
		constants.LabelK8sManagedBy: constants.ClusterLoggingOperator,
	}
}

// Initialize sets name, namespace and type metadata deduced from Go type.
func Initialize(o runtime.Object, namespace, name string, visitors ...func(o runtime.Object)) {
	m := Meta(o)
	m.SetNamespace(namespace)
	m.SetName(name)
	o.GetObjectKind().SetGroupVersionKind(GroupVersionKind(o))
	for _, visitor := range visitors {
		visitor(o)
	}
}

// ServiceDomainName returns "name.namespace.svc".
func ServiceDomainName(o runtime.Object) string {
	m := Meta(o)
	return fmt.Sprintf("%s.%s.svc", m.GetName(), m.GetNamespace())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
