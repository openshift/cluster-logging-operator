package stub

import (
	"fmt"
	"strings"
	"testing"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/runtime"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func TestAllowedToReconcileWhenNameIsTheSingletonName(t *testing.T) {
	spec := newClusterLogging("instance")
	if !allowedToReconcile(spec) {
		t.Errorf("Exp. to be allowed to reconcile ClusterLogging named %q", spec.Name)
	}
}
func TestAllowedToReconcileWhenNameIsNotTheSingletonName(t *testing.T) {
	spec := newClusterLogging("myclusterLogging")
	if allowedToReconcile(spec) {
		t.Errorf("Exp. to not be allowed to reconcile ClusterLogging named %q", spec.Name)
	}
}

func newClusterLogging(name string) *logging.ClusterLogging {
	return &logging.ClusterLogging{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

}
func TestUpdateStatusToIgnoreWhenNoError(t *testing.T) {
	mockRuntime := &runtime.OperatorRuntime{
		RetryOnConflict: func(backoff wait.Backoff, fn func() error) error {
			return fn()
		},
		Update: func(object sdk.Object) error {
			return nil
		},
	}

	spec := newClusterLogging("myclusterLogging")
	if err := updateStatusToIgnore(spec, mockRuntime); err != nil {
		t.Errorf("Exp. no error when updating the status for %q: %v", spec.Name, err)
	}
	if spec.Status.Message != singletonMessage {
		t.Errorf("Exp. ClusterLogging status message to indicate it is not the singleton instance but was %q", spec.Status.Message)
	}
}
func TestUpdateStatusToIgnoreUpdatingWhenMessageIsAlreadyCorrect(t *testing.T) {
	mockRuntime := &runtime.OperatorRuntime{
		RetryOnConflict: func(backoff wait.Backoff, fn func() error) error {
			return fn()
		},
		Update: func(object sdk.Object) error {
			var spec = object.(*logging.ClusterLogging)
			if spec.Status.Message == singletonMessage {
				return fmt.Errorf("Should not be updating the message")
			}
			return nil
		},
	}

	spec := newClusterLogging("myclusterLogging")
	spec.Status.Message = singletonMessage
	if err := updateStatusToIgnore(spec, mockRuntime); err != nil {
		t.Errorf("Exp. no error when updating the status for %q: %v", spec.Name, err)
	}
	if spec.Status.Message != singletonMessage {
		t.Errorf("Exp. ClusterLogging status message to indicate it is not the singleton instance but was %q", spec.Status.Message)
	}
}
func TestUpdateStatusWhenRetryOnConflictError(t *testing.T) {
	mockRuntime := &runtime.OperatorRuntime{
		RetryOnConflict: func(backoff wait.Backoff, fn func() error) error {
			return fmt.Errorf("RetryOnConflictTestError")
		},
		Update: func(object sdk.Object) error {
			return nil
		},
	}

	spec := newClusterLogging("myclusterLogging")
	err := updateStatusToIgnore(spec, mockRuntime)
	if err == nil {
		t.Errorf("Exp. error while testing RetryOnConflict Error for %q: %v", spec.Name, err)
	}
	if !strings.Contains(err.Error(), "RetryOnConflictTestError") {
		t.Errorf("Exp. an error messagge while testing RetryOnConflict Error for %q: %v", spec.Name, err)
	}
}
