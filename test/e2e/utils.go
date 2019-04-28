package e2e

import (
	"context"
	"testing"
	"time"

	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
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

func WaitForDaemonSet(t *testing.T, kubeclient kubernetes.Interface, namespace, name string, retryInterval, timeout time.Duration) error {
	nodes, err := kubeclient.Core().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	nodeCount := len(nodes.Items)
	err = wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		daemonset, err := kubeclient.AppsV1().DaemonSets(namespace).Get(name, metav1.GetOptions{IncludeUninitialized: true})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s daemonset\n", name)
				return false, nil
			}
			return false, err
		}
		if int(daemonset.Status.NumberReady) == nodeCount {
			return true, nil
		}
		t.Logf("Waiting for full availability of %s daemonset (%d/%d)\n", name, int(daemonset.Status.NumberReady), nodeCount)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Daemonset %s available (%d/%d)\n", name, nodeCount, nodeCount)
	return nil
}

func CheckForDaemonSetImageName(t *testing.T, kubeclient kubernetes.Interface, namespace, name string, imageName string, retryInterval, timeout time.Duration) error {
	t.Logf("Checking operator updates %s daemonset to image: %q\n", name, imageName)
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		daemonset, err := kubeclient.AppsV1().DaemonSets(namespace).Get(name, metav1.GetOptions{IncludeUninitialized: true})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s daemonset\n", name)
				return false, nil
			}
			return false, err
		}

		for _, container := range daemonset.Spec.Template.Spec.Containers {
			if imageName == container.Image {
				return true, nil
			}
		}

		t.Logf("Waiting for image change of %s daemonset (%q)\n", name, imageName)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Daemonset image found (%q)\n", imageName)
	return nil
}

func CheckForDeploymentImageName(t *testing.T, kubeclient kubernetes.Interface, namespace, name string, imageName string, retryInterval, timeout time.Duration) error {
	t.Logf("Checking operator updates %s deployment to image: %q\n", name, imageName)
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		deployment, err := kubeclient.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{IncludeUninitialized: true})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s deployment\n", name)
				return false, nil
			}
			return false, err
		}

		for _, container := range deployment.Spec.Template.Spec.Containers {
			if imageName == container.Image {
				return true, nil
			}
		}

		t.Logf("Waiting for image change of %s deployment (%q)\n", name, imageName)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Deployment image found (%q)\n", imageName)
	return nil
}

func CheckForCronJobImageName(t *testing.T, kubeclient kubernetes.Interface, namespace, name string, imageName string, retryInterval, timeout time.Duration) error {
	t.Logf("Checking operator updates %s cronjob to image: %q\n", name, imageName)
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		cronjob, err := kubeclient.BatchV1beta1().CronJobs(namespace).Get(name, metav1.GetOptions{IncludeUninitialized: true})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s cronjob\n", name)
				return false, nil
			}
			return false, err
		}

		for _, container := range cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers {
			if imageName == container.Image {
				return true, nil
			}
		}

		t.Logf("Waiting for image change of %s cronjob (%q)\n", name, imageName)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Cronjob image found (%q)\n", imageName)
	return nil
}

func CheckForElasticsearchImageName(t *testing.T, client framework.FrameworkClient, namespace, name string, imageName string, retryInterval, timeout time.Duration) error {
	t.Logf("Checking operator updates %s elasticsearch CR to image: %q\n", name, imageName)
	elasticsearch := &elasticsearch.Elasticsearch{}

	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		err = client.Get(context.Background(), dynclient.ObjectKey{
			Namespace: namespace,
			Name:      name,
		}, elasticsearch)
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s elasticsearch\n", name)
				return false, nil
			}
			return false, err
		}

		if imageName == elasticsearch.Spec.Spec.Image {
			return true, nil
		}

		t.Logf("Waiting for image change of %s elasticsearch (%q)\n", name, imageName)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Elasticsearch image found (%q)\n", imageName)
	return nil
}

func getValueFromEnvVar(envVars []core.EnvVar, name string) string {
	for _, envVar := range envVars {
		if envVar.Name == name {
			return envVar.Value
		}
	}
	return ""
}
