package observability

import (
	"context"
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var lokiStackGVK = schema.GroupVersionKind{
	Group:   "loki.grafana.com",
	Version: "v1",
	Kind:    "LokiStack",
}

// validateName rejects ClusterLogForwarder names that match a LokiStack CR in the same
// namespace. Both operators create a ConfigMap named "{name}-config", which causes
// the configurations to overwrite each other.
func validateName(forwarderContext internalcontext.ForwarderContext) {
	clf := forwarderContext.Forwarder
	key := types.NamespacedName{Name: clf.Name, Namespace: clf.Namespace}

	lokiStack := &unstructured.Unstructured{}
	lokiStack.SetGroupVersionKind(lokiStackGVK)

	if err := forwarderContext.Client.Get(context.TODO(), key, lokiStack); err != nil {
		if errors.IsNotFound(err) || meta.IsNoMatchError(err) {
			internalobs.RemoveConditionByType(&clf.Status.Conditions, obs.ConditionTypeName)
			return
		}
		return
	}

	configMapName := clf.Name + "-config"
	condition := internalobs.NewCondition(obs.ConditionTypeName, obs.ConditionFalse, obs.ReasonNameConflict, "")
	condition.Message = fmt.Sprintf(
		`ClusterLogForwarder name %q conflicts with LokiStack %q in namespace %q: both use ConfigMap %q`,
		clf.Name, clf.Name, clf.Namespace, configMapName,
	)
	internalobs.SetCondition(&clf.Status.Conditions, condition)
}
