// Licensed to Red Hat, Inc under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Red Hat, Inc licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package k8shandler

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//NewServiceAccount stubs a new instance of ServiceAccount
func NewServiceAccount(accountName string, namespace string) *core.ServiceAccount {
	return &core.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: core.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      accountName,
			Namespace: namespace,
		},
	}
}

//CreateOrUpdateServiceAccount creates or updates a ServiceAccount for logging with the given name
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateServiceAccount(name string, annotations *map[string]string) error {

	serviceAccount := NewServiceAccount(name, clusterRequest.Cluster.Namespace)
	if annotations != nil {
		if serviceAccount.GetObjectMeta().GetAnnotations() == nil {
			serviceAccount.GetObjectMeta().SetAnnotations(make(map[string]string))
		}
		for key, value := range *annotations {
			serviceAccount.GetObjectMeta().GetAnnotations()[key] = value
		}
	}

	utils.AddOwnerRefToObject(serviceAccount, utils.AsOwner(clusterRequest.Cluster))

	if err := clusterRequest.Create(serviceAccount); err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating %v serviceaccount: %v", serviceAccount.Name, err)
		}

		current := &core.ServiceAccount{}
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err = clusterRequest.Get(serviceAccount.Name, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %v serviceaccount: %v", serviceAccount.Name, err)
			}
			if annotations != nil && serviceAccount.GetObjectMeta().GetAnnotations() != nil {
				if current.GetObjectMeta().GetAnnotations() == nil {
					current.GetObjectMeta().SetAnnotations(make(map[string]string))
				}
				for key, value := range serviceAccount.GetObjectMeta().GetAnnotations() {
					current.GetObjectMeta().GetAnnotations()[key] = value
				}
			}
			if err = clusterRequest.Update(current); err != nil {
				return err
			}
			return nil
		})
		if retryErr != nil {
			return retryErr
		}
	}
	return nil

}

//RemoveServiceAccount of given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveServiceAccount(serviceAccountName string) error {

	serviceAccount := NewServiceAccount(serviceAccountName, clusterRequest.Cluster.Namespace)

	if serviceAccountName == "logcollector" {
		// remove our finalizer from the list and update it.
		serviceAccount.ObjectMeta.Finalizers = utils.RemoveString(serviceAccount.ObjectMeta.Finalizers, metav1.FinalizerDeleteDependents)
	}

	err := clusterRequest.Delete(serviceAccount)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v service account: %v", serviceAccountName, err)
	}

	return nil
}

func NewLogCollectorServiceAccountRef(uid types.UID) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion:         "v1", // apiversion for serviceaccounts/finalizers in cluster-logging.<VER>.clusterserviceversion.yaml
		Kind:               "ServiceAccount",
		Name:               "logcollector",
		UID:                uid,
		BlockOwnerDeletion: utils.GetBool(true),
		Controller:         utils.GetBool(true),
	}
}
