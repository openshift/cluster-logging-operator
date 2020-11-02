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
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

//NewConfigMap stubs an instance of Configmap
func NewConfigMap(configmapName string, namespace string, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmapName,
			Namespace: namespace,
		},
		Data: data,
	}
}

//CreateOrUpdateConfigMap creates a new config map resource unless it exists whereas it will update
//the existing config map if the data section changed.
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateConfigMap(configMap *corev1.ConfigMap) error {
	err := clusterRequest.Create(configMap)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating configmap: %v", err)
		}

		current := &corev1.ConfigMap{}

		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err = clusterRequest.Get(configMap.Name, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %v configmap for %q: %v", configMap.Name, clusterRequest.Cluster.Name, err)
			}

			if reflect.DeepEqual(configMap.Data, current.Data) {
				return nil
			}
			current.Data = configMap.Data

			changed := false
			// if configMap specified labels ensure that current has them...
			if len(configMap.ObjectMeta.Labels) > 0 {
				for key, val := range configMap.ObjectMeta.Labels {
					if currentVal, ok := current.ObjectMeta.Labels[key]; ok {
						if currentVal != val {
							current.ObjectMeta.Labels[key] = val
							changed = true
						}
					} else {
						current.ObjectMeta.Labels[key] = val
						changed = true
					}
				}
			} else {
				return nil
			}
			if !changed {
				// shortcut updating -- we didn't change anything
				return nil
			}

			return clusterRequest.Update(current)
		})
		return retryErr
	}
	return nil
}

//RemoveConfigMap with a given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveConfigMap(configmapName string) error {

	configMap := NewConfigMap(
		configmapName,
		clusterRequest.Cluster.Namespace,
		map[string]string{},
	)

	err := clusterRequest.Delete(configMap)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v configmap: %v", configmapName, err)
	}

	return nil
}
