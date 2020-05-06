package k8shandler

import (
	"testing"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"

	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//TODO: Remove this in the next release after removing old kibana code completely
func TestHasCLORef(t *testing.T) {
	clr := ClusterLoggingRequest{
		client: nil,
		cluster: &logging.ClusterLogging{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:                       "cluster-logging",
				GenerateName:               "",
				Namespace:                  "",
				SelfLink:                   "",
				UID:                        "123",
				ResourceVersion:            "",
				Generation:                 0,
				CreationTimestamp:          metav1.Time{},
				DeletionTimestamp:          nil,
				DeletionGracePeriodSeconds: nil,
				Labels:                     nil,
				Annotations:                nil,
				OwnerReferences:            nil,
				Initializers:               nil,
				Finalizers:                 nil,
				ClusterName:                "",
			},
			Spec:   logging.ClusterLoggingSpec{},
			Status: logging.ClusterLoggingStatus{},
		},
		ForwarderSpec: logging.ClusterLogForwarderSpec{},
		Collector:     nil,
	}

	obj := &apps.Deployment{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:                       "test-deployment",
			GenerateName:               "",
			Namespace:                  "",
			SelfLink:                   "",
			UID:                        "",
			ResourceVersion:            "",
			Generation:                 0,
			CreationTimestamp:          metav1.Time{},
			DeletionTimestamp:          nil,
			DeletionGracePeriodSeconds: nil,
			Labels:                     nil,
			Annotations:                nil,
			OwnerReferences:            nil,
			Initializers:               nil,
			Finalizers:                 nil,
			ClusterName:                "",
		},
		Spec:   apps.DeploymentSpec{},
		Status: apps.DeploymentStatus{},
	}

	utils.AddOwnerRefToObject(obj, utils.AsOwner(clr.cluster))

	t.Log("refs:", obj.GetOwnerReferences())
	if !HasCLORef(obj, &clr) {
		t.Error("no owner reference found but it should be found")
	}
}

func TestAreRefsEqual(t *testing.T) {
	r1 := metav1.OwnerReference{
		APIVersion: logging.SchemeGroupVersion.String(),
		Kind:       "ClusterLogging",
		Name:       "testRef",
		Controller: func() *bool { t := true; return &t }(),
	}

	r2 := metav1.OwnerReference{
		APIVersion: logging.SchemeGroupVersion.String(),
		Kind:       "ClusterLogging",
		Name:       "testRef",
		Controller: func() *bool { t := true; return &t }(),
	}

	if !AreRefsEqual(&r1, &r2) {
		t.Error("refs are not equal but they should be")
	}
}
