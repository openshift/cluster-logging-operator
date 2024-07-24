package operator

import (
	"context"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	internalruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	opv1 "github.com/operator-framework/api/pkg/operators/v1"
	opv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"strconv"
	"strings"
	"time"
)

const (
	SourceNamespaceMarketplace   = "openshift-marketplace"
	CatalogSourceRedHatOperators = "redhat-operators"

	PackageNameClusterLogging = "cluster-logging"
)

var (
	packageDeploymentmap = mapWithDefault{
		PackageNameClusterLogging: "cluster-logging-operator",
	}
)

type mapWithDefault map[string]string

func (m mapWithDefault) Name(key string) string {
	if value, found := m[key]; found {
		return value
	}
	return key
}

type OperatorDeployment struct {
	operatorGroup *opv1.OperatorGroup
	subscription  *opv1alpha1.Subscription
	test          framework.Test
}

func NewDeployment(test framework.Test, namespace, packageName, channel string) *OperatorDeployment {
	runtime.Must(opv1.AddToScheme(test.Client().Scheme()))
	runtime.Must(opv1alpha1.AddToScheme(test.Client().Scheme()))
	return &OperatorDeployment{
		test: test,
		subscription: &opv1alpha1.Subscription{
			TypeMeta: metav1.TypeMeta{
				Kind:       opv1alpha1.SubscriptionKind,
				APIVersion: opv1alpha1.SubscriptionCRDAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      packageName,
			},
			Spec: &opv1alpha1.SubscriptionSpec{
				Channel:                channel,
				Package:                packageName,
				CatalogSource:          CatalogSourceRedHatOperators,
				CatalogSourceNamespace: SourceNamespaceMarketplace,
			},
		},
		operatorGroup: &opv1.OperatorGroup{
			TypeMeta: metav1.TypeMeta{
				Kind:       opv1.OperatorGroupKind,
				APIVersion: opv1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      packageDeploymentmap.Name(packageName),
			},
			Spec: opv1.OperatorGroupSpec{},
		},
	}
}

func (od *OperatorDeployment) FromSource(sourceNamespace, source string) *OperatorDeployment {
	od.subscription.Spec.CatalogSource = source
	od.subscription.Spec.CatalogSourceNamespace = sourceNamespace
	return od
}

// UpdateSubscription updates a subscription and assumes the catalogsource and image already exist
// on the cluster.  The source, and sourceNamespace empty means the those spec remain the same
func (od *OperatorDeployment) UpdateSubscription(channel, source, sourceNamespace string) error {
	dep := internalruntime.NewDeployment(od.subscription.Namespace, packageDeploymentmap.Name(od.subscription.Spec.Package))
	if err := od.test.Client().Get(dep); err != nil {
		log.V(0).Error(err, "Error retrieving the operator deployment", "deployment", dep)
		return err
	}
	initialVersion, err := strconv.Atoi(dep.ResourceVersion)
	if err != nil {
		log.V(0).Error(err, "Error getting the deployment ResourceVersion")
		return err
	}

	if err := od.test.Client().Get(od.subscription); err != nil {
		log.V(0).Error(err, "Error retrieving the subscription", "subscription", od.subscription)
		return err
	}
	od.subscription.Spec.Channel = channel
	if source != "" {
		od.subscription.Spec.CatalogSource = source
	}
	if sourceNamespace != "" {
		od.subscription.Spec.CatalogSourceNamespace = sourceNamespace
	}
	if err := od.test.Client().Update(od.subscription); err != nil {
		log.V(0).Error(err, "Error updating the subscription", "subscription", od.subscription)
		return err
	}
	if err := wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 5*time.Minute, true, func(cxt context.Context) (done bool, err error) {
		if err := od.test.Client().Get(dep); err != nil {
			return true, err
		}
		after, err := strconv.Atoi(dep.ResourceVersion)
		if err != nil {
			return true, err
		}
		return after > initialVersion, nil
	}); err != nil {
		log.V(0).Error(err, "Error checking that deployment ResourceVersion changed")
		return err
	}

	return od.waitForDeployment()
}

// Deploy deploys an operator into an existing namespace and re-creates resources as necessary
// (i.e. subscription, operatorgroup)
func (od *OperatorDeployment) Deploy() error {
	od.test.AddCleanup(func() error {
		return od.test.Client().Delete(od.operatorGroup)
	})
	if err := od.test.Client().Recreate(od.operatorGroup); err != nil {
		log.V(0).Error(err, "unable to create the operator group", "OperatorGroup", od.operatorGroup)
		return err
	}
	od.test.AddCleanup(func() error {
		return od.test.Client().Delete(od.subscription)
	})
	if err := od.test.Client().Recreate(od.subscription); err != nil {
		log.V(0).Error(err, "unable to create the subscription", "Subscription", od.subscription)
		return err
	}
	return od.waitForDeployment()
}

func (od *OperatorDeployment) waitForDeployment() error {
	name := packageDeploymentmap.Name(od.subscription.Spec.Package)
	if err := wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 3*time.Minute, true, func(cxt context.Context) (done bool, err error) {
		if out, err := oc.Literal().From(fmt.Sprintf("oc -n %s wait deployment/%s --for=condition=available", od.subscription.Namespace, name)).Run(); err != nil {
			log.V(6).Error(err, "Error waiting on deployment", "out", out, "namespace", od.subscription.Namespace, "name", name)
			if strings.Contains(out, "NotFound") {
				return false, nil
			}
			return true, err
		}
		return true, nil
	}); err != nil {
		log.V(0).Error(err, "Operator did not deploy", "Subscription", od.subscription)
		return err
	}
	return nil
}
