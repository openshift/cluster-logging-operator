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
	"reflect"
	"testing"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/google/go-cmp/cmp"
)

func TestClusterLoggingRequest_CreateOrUpdateSecretNoop(t *testing.T) {
	// mimic ca.crt
	caCrtStr := `-----BEGIN CERTIFICATE-----
This is a ca cert.
-----END CERTIFICATE-----`

	// mimic ca.key
	caKeyStr := `-----BEGIN PRIVATE KEY-----
This is a private key.
-----END PRIVATE KEY-----`

	initSecret := NewSecret(
		"test-secret",
		"test-namespace",
		map[string][]byte{
			"testca":  []byte(caCrtStr),
			"testkey": []byte(caKeyStr),
		})

	clusterLoggingRequest := &ClusterLoggingRequest{
		Client: fake.NewFakeClient(),
		Cluster: &logging.ClusterLogging{
			ObjectMeta: v1.ObjectMeta{
				Name:      "instance",
				Namespace: "test-namespace",
			},
		},
	}

	// create
	if err := clusterLoggingRequest.CreateOrUpdateSecret(initSecret); err != nil {
		t.Errorf("CreateOrUpdateSecret failed 1 [%v]", err)
	}

	firstSecret, err := clusterLoggingRequest.GetSecret("test-secret")
	if err != nil {
		t.Fatalf("GetSecret failed 1 [%v]", err)
	}

	// update
	// Notice: This is indented by breaking change:
	// https://github.com/kubernetes-sigs/controller-runtime/pull/832
	firstSecret.ObjectMeta.ResourceVersion = ""
	if err := clusterLoggingRequest.CreateOrUpdateSecret(firstSecret); err != nil {
		t.Errorf("CreateOrUpdateSecret failed 2 [%v]", err)
	}

	secondSecret, err := clusterLoggingRequest.GetSecret("test-secret")
	if err != nil {
		t.Errorf("GetSecret failed 2 [%v]", err)
	}

	diff := cmp.Diff(firstSecret, secondSecret)
	if diff == "" {
		t.Logf("Initial secret [%p]\n%v", initSecret, initSecret)
		t.Logf("First secret [%p]\n%v", firstSecret, firstSecret)
		t.Logf("Second secret [%p]\n%v", secondSecret, secondSecret)
	} else {
		t.Errorf("First secret != Second secret:\n %s", diff)
	}
}

func TestClusterLoggingRequest_CreateOrUpdateSecretSameKeysNewValues(t *testing.T) {
	// mimic ca.crt
	caCrtStr := `-----BEGIN CERTIFICATE-----
This is a ca cert.
-----END CERTIFICATE-----`

	// mimic ca.key
	caKeyStr := `-----BEGIN PRIVATE KEY-----
This is a private key.
-----END PRIVATE KEY-----`

	initSecret := NewSecret(
		"test-secret",
		"test-namespace",
		map[string][]byte{
			"testca":  []byte(caCrtStr),
			"testkey": []byte(caKeyStr),
		})

	clusterLoggingRequest := &ClusterLoggingRequest{
		Client: fake.NewFakeClient(),
		Cluster: &logging.ClusterLogging{
			ObjectMeta: v1.ObjectMeta{
				Name:      "instance",
				Namespace: "test-namespace",
			},
		},
	}

	// create
	if err := clusterLoggingRequest.CreateOrUpdateSecret(initSecret); err != nil {
		t.Errorf("CreateOrUpdateSecret failed 1 [%v]", err)
	}

	firstSecret, err := clusterLoggingRequest.GetSecret("test-secret")
	if err != nil {
		t.Errorf("GetSecret failed 1 [%v]", err)
	}

	caCrtStr = `-----BEGIN CERTIFICATE-----
This is a new ca cert.
-----END CERTIFICATE-----`

	caKeyStr = `-----BEGIN PRIVATE KEY-----
This is a new private key.
-----END PRIVATE KEY-----`

	newSecret := NewSecret(
		"test-secret",
		"test-namespace",
		map[string][]byte{
			"testca":  []byte(caCrtStr),
			"testkey": []byte(caKeyStr),
		})

	// update
	if err := clusterLoggingRequest.CreateOrUpdateSecret(newSecret); err != nil {
		t.Errorf("CreateOrUpdateSecret failed 2 [%v]", err)
	}

	secondSecret, err := clusterLoggingRequest.GetSecret("test-secret")
	if err != nil {
		t.Errorf("GetSecret failed 2 [%v]", err)
	}

	if reflect.DeepEqual(firstSecret, secondSecret) {
		t.Errorf("First secret [%v] == Second secret [%v]", firstSecret, secondSecret)
	} else {
		t.Logf("Initial secret [%p]\n%v", initSecret, initSecret)
		t.Logf("Initial secret [%p]\n%v", newSecret, newSecret)
		t.Logf("First secret [%p]\n%v", firstSecret, firstSecret)
		t.Logf("Second secret [%p]\n%v", secondSecret, secondSecret)
	}
}

func TestClusterLoggingRequest_CreateOrUpdateSecretNewKeysSameValues(t *testing.T) {
	// mimic ca.crt
	caCrtStr := `-----BEGIN CERTIFICATE-----
This is a ca cert.
-----END CERTIFICATE-----`

	// mimic ca.key
	caKeyStr := `-----BEGIN PRIVATE KEY-----
This is a private key.
-----END PRIVATE KEY-----`

	initSecret := NewSecret(
		"test-secret",
		"test-namespace",
		map[string][]byte{
			"testca":  []byte(caCrtStr),
			"testkey": []byte(caKeyStr),
		})

	clusterLoggingRequest := &ClusterLoggingRequest{
		Client: fake.NewFakeClient(),
		Cluster: &logging.ClusterLogging{
			ObjectMeta: v1.ObjectMeta{
				Name:      "instance",
				Namespace: "test-namespace",
			},
		},
	}

	// create
	if err := clusterLoggingRequest.CreateOrUpdateSecret(initSecret); err != nil {
		t.Errorf("CreateOrUpdateSecret failed 1 [%v]", err)
	}

	firstSecret, err := clusterLoggingRequest.GetSecret("test-secret")
	if err != nil {
		t.Errorf("GetSecret failed 1 [%v]", err)
	}

	newSecret := NewSecret(
		"test-secret",
		"test-namespace",
		map[string][]byte{
			"newtestca":  []byte(caCrtStr),
			"newtestkey": []byte(caKeyStr),
		})

	// update
	if err := clusterLoggingRequest.CreateOrUpdateSecret(newSecret); err != nil {
		t.Errorf("CreateOrUpdateSecret failed 2 [%v]", err)
	}

	secondSecret, err := clusterLoggingRequest.GetSecret("test-secret")
	if err != nil {
		t.Errorf("GetSecret failed 2 [%v]", err)
	}

	if reflect.DeepEqual(firstSecret, secondSecret) {
		t.Errorf("First secret [%v] == Second secret [%v]", firstSecret, secondSecret)
	} else {
		t.Logf("Initial secret [%p]\n%v", initSecret, initSecret)
		t.Logf("Initial secret [%p]\n%v", newSecret, newSecret)
		t.Logf("First secret [%p]\n%v", firstSecret, firstSecret)
		t.Logf("Second secret [%p]\n%v", secondSecret, secondSecret)
	}
}

func TestClusterLoggingRequest_CreateOrUpdateSecretRemovingKey(t *testing.T) {
	// mimic ca.crt
	caCrtStr := `-----BEGIN CERTIFICATE-----
This is a ca cert.
-----END CERTIFICATE-----`

	// mimic ca.key
	caKeyStr := `-----BEGIN PRIVATE KEY-----
This is a private key.
-----END PRIVATE KEY-----`

	initSecret := NewSecret(
		"test-secret",
		"test-namespace",
		map[string][]byte{
			"testca":  []byte(caCrtStr),
			"testkey": []byte(caKeyStr),
		})

	clusterLoggingRequest := &ClusterLoggingRequest{
		Client: fake.NewFakeClient(),
		Cluster: &logging.ClusterLogging{
			ObjectMeta: v1.ObjectMeta{
				Name:      "instance",
				Namespace: "test-namespace",
			},
		},
	}

	// create
	if err := clusterLoggingRequest.CreateOrUpdateSecret(initSecret); err != nil {
		t.Errorf("CreateOrUpdateSecret failed 1 [%v]", err)
	}

	firstSecret, err := clusterLoggingRequest.GetSecret("test-secret")
	if err != nil {
		t.Errorf("GetSecret failed 1 [%v]", err)
	}

	newSecret := NewSecret(
		"test-secret",
		"test-namespace",
		map[string][]byte{
			"testca": []byte(caCrtStr),
		})

	// update
	if err := clusterLoggingRequest.CreateOrUpdateSecret(newSecret); err != nil {
		t.Errorf("CreateOrUpdateSecret failed 2 [%v]", err)
	}

	secondSecret, err := clusterLoggingRequest.GetSecret("test-secret")
	if err != nil {
		t.Errorf("GetSecret failed 2 [%v]", err)
	}

	if reflect.DeepEqual(firstSecret, secondSecret) {
		t.Errorf("First secret [%v] == Second secret [%v]", firstSecret, secondSecret)
	} else {
		t.Logf("Initial secret [%p]\n%v", initSecret, initSecret)
		t.Logf("Initial secret [%p]\n%v", newSecret, newSecret)
		t.Logf("First secret [%p]\n%v", firstSecret, firstSecret)
		t.Logf("Second secret [%p]\n%v", secondSecret, secondSecret)
	}
}
