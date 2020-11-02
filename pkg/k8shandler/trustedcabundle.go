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

	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

//createOrGetTrustedCABundleConfigMap creates or returns an existing Trusted CA Bundle ConfigMap.
//By setting label "config.openshift.io/inject-trusted-cabundle: true", the cert is automatically filled/updated.
func (clusterRequest *ClusterLoggingRequest) createOrGetTrustedCABundleConfigMap(name string) (*corev1.ConfigMap, error) {
	configMap := NewConfigMap(
		name,
		clusterRequest.Cluster.Namespace,
		map[string]string{
			constants.TrustedCABundleKey: "",
		},
	)
	configMap.ObjectMeta.Labels = make(map[string]string)
	configMap.ObjectMeta.Labels[constants.InjectTrustedCABundleLabel] = "true"

	utils.AddOwnerRefToObject(configMap, utils.AsOwner(clusterRequest.Cluster))

	err := clusterRequest.Create(configMap)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("failed to create trusted CA bundle config map %q for %q: %s", name, clusterRequest.Cluster.Name, err)
		}

		// Get the existing config map which may include an injected CA bundle
		if err = clusterRequest.Get(configMap.Name, configMap); err != nil {
			if errors.IsNotFound(err) {
				// the object doesn't exist -- it was likely culled
				// recreate it on the next time through if necessary
				return nil, fmt.Errorf("failed to find trusted CA bundle config map %q for %q: %s", name, clusterRequest.Cluster.Name, err)
			}
			return nil, fmt.Errorf("failed to get trusted CA bundle config map %q for %q: %s", name, clusterRequest.Cluster.Name, err)
		}
	}
	return configMap, err
}

func hasTrustedCABundle(configMap *corev1.ConfigMap) bool {
	if configMap == nil {
		return false
	}
	caBundle, ok := configMap.Data[constants.TrustedCABundleKey]
	return ok && caBundle != ""
}

func calcTrustedCAHashValue(configMap *corev1.ConfigMap) (string, error) {
	hashValue := ""
	var err error

	if configMap == nil {
		return hashValue, nil
	}
	caBundle, ok := configMap.Data[constants.TrustedCABundleKey]
	if ok && caBundle != "" {
		hashValue, err = utils.CalculateMD5Hash(caBundle)
		if err != nil {
			return "", err
		}
	}

	if !ok {
		return "", fmt.Errorf("Expected key %v does not exist in %v", constants.TrustedCABundleKey, configMap.Name)
	}

	return hashValue, nil
}
