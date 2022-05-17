package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/ViaQ/logerr/v2/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

func WaitForDaemonSet(t *testing.T, kubeclient kubernetes.Interface, namespace, name string, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		_, err = kubeclient.AppsV1().DaemonSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s daemonset\n", name)
				return false, nil
			}
			log.NewLogger("e2e-utils").Error(err, "Error getting Daemonsets")
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return err
	}
	t.Log("Daemonset available\n")
	return nil
}
