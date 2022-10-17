package kibana

import (
	"context"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RemoveIfOwnedByCLO need for smooth upgrade CLO to the 5.4 version, after moving certificates generation to the EO side
// see details: https://issues.redhat.com/browse/LOG-1923
func RemoveIfOwnedByCLO(k8Client client.Client, namespace string, owner metav1.OwnerReference) (err error) {
	secret := runtime.NewSecret(namespace, constants.KibanaProxyName, nil)
	if err := k8Client.Get(context.TODO(), client.ObjectKeyFromObject(secret), secret); err != nil && !errors.IsNotFound(err) {
		return err
	}
	if utils.IsOwnedBy(secret.GetOwnerReferences(), owner) {
		if err := k8Client.Delete(context.TODO(), secret); err != nil && !errors.IsNotFound(err) {
			log.Error(err, fmt.Sprintf("Can't remove %s secret", constants.KibanaProxyName))
			return err
		}
	}

	secret = runtime.NewSecret(namespace, constants.KibanaName, nil)
	if utils.IsOwnedBy(secret.GetOwnerReferences(), owner) {
		objects := []client.Object{
			secret,
			runtime.NewService(namespace, constants.KibanaName),
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: constants.KibanaName},
			},
			runtime.NewRoute(namespace, constants.KibanaName, "", ""),
		}

		for _, obj := range objects {
			if err := k8Client.Delete(context.TODO(), obj); err != nil && !errors.IsNotFound(err) {
				log.Error(err, fmt.Sprintf("Can't remove object %s/%s/%s", obj.GetObjectKind(), obj.GetNamespace(), obj.GetName()))
				return err
			}
		}
	}
	return nil
}
