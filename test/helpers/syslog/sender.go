package syslog

import (
	_ "embed"
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
)

const (
	sender = "sender"
	image  = "quay.io/quay/busybox"
)

func AddSenderContainer(pb *runtime.PodBuilder) error {
	pb.AddContainer(sender, image).WithCmd([]string{"sh", "-c", "sleep infinity"}).End()
	return nil
}

func WriteToSyslogInputWithNetcat(framework *functional.CollectorFunctionalFramework, inputName, msg string) error {
	for _, input := range framework.Forwarder.Spec.Inputs {
		if input.Type == obs.InputTypeReceiver && input.Receiver != nil && input.Receiver.Type == obs.ReceiverTypeSyslog && input.Name == inputName {
			cmd := fmt.Sprintf("echo %q | nc 127.0.0.1 %d", msg, input.Receiver.Port)
			_, err := framework.RunCommand(sender, "sh", "-c", cmd)
			return err
		}
	}
	return fmt.Errorf("WriteToHttpInput: no HTTP input named %s", inputName)
}
