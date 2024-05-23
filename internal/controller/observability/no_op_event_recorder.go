package observability

import "k8s.io/apimachinery/pkg/runtime"

var (
	noOpEventRecorder = NoOpEventRecorder{}
)

type NoOpEventRecorder struct{}

func (n NoOpEventRecorder) Event(object runtime.Object, eventtype, reason, message string) {
}

func (n NoOpEventRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
}

func (n NoOpEventRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
}
