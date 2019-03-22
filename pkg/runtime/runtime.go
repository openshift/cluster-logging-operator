package runtime

import (
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

//RetryOnConflict is a place for "k8s.io/client-go/util/retry#RetryOnConflict"
type RetryOnConflict func(backoff wait.Backoff, fn func() error) error

//SdkUpdate is a place holder for "github.com/operator-framework/operator-sdk/pkg/sdk#Update"
type SdkUpdate func(object sdk.Object) error

//OperatorRuntime is an adapter to the underlying runtime for the operator to
//all mocking for testing
type OperatorRuntime struct {
	RetryOnConflict RetryOnConflict
	Update          SdkUpdate
}

//New returns the default runtime
func New() *OperatorRuntime {
	return &OperatorRuntime{
		RetryOnConflict: retry.RetryOnConflict,
		Update:          sdk.Update,
	}
}
