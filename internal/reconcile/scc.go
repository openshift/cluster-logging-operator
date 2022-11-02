package reconcile

import (
	"context"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	security "github.com/openshift/api/security/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/scc"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SecurityContextConstraints(k8Client client.Client, desired *security.SecurityContextConstraints) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &security.SecurityContextConstraints{}
		key := client.ObjectKeyFromObject(desired)
		if err := k8Client.Get(context.TODO(), key, current); err != nil {
			if errors.IsNotFound(err) {
				return k8Client.Create(context.TODO(), desired)
			}
			return fmt.Errorf("failed to get %v SecurityContextConstraints: %w", key, err)
		}

		same := false
		if same, _ = scc.AreSame(*current, *desired); same {
			log.V(3).Info("SecurityContextConstraints are the same skipping update")
			return nil
		}
		return k8Client.Update(context.TODO(), current)
	})
	return retryErr
}
