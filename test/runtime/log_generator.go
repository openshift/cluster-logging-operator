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

package runtime

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// NewLogGenerator creates a pod that will print `count` lines to stdout, waiting for
// `delay` between each line.  Lines are of the form "<timestamp> [n] `message`"
// where n is the number of lines output so far. Once done printing the pod will
// be idle but will not exit until deleted.
func NewLogGenerator(namespace, name string, count int, delay time.Duration, message string) *corev1.Pod {
	cmd := fmt.Sprintf(`i=0; while [ $i -lt %v ]; do echo "$(date) [$i]: %v"; i=$((i+1)); sleep %f; done; sleep infinity`, count, message, delay.Seconds())
	l := NewPod(namespace, "log-generator", corev1.Container{
		Name:    name,
		Image:   "busybox",
		Command: []string{"sh", "-c", cmd}},
	)
	l.Spec.RestartPolicy = corev1.RestartPolicyNever
	return l
}
