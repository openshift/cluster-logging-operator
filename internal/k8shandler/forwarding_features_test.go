package k8shandler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/k8shandler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("#EvaluateAnnotationsForEnabledCapabilities", func() {

	It("should do nothing if the forwarder is nil", func() {
		options := generator.Options{}
		EvaluateAnnotationsForEnabledCapabilities(nil, options)
		Expect(options).To(BeEmpty(), "Exp no entries added to the options")
	})
	DescribeTable("when forwarder is not nil", func(enabledOption, value string, pairs ...string) {
		if len(pairs)%2 != 0 {
			Fail("Annotations must be passed as pairs to the test table")
		}
		options := generator.Options{}
		forwarder := &logging.ClusterLogForwarder{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{},
			},
		}
		for i := 0; i < len(pairs); i = i + 2 {
			key := pairs[i]
			value := pairs[i+1]
			forwarder.Annotations[key] = value
		}
		EvaluateAnnotationsForEnabledCapabilities(forwarder, options)
		if enabledOption == "" {
			Expect(options).To(BeEmpty(), "Exp. the option to be disabled")
		} else {
			Expect(options[enabledOption]).To(Equal(value), "Exp the option to equal the given value")
		}

	},
		Entry("enables TLS Security profile for enabled", PreviewTLSSecurityProfile, "", PreviewTLSSecurityProfile, "enabled"),
		Entry("enables TLS Security profile for enabled", PreviewTLSSecurityProfile, "", PreviewTLSSecurityProfile, "eNabled"),
		Entry("disables TLS Security profile for true", "", "", PreviewTLSSecurityProfile, "true"),
		Entry("enables old remote syslog for enabled", UseOldRemoteSyslogPlugin, "", UseOldRemoteSyslogPlugin, "enabled"),
		Entry("enables old remote syslog for enabled", UseOldRemoteSyslogPlugin, "", UseOldRemoteSyslogPlugin, "eNabled"),
		Entry("disables old remote syslog for true", "", "", UseOldRemoteSyslogPlugin, "true"),
		Entry("enables debug for true", helpers.EnableDebugOutput, "true", AnnotationDebugOutput, "true"),
		Entry("enables debug for True", helpers.EnableDebugOutput, "true", AnnotationDebugOutput, "True"),
		Entry("disables debug for anything else", "", "", AnnotationDebugOutput, "abcdef"),
	)

})
