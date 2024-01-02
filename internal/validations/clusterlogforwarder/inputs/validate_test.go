package inputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("#Validate", func() {
	var (
		clfStatus *loggingv1.ClusterLogForwarderStatus
		extras    map[string]bool
		inputs    []loggingv1.InputSpec
	)

	BeforeEach(func() {
		clfStatus = &loggingv1.ClusterLogForwarderStatus{}
		extras = map[string]bool{}
	})

	It("should fail if input does not have a name", func() {
		inputs = []loggingv1.InputSpec{
			{Name: ""},
		}
		Verify(inputs, clfStatus, extras)
		Expect(clfStatus.Inputs["input_0_"]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "input must have a name"))
	})

	It("should fail if input name is one of the reserved names: application, infrastructure, audit", func() {
		inputs = []loggingv1.InputSpec{
			{Name: loggingv1.InputNameApplication},
		}
		Verify(inputs, clfStatus, extras)
		Expect(clfStatus.Inputs[loggingv1.InputNameApplication]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "input name \"application\" is reserved"))
	})
	It("should succeed if input name is one of the reserved names: application, infrastructure, audit and was migrated", func() {
		inputs = []loggingv1.InputSpec{
			{Name: loggingv1.InputNameApplication},
		}
		extras := map[string]bool{constants.MigrateInputApplication: true}
		Verify(inputs, clfStatus, extras)
		Expect(clfStatus.Inputs[loggingv1.InputNameApplication]).ToNot(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "input name \"application\" is reserved"))
	})
	It("should fail if inputspec names are not unique", func() {
		inputs = []loggingv1.InputSpec{
			{Name: "my-app-logs",
				Application: &loggingv1.Application{}},
			{Name: "my-app-logs",
				Application: &loggingv1.Application{}},
		}
		Verify(inputs, clfStatus, extras)
		Expect(clfStatus.Inputs["my-app-logs"]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "duplicate name: \"my-app-logs\""))
	})

	It("should fail when inputspec doesn't define one of application, infrastructure, audit or receiver", func() {
		inputs = []loggingv1.InputSpec{
			{Name: "my-app-logs"},
		}
		Verify(inputs, clfStatus, extras)
		Expect(clfStatus.Inputs["my-app-logs"]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "inputspec must define one or more of application, infrastructure, audit or receiver"))
	})

	It("should fail validation for invalid receiver specs", func() {
		checkReceiver := func(receiverSpec *loggingv1.ReceiverSpec, expectedErrMsg string, extras map[string]bool) {
			const inputName = `receiver`
			inputs = []loggingv1.InputSpec{
				{
					Name:     inputName,
					Receiver: receiverSpec,
				},
			}
			Verify(inputs, clfStatus, extras)
			Expect(clfStatus.Inputs[inputName]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, expectedErrMsg))
		}

		checkPortAndHTTPFormat := func(port int32, format string, expectedErrMsg string) {
			checkReceiver(
				&loggingv1.ReceiverSpec{
					Type: loggingv1.ReceiverTypeHttp,
					ReceiverTypeSpec: &loggingv1.ReceiverTypeSpec{
						HTTP: &loggingv1.HTTPReceiver{
							Port:   port,
							Format: format,
						},
					},
				},
				expectedErrMsg,
				map[string]bool{constants.VectorName: true},
			)
		}

		checkPortAndSyslogProtocol := func(port int32, protocol string, expectedErrMsg string) {
			checkReceiver(
				&loggingv1.ReceiverSpec{
					Type: loggingv1.ReceiverTypeSyslog,
					ReceiverTypeSpec: &loggingv1.ReceiverTypeSpec{
						Syslog: &loggingv1.SyslogReceiver{
							Port:     port,
							Protocol: protocol,
						},
					},
				},
				expectedErrMsg,
				map[string]bool{constants.VectorName: true},
			)
		}

		checkReceiverType := func(receiverType string, expectedErrMsg string) {
			checkReceiver(
				&loggingv1.ReceiverSpec{
					Type:             receiverType,
					ReceiverTypeSpec: &loggingv1.ReceiverTypeSpec{},
				},
				expectedErrMsg,
				map[string]bool{constants.VectorName: true},
			)
		}

		checkReceiverMismatchTypeHttp := func(expectedErrMsg string) {
			checkReceiver(
				&loggingv1.ReceiverSpec{
					Type: loggingv1.ReceiverTypeHttp,
					ReceiverTypeSpec: &loggingv1.ReceiverTypeSpec{
						Syslog: &loggingv1.SyslogReceiver{},
					},
				},
				expectedErrMsg,
				map[string]bool{constants.VectorName: true},
			)
		}

		checkReceiverMismatchTypeSyslog := func(expectedErrMsg string) {
			checkReceiver(
				&loggingv1.ReceiverSpec{
					Type: loggingv1.ReceiverTypeSyslog,
					ReceiverTypeSpec: &loggingv1.ReceiverTypeSpec{
						HTTP: &loggingv1.HTTPReceiver{},
					},
				},
				expectedErrMsg,
				map[string]bool{constants.VectorName: true},
			)
		}

		for _, port := range []int32{-1, 53, 80_000} {
			checkPortAndHTTPFormat(port, loggingv1.FormatKubeAPIAudit, `invalid port specified for HTTP receiver`)
		}
		checkPortAndHTTPFormat(8080, `no_such_format`, `invalid format specified for HTTP receiver`)
		for _, port := range []int32{-1, 53, 80_000} {
			checkPortAndSyslogProtocol(port, "tcp", `invalid port specified for Syslog receiver`)
		}
		checkReceiverMismatchTypeHttp(`mismatched Type specified for receiver, specified HTTP and have Syslog`)
		checkReceiverMismatchTypeSyslog(`mismatched Type specified for receiver, specified Syslog and have HTTP`)
		checkPortAndSyslogProtocol(10514, "http", `invalid protocol specified for Syslog receiver`)
		checkReceiverType("wrong-receiver", `invalid Type specified for receiver`)
		checkReceiver(&loggingv1.ReceiverSpec{}, `invalid ReceiverTypeSpec specified for receiver`, map[string]bool{constants.VectorName: true})
		checkReceiver(&loggingv1.ReceiverSpec{}, `ReceiverSpecs are only supported for the vector log collector`, map[string]bool{})
	})

	It("should remove all inputs if even one inputspec is invalid", func() {
		inputs = []loggingv1.InputSpec{
			{Name: "my-app-logs",
				Application: &loggingv1.Application{}},
			{Name: "invalid-input"},
		}
		Verify(inputs, clfStatus, extras)
		Expect(clfStatus.Inputs["my-app-logs"]).To(HaveCondition("Ready", true, "", ""))
		Expect(clfStatus.Inputs["invalid-input"]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "inputspec must define one or more of application, infrastructure, audit or receiver"))
	})

	It("should validate correctly with one valid input spec", func() {
		inputs = []loggingv1.InputSpec{
			{Name: "my-app-logs",
				Application: &loggingv1.Application{}},
		}
		Verify(inputs, clfStatus, extras)
		Expect(clfStatus.Inputs["my-app-logs"]).To(HaveCondition("Ready", true, "", ""))
	})

	It("should validate correctly with more than one valid input spec", func() {
		inputs = []loggingv1.InputSpec{
			{Name: "my-app-logs",
				Application: &loggingv1.Application{}},
			{Name: "my-infra-logs",
				Infrastructure: &loggingv1.Infrastructure{}},
			{Name: "my-audit-logs",
				Audit: &loggingv1.Audit{}},
		}
		Verify(inputs, clfStatus, extras)
		Expect(inputs).To(HaveLen(3))
		Expect(clfStatus.Inputs["my-app-logs"]).To(HaveCondition("Ready", true, "", ""))
		Expect(clfStatus.Inputs["my-infra-logs"]).To(HaveCondition("Ready", true, "", ""))
		Expect(clfStatus.Inputs["my-audit-logs"]).To(HaveCondition("Ready", true, "", ""))
	})

	It("should validate correctly when input spec defines non receiver sources", func() {
		inputs = []loggingv1.InputSpec{
			{Name: "all-logs",
				Application:    &loggingv1.Application{},
				Infrastructure: &loggingv1.Infrastructure{},
				Audit:          &loggingv1.Audit{},
			},
		}
		Verify(inputs, clfStatus, extras)
		Expect(clfStatus.Inputs["all-logs"]).To(HaveCondition("Ready", true, "", ""))
	})

	It("should be valid with multiple input specs, multiple input source specs", func() {
		inputs = []loggingv1.InputSpec{
			{Name: "all-logs",
				Application:    &loggingv1.Application{},
				Infrastructure: &loggingv1.Infrastructure{},
				Audit:          &loggingv1.Audit{}},
			{Name: "app-infra-logs",
				Application:    &loggingv1.Application{},
				Infrastructure: &loggingv1.Infrastructure{},
			},
		}
		Verify(inputs, clfStatus, extras)
		Expect(clfStatus.Inputs["all-logs"]).To(HaveCondition("Ready", true, "", ""))
		Expect(clfStatus.Inputs["app-infra-logs"]).To(HaveCondition("Ready", true, "", ""))
	})

	It("should fail if input spec has multiple limits defined", func() {
		inputs = []loggingv1.InputSpec{
			{
				Name: "custom-app",
				Application: &loggingv1.Application{
					ContainerLimit: &loggingv1.LimitSpec{
						MaxRecordsPerSecond: 100,
					},
					GroupLimit: &loggingv1.LimitSpec{
						MaxRecordsPerSecond: 200,
					},
				},
			},
		}
		Verify(inputs, clfStatus, extras)
		Expect(clfStatus.Inputs["custom-app"]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "inputspec must define only one of container or group limit"))
	})
	It("should be valid if input has a positive limit threshold", func() {
		inputs = []loggingv1.InputSpec{
			{
				Name: "custom-app-container-limit",
				Application: &loggingv1.Application{
					ContainerLimit: &loggingv1.LimitSpec{
						MaxRecordsPerSecond: 100,
					},
				},
			},
			{
				Name: "custom-app-group-limit",
				Application: &loggingv1.Application{
					GroupLimit: &loggingv1.LimitSpec{
						MaxRecordsPerSecond: 200,
					},
				},
			},
		}
		Verify(inputs, clfStatus, extras)
		Expect(clfStatus.Inputs["custom-app-container-limit"]).To((HaveCondition("Ready", true, "", "")))
		Expect(clfStatus.Inputs["custom-app-group-limit"]).To(HaveCondition("Ready", true, "", ""))
	})
	It("should fail if input has a negative limit threshold", func() {
		inputs = []loggingv1.InputSpec{
			{
				Name: "custom-app-container-limit",
				Application: &loggingv1.Application{
					ContainerLimit: &loggingv1.LimitSpec{
						MaxRecordsPerSecond: -100,
					},
				},
			},
			{
				Name: "custom-app-group-limit",
				Application: &loggingv1.Application{
					GroupLimit: &loggingv1.LimitSpec{
						MaxRecordsPerSecond: -200,
					},
				},
			},
		}
		Verify(inputs, clfStatus, extras)
		Expect(clfStatus.Inputs["custom-app-container-limit"]).To((HaveCondition("Ready", false, loggingv1.ReasonInvalid, "inputspec cannot have a negative limit threshold")))
		Expect(clfStatus.Inputs["custom-app-group-limit"]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "inputspec cannot have a negative limit threshold"))
	})

})
