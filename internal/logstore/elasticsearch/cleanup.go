package elasticsearch

import (
	"context"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// need for smooth upgrade CLO to the 5.4 version, after moving certificates generation to the EO side
// see details: https://issues.redhat.com/browse/LOG-1923
func RemoveIfSecretOwnedByCLO(k8sClient client.Client, namespace string, ownerRef v1.OwnerReference) (err error) {
	secret := runtime.NewSecret(namespace, constants.ElasticsearchName, nil)
	key := client.ObjectKeyFromObject(secret)
	if err := k8sClient.Get(context.TODO(), key, secret); err != nil && !errors.IsNotFound(err) {
		return err
	}
	if utils.IsOwnedBy(secret.GetOwnerReferences(), ownerRef) {
		if err = Remove(k8sClient, namespace); err != nil {
			return err
		}
	}
	return nil
}

func Remove(k8sClient client.Client, namespace string) (err error) {
	secret := runtime.NewSecret(namespace, constants.ElasticsearchName, nil)
	if err = k8sClient.Delete(context.TODO(), secret); err != nil && !errors.IsNotFound(err) {
		log.V(1).Error(err, "Issue removing elasticsearch secret")
		return
	}

	if err = RemoveCustomResource(k8sClient, namespace, constants.ElasticsearchName); err != nil && !(errors.IsNotFound(err) || meta.IsNoMatchError(err)) {
		log.V(1).Error(err, "Issue removing elasticsearch CR")
		return
	}
	return nil
}

func RemoveCustomResource(k8sClient client.Client, namespace, name string) (err error) {
	cr := NewEmptyElasticsearchCR(namespace, name)
	if err = k8sClient.Delete(context.TODO(), cr); err != nil {
		return
	}
	if err != nil && !errors.IsNotFound(err) && !(errors.IsNotFound(err) || meta.IsNoMatchError(err)) {
		return fmt.Errorf("failure deleting %s/%s elasticsearch CR: %v", cr.Namespace, cr.Name, err)
	}

	return nil
}
