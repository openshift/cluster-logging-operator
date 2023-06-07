package elasticsearch

import (
	"context"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
