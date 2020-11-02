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

package certificate_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/helpers/certificate"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Certificate", func() {

	It("enables client-server TLS connection with local CA", func() {
		ca := certificate.NewCA(nil) // Self-signed CA
		cert := certificate.NewCert(ca, "localhost", net.IPv4(127, 0, 0, 1), net.IPv6loopback)
		server := cert.StartServer(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "success!") }), nil)
		defer server.Close()

		client := ca.Client(nil)
		resp, err := client.Get(server.URL)
		ExpectOK(err)
		b, err := ioutil.ReadAll(resp.Body)
		ExpectOK(err)
		Expect(strings.TrimSpace(string(b))).To(Equal("success!"))

		_, err = http.Get(server.URL) // Should fail with plain HTTP
		Expect(err).To(HaveOccurred())
	})

	It("enables TLS mutual authentication", func() {
		ca := certificate.NewCA(nil) // Self-signed CA
		serverCert := certificate.NewCert(ca)
		server := serverCert.StartServer(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "success!") }), ca)
		defer server.Close()

		clientCert := certificate.NewClient(ca)
		client := ca.Client(clientCert)
		resp, err := client.Get(server.URL)
		ExpectOK(err)
		b, err := ioutil.ReadAll(resp.Body)
		ExpectOK(err)
		Expect(strings.TrimSpace(string(b))).To(Equal("success!"))

		client = ca.Client(nil) // Should fail without mutual auth.
		_, err = client.Get(server.URL)
		Expect(err).To(HaveOccurred())

		_, err = http.Get(server.URL) // Should fail with plain HTTP
		Expect(err).To(HaveOccurred())

	})
})
