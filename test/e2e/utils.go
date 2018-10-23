package e2e

import (
	"testing"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

func WaitForCronJob(t *testing.T, kubeclient kubernetes.Interface, namespace, name string, replicas int, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		cronjob, err := kubeclient.BatchV1beta1().CronJobs(namespace).Get(name, metav1.GetOptions{IncludeUninitialized: true})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s cronjob\n", name)
				return false, nil
			}
			return false, err
		}

		if len(cronjob.Status.Active) == replicas {
			return true, nil
		}
		t.Logf("Waiting for full availability of %s cronjob (%d/%d)\n", name, len(cronjob.Status.Active), replicas)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Cronjob available (%d/%d)\n", replicas, replicas)
	return nil
}

func WaitForDaemonSet(t *testing.T, kubeclient kubernetes.Interface, namespace, name string, replicas int, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		daemonset, err := kubeclient.AppsV1().DaemonSets(namespace).Get(name, metav1.GetOptions{IncludeUninitialized: true})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s daemonset\n", name)
				return false, nil
			}
			return false, err
		}

		if int(daemonset.Status.NumberReady) == replicas {
			return true, nil
		}
		t.Logf("Waiting for full availability of %s daemonset (%d/%d)\n", name, int(daemonset.Status.NumberReady), replicas)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Daemonset available (%d/%d)\n", replicas, replicas)
	return nil
}
