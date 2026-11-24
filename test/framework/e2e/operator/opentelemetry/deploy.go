package opentelemetry

import (
	"context"
	"fmt"
	"time"

	clolog "github.com/ViaQ/logerr/v2/log/static"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Namespace is where the Red Hat build of the OpenTelemetry operator is
	// installed
	Namespace = "openshift-opentelemetry-collector"

	// controllerManagerDeployment is the operator's controller deployment
	controllerManagerDeployment = "opentelemetry-operator-controller-manager"

	// packageName is the Red Hat build of the OpenTelemetry operator package
	// in the redhat-operators catalog
	packageName = "opentelemetry-product"

	// channel is the catalog channel to subscribe to
	channel = "stable"

	retryInterval = 1 * time.Second
	timeout       = 5 * time.Minute

	operatorGroupURI = "apis/operators.coreos.com/v1/namespaces/%s/operatorgroups/%s"
	subscriptionURI  = "apis/operators.coreos.com/v1alpha1/namespaces/%s/subscriptions/%s"

	operatorGroupYaml = `
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: opentelemetry-operator-group
  namespace: openshift-opentelemetry-collector
spec:
  targetNamespaces: [ ]
  upgradeStrategy: Default
`
	subscriptionYaml = `
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: opentelemetry-product
  namespace: openshift-opentelemetry-collector
spec:
  channel: stable
  name: opentelemetry-product
  source: redhat-operators
  sourceNamespace: openshift-marketplace
  installPlanApproval: Automatic
`
)

// Deploy installs the Red Hat build of the OpenTelemetry operator, or skips
// the install when it is already deployed. Only install when the operator
// deployment is confirmed absent: reinstalling would register cleanups that
// delete an operator we did not create, so an existing deployment (e.g.
// pre-installed on the cluster or left over from a prior run) is handled
// without entering the install path
func Deploy(f *framework.E2ETestFramework) error {
	dep, err := f.KubeClient.AppsV1().Deployments(Namespace).Get(context.TODO(), controllerManagerDeployment, metav1.GetOptions{})
	switch {
	case err == nil:
		if dep.Spec.Replicas != nil && dep.Status.AvailableReplicas == *dep.Spec.Replicas {
			clolog.V(1).Info("opentelemetry operator already available; skipping install",
				"namespace", Namespace, "deployment", controllerManagerDeployment)
			return nil
		}
		// Deployment exists but is not yet fully available; wait for it rather
		// than attempting a second install.
		clolog.V(1).Info("opentelemetry operator deployment exists but not yet available; waiting",
			"namespace", Namespace, "deployment", controllerManagerDeployment)
		return f.WaitForDeployment(Namespace, controllerManagerDeployment, retryInterval, timeout)
	case !apierrors.IsNotFound(err):
		// Unexpected error looking up the deployment; do not guess that it is absent.
		return err
	}

	clolog.V(1).Info("deploying opentelemetry operator", "namespace", Namespace, "package", packageName, "channel", channel)

	// Create namespace
	f.CreateNamespace(Namespace)
	f.AddCleanup(func() error {
		return f.KubeClient.CoreV1().Namespaces().Delete(context.TODO(), Namespace, metav1.DeleteOptions{})
	})

	// Create operator group
	ogUri := fmt.Sprintf(operatorGroupURI, Namespace, "opentelemetry-operator-group")
	if err = f.Client().RESTClient().Post().
		RequestURI(ogUri).
		SetHeader("Content-Type", "application/yaml").
		Body([]byte(operatorGroupYaml)).
		Do(context.TODO()).Error(); err != nil {
		return err
	}

	// Create subscription
	subUri := fmt.Sprintf(subscriptionURI, Namespace, packageName)
	if err = f.Client().RESTClient().Post().
		RequestURI(subUri).
		SetHeader("Content-Type", "application/yaml").
		Body([]byte(subscriptionYaml)).
		Do(context.TODO()).Error(); err != nil {
		return err
	}

	return f.WaitForDeployment(Namespace, controllerManagerDeployment, retryInterval, timeout)
}
