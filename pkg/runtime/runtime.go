package runtime

import (
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

//RetryOnConflict is a place for "k8s.io/client-go/util/retry#RetryOnConflict"
type RetryOnConflict func(backoff wait.Backoff, fn func() error) error

//SdkGet is a place holder for "github.com/operator-framework/operator-sdk/pkg/sdk#Get"
// type SdkGet func(object sdk.Object, opts ...sdk.GetOption) error

//SdkUpdate is a place holder for "github.com/operator-framework/operator-sdk/pkg/sdk#Update"
type SdkUpdate func(object sdk.Object) error

//SdkCreate is a place holder for "github.com/operator-framework/operator-sdk/pkg/sdk#Create"
type SdkCreate func(object sdk.Object) error

//OperatorRuntime is an adapter to the underlying runtime for the operator to
//all mocking for testing
type OperatorRuntime struct {
	RetryOnConflict RetryOnConflict
	// Get             SdkGet
	Update SdkUpdate
	Create SdkCreate
}

//New returns the default runtime
func New() *OperatorRuntime {
	return &OperatorRuntime{
		RetryOnConflict: retry.RetryOnConflict,
		// Get:             sdk.Get,
		Update: sdk.Update,
		Create: sdk.Create,
	}
}

// //CreateOrUpdateSecret creates or updates a secret and retries on conflict
// func CreateOrUpdateSecret(secret sdk.Object) (err error) {
// 	err = sdk.Create(secret)
// 	if err != nil {
// 		if !errors.IsAlreadyExists(err) {
// 			return fmt.Errorf("Failure constructing %v secret: %v", secret.Name, err)
// 		}

// 		current := secret.DeepCopy()
// 		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
// 			if err = sdk.Get(current); err != nil {
// 				if errors.IsNotFound(err) {
// 					// the object doesn't exist -- it was likely culled
// 					// recreate it on the next time through if necessary
// 					return nil
// 				}
// 				return fmt.Errorf("Failed to get %v secret: %v", secret.Name, err)
// 			}

// 			current.Data = secret.Data
// 			if err = sdk.Update(current); err != nil {
// 				return err
// 			}
// 			return nil
// 		})
// 		if retryErr != nil {
// 			return retryErr
// 		}
// 	}

// 	return nil
// }
