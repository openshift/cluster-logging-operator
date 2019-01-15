package k8shandler

import (
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterLogging struct {
	*logging.ClusterLogging
}

func NewClusterLogging(cluster *logging.ClusterLogging) *ClusterLogging {
	return &ClusterLogging{cluster}
}

func (cluster *ClusterLogging) Type() metav1.TypeMeta {
	return cluster.TypeMeta
}
func (cluster *ClusterLogging) Meta() metav1.ObjectMeta {
	return cluster.ObjectMeta
}

func (cluster *ClusterLogging) SchemeGroupVersion() string {
	return logging.SchemeGroupVersion.String()
}

//AddOwnerRefTo adds an ownerRef from this ClusterLogging instance to the given object
func (cluster *ClusterLogging) AddOwnerRefTo(object metav1.Object) {
	ownerRef := utils.AsOwner(cluster)
	utils.AddOwnerRefToObject(object, ownerRef)
}

//CreateOrUpdateServiceAccount creates or updates a ServiceAccount for logging with the given name
func (logging *ClusterLogging) CreateOrUpdateServiceAccount(name string) error {

	serviceAccount := utils.NewServiceAccount(name, logging.Namespace)

	logging.AddOwnerRefTo(serviceAccount)

	err := sdk.Create(serviceAccount)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating %s service account: %v", name, err)
	}

	return nil
}

//Exists checks the api server for this ClusterLogging
//returns true if it does; false otherwise
func (cluster *ClusterLogging) Exists() bool {
	clone := cluster.DeepCopy()
	if err := sdk.Get(clone); err != nil {
		if apierrors.IsNotFound(err) {
			return false
		}
		logrus.Errorf("Failed to check for ClusterLogging object: %v", err)
		return false
	}

	return true
}
