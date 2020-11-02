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

package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test"
	. "github.com/openshift/cluster-logging-operator/test/client"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

var _ = Describe("Watch", func() {
	var t *Test

	BeforeEach(func() { t = NewTest() })
	AfterEach(func() { t.Close() })

	It("ForObject watches resources", func() {
		o := runtime.NewConfigMap(t.NS.Name, "test", map[string]string{"a": "b"})
		w, err := t.WatchObject(o)
		ExpectOK(err)
		defer w.Stop()
		ExpectOK(t.Create(o))
		e := <-w.ResultChan()
		Expect(e.Type).To(Equal(watch.Added), "%v", test.YAMLString(e))
		Expect(e.Object).To(BeAssignableToTypeOf(o), "%v", test.YAMLString(e.Object))
		o2 := e.Object.(*corev1.ConfigMap)
		Expect(o2.Name).To(Equal(o.Name))
		Expect(o2.Data).To(Equal(o.Data))
		o.Data["x"] = "y"
		ExpectOK(t.Update(o))
		e = <-w.ResultChan()
		o2 = e.Object.(*corev1.ConfigMap)
		Expect(o2.Data).To(Equal(o.Data))
	})

	It("waits for a pod to be running", func() {
		pod := runtime.NewPod(t.NS.Name, "run", corev1.Container{
			Name: "testpod", Image: "docker.io/library/busybox", Args: []string{"sleep", "1h"},
		})
		ExpectOK(t.Create(pod))
		ExpectOK(t.WaitFor(pod, PodRunning), test.YAMLString(pod))
		Expect(pod.Status.Phase).To(Equal(corev1.PodRunning))
	})

	It("waits for a pod that succeeds", func() {
		pod := runtime.NewPod(t.NS.Name, "run", corev1.Container{
			Name: "testpod", Image: "docker.io/library/busybox", Args: []string{"true"},
		})
		pod.Spec.RestartPolicy = corev1.RestartPolicyNever
		ExpectOK(t.Create(pod))
		ExpectOK(t.WaitFor(pod, PodSucceeded), test.YAMLString(pod))
		Expect(pod.Status.Phase).To(Equal(corev1.PodSucceeded))
	})

	It("returns when a pod fails", func() {
		pod := runtime.NewPod(t.NS.Name, "run", corev1.Container{
			Name: "testpod", Image: "docker.io/library/busybox", Args: []string{"false"},
		})
		pod.Spec.RestartPolicy = corev1.RestartPolicyNever
		ExpectOK(t.Create(pod))
		Expect(t.WaitFor(pod, PodSucceeded)).To(MatchError(ErrWatchClosed), test.YAMLString(pod))
		Expect(pod.Status.Phase).To(Equal(corev1.PodFailed))
	})

	It("times out waiting for non-existent pod", func() {
		Expect(t.WithTimeout(test.DefaultFailureTimeout).WaitFor(runtime.NewPod(t.NS.Name, "no-such-pod"), PodRunning)).To(HaveOccurred())
	})
})
