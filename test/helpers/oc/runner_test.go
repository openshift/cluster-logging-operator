//nolint:errcheck
package oc_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/format"

	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

var _ = Describe("sanitizeTokenInString", func() {
	DescribeTable("sanitizing tokens in strings",
		func(input, expected string) {
			TruncatedDiff = false
			result := oc.SanitizeTokenInString(input)
			Expect(result).To(Equal(expected))
		},
		Entry("should sanitize --from-literal=token= pattern",
			"oc create secret generic sa-token-secret --from-literal=token=eyJhbGciOiJSUzI1NiIsImtpZCI6IjEyMzQ1Njc4OTAifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6InNhLXRva2VuLXNlY3JldCJ9.signature -n openshift-logging",
			"oc create secret generic sa-token-secret --from-literal=token=[REDACTED] -n openshift-logging",
		),
		Entry("should sanitize token from multiline output",
			`
% Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                       Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0*   Trying 172.30.76.28:8443...
* TLSv1.0 (OUT), TLS header, Certificate Status (22):
} [5 bytes data]
> Authorization: Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjRod1B2Y3FaZDJTamhSdmdTaVB2WFh1MFRIdGhvQmt1ODl4VTJ0UW94U2MifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjIl0sImV4cCI6MTc3NjY5NjU3NCwiaWF0IjoxNzc2NjkyOTc0LCJpc3MiOiJodHRwczovL2t1YmVybmV0ZXMuZGVmYXVsdC5zdmMiLCJqdGkiOiJlNmUwYWQ3NS0yMWQ1LTRiMDYtOWE5My03ZDU0OWE4MjQ0NTciLCJrdWJlcm5ldGVzLmlvIjp7Im5hbWVzcGFjZSI6Im9wZW5zaGlmdC1sb2dnaW5nIiwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImNsdXN0ZXItbG9nZ2luZy1vcGVyYXRvciIsInVpZCI6ImM2OTk4ZWY5LTE0M2ItNDAxNS1iMTAwLTBlNWYyMDFmZTE2ZSJ9fSwibmJmIjoxNzc2NjkyOTc0LCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6b3BlbnNoaWZ0LWxvZ2dpbmc6Y2x1c3Rlci1sb2dnaW5nLW9wZXJhdG9yIn0.fh
`,
			`
% Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                       Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0*   Trying 172.30.76.28:8443...
* TLSv1.0 (OUT), TLS header, Certificate Status (22):
} [5 bytes data]
> Authorization: Bearer [REDACTED]
`,
		),
		Entry("should sanitize JWT token in output",
			"token: eyJhbGciOiJSUzI1NiIsImtpZCI6IjEyMzQ1Njc4OTAifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50In0.signature",
			"token: [REDACTED]",
		),
		Entry("should sanitize multiple JWT tokens",
			"old: eyJhbGciOiJSUzI1NiIsImtpZCI6IjEyMzQ1Njc4OTAifQ.eyJpc3MiOiJrdWJlcm5ldGVzIn0.sig1 new: eyJhbGciOiJSUzI1NiIsImtpZCI6IjEyMzQ1Njc4OTAifQ.eyJpc3MiOiJrdWJlcm5ldGVzIn0.sig2",
			"old: [REDACTED] new: [REDACTED]",
		),
		Entry("should not sanitize when no token present",
			"oc get pods -n openshift-logging",
			"oc get pods -n openshift-logging",
		),
		Entry("should not sanitize short strings starting with eyJ",
			"eyJ short",
			"eyJ short",
		),
		Entry("should handle empty string",
			"",
			"",
		),
	)
})

var _ = Describe("sanitizeTokensInArgs", func() {
	DescribeTable("sanitizing tokens in argument arrays",
		func(input, expected []string) {
			result := oc.SanitizeTokensInArgs(input)
			Expect(result).To(Equal(expected))
		},
		Entry("should sanitize --from-literal=token= in args",
			[]string{
				"create",
				"secret",
				"generic",
				"sa-token-secret",
				"--from-literal=token=eyJhbGciOiJSUzI1NiIsImtpZCI6IjEyMzQ1Njc4OTAifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50In0.signature",
				"-n",
				"openshift-logging",
			},
			[]string{
				"create",
				"secret",
				"generic",
				"sa-token-secret",
				"--from-literal=token=[REDACTED]",
				"-n",
				"openshift-logging",
			},
		),
		Entry("should sanitize JWT token argument",
			[]string{
				"create",
				"--raw",
				"/api/v1/token",
				"eyJhbGciOiJSUzI1NiIsImtpZCI6IjEyMzQ1Njc4OTAifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50In0.signature",
			},
			[]string{
				"create",
				"--raw",
				"/api/v1/token",
				"[REDACTED]",
			},
		),
		Entry("should not sanitize when no tokens in args",
			[]string{
				"get",
				"pods",
				"-n",
				"openshift-logging",
			},
			[]string{
				"get",
				"pods",
				"-n",
				"openshift-logging",
			},
		),
		Entry("should sanitize multiple token arguments",
			[]string{
				"create",
				"--from-literal=token=eyJhbGciOiJSUzI1NiIsImtpZCI6IjEyMzQ1Njc4OTAifQ.eyJpc3MiOiJrdWJlcm5ldGVzIn0.sig1",
				"--another-token",
				"eyJhbGciOiJSUzI1NiIsImtpZCI6IjEyMzQ1Njc4OTAifQ.eyJpc3MiOiJrdWJlcm5ldGVzIn0.sig2",
			},
			[]string{
				"create",
				"--from-literal=token=[REDACTED]",
				"--another-token",
				"[REDACTED]",
			},
		),
		Entry("should handle empty args",
			[]string{},
			[]string{},
		),
		Entry("should not sanitize short eyJ strings",
			[]string{
				"get",
				"eyJ",
			},
			[]string{
				"get",
				"eyJ",
			},
		),
	)
})
