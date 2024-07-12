package serviceaccount

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Get(k8sClient client.Client, namespace, name string) (proto *corev1.ServiceAccount, err error) {
	key := types.NamespacedName{Name: name, Namespace: namespace}
	proto = &corev1.ServiceAccount{}
	if err := k8sClient.Get(context.TODO(), key, proto); err != nil {
		if errors.IsNotFound(err) {
			return nil, err
		}
		return proto, err
	}

	// Do not modify cached copy
	return proto.DeepCopy(), nil
}
