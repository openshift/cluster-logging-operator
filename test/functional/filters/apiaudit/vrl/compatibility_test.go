package apiaudit

import (
	"encoding/json"
	"io"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
)

var _ = Describe("splunk-exporter compatibility", func() {

	It("generates the same output as git@gitlab.cee.redhat.com:service/splunk-audit-exporter.git", func() {
		// Note: test data was generated as follows:
		//	cd testdata; audit-exporter --follow=false --input audit_in.log  --policy example_policy.yaml  > audit_out_exporter.log

		// Read events from audit_in.log
		in, err := os.Open("testdata/audit_in.log")
		Expect(err).NotTo(HaveOccurred())
		cmd := vectorCmd(readPolicy("testdata/example_policy.yaml"))
		cmd.Stdin = in

		// Read actual output events from stdout.
		rGot, err := cmd.StdoutPipe()
		Expect(err).NotTo(HaveOccurred())
		Expect(cmd.Start()).To(Succeed())

		// Read expected events from stored exporter output file.
		rWant, err := os.Open("testdata/audit_out_exporter.log")
		Expect(err).NotTo(HaveOccurred())

		// Compare events
		dWant, dGot := json.NewDecoder(rWant), json.NewDecoder(rGot)
		n := 0
		defer func() { Expect(n).NotTo(Equal(0)) }() // Make sure we process some events
		for {
			var want, got auditv1.Event
			errWant, errGot := dWant.Decode(&want), dGot.Decode(&got)
			if errWant == io.EOF && errGot == io.EOF {
				return // Both readers finished at the same time, success
			}
			Expect(errGot).NotTo(HaveOccurred(), "actual events")
			Expect(errWant).NotTo(HaveOccurred(), "expected events")

			// Ignore differences in the response body if both events have one.
			// The exporter removes some response metadata, this filter does not.
			if want.ResponseObject != nil && got.ResponseObject != nil {
				want.ResponseObject, got.ResponseObject = nil, nil
			}
			Expect(test.JSONString(want)).To(Equal(test.JSONString(got)))
			n++
		}
	})
})
