// Tests ported from splunk-audit-exporter

package apiaudit

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	. "k8s.io/apiserver/pkg/apis/audit/v1"
)

var TestEvent1 Event = Event{
	Verb:       "create",
	RequestURI: "/api/v1/namespaces/default/pods/test/exec",
	Level:      LevelRequestResponse,
	AuditID:    "00000000-0000-0000-0000000000000",
	ObjectRef: &ObjectReference{
		Resource:    "pods",
		Namespace:   "default",
		Subresource: "exec",
	},
	StageTimestamp: metav1.NewMicroTime(time.Now().Add(-1 * time.Hour)),
	User: authv1.UserInfo{
		UID:      "",
		Extra:    map[string]authv1.ExtraValue{},
		Username: "user",
		Groups:   []string{"system:authenticated"},
	},
	TypeMeta: metav1.TypeMeta{
		Kind:       "Event",
		APIVersion: "audit.k8s.io/v1",
	},
}

var TestEvent2 Event = Event{
	Verb:       "get",
	RequestURI: "/openapi/v2/test?timeout=30",
	Level:      LevelMetadata,
	AuditID:    "",
	ResponseStatus: &metav1.Status{
		Status: "",
		Code:   200,
	},
	ObjectRef:                nil,
	Stage:                    StageResponseComplete,
	SourceIPs:                []string{},
	StageTimestamp:           metav1.NewMicroTime(time.Now().Add(1 * time.Hour)),
	Annotations:              map[string]string{},
	RequestReceivedTimestamp: metav1.NowMicro(),
	UserAgent:                "",
	ImpersonatedUser:         &authv1.UserInfo{},
	RequestObject:            nil,
	ResponseObject:           nil,
	User: authv1.UserInfo{
		UID:      "",
		Extra:    map[string]authv1.ExtraValue{},
		Username: "user",
		Groups:   []string{"system:authenticated"},
	},
	TypeMeta: metav1.TypeMeta{
		Kind:       "Event",
		APIVersion: "k8s.io/v1",
	},
}

var TestEventServiceAccountReadOnly = Event{
	Verb:       "list",
	RequestURI: "/api/v1/namespaces/default/pods",
	Level:      LevelMetadata,
	AuditID:    "00000000-0000-0000-0000000000000",
	ObjectRef: &ObjectReference{
		Resource:  "pods",
		Namespace: "default",
	},
	User: authv1.UserInfo{
		UID:      "",
		Extra:    map[string]authv1.ExtraValue{},
		Username: "system:serviceaccount:default:test",
		Groups:   []string{"system:serviceaccounts"},
	},
	ResponseStatus: &metav1.Status{
		Status: "",
		Code:   200,
	},
	TypeMeta: metav1.TypeMeta{
		Kind:       "Event",
		APIVersion: "audit.k8s.io/v1",
	},
}

var TestEventResponseTooManyRequests Event = Event{
	Verb:        "list",
	RequestURI:  "/api/v1/namespaces/default/pods",
	Level:       LevelMetadata,
	Annotations: map[string]string{},
	AuditID:     "00000000-0000-0000-0000000000000",
	ObjectRef: &ObjectReference{
		Resource:  "pods",
		Namespace: "default",
	},
	RequestReceivedTimestamp: metav1.NowMicro(),
	User: authv1.UserInfo{
		UID:      "",
		Extra:    map[string]authv1.ExtraValue{},
		Username: "system:serviceaccount:default:test",
		Groups:   []string{"system:serviceaccounts"},
	},
	ResponseStatus: &metav1.Status{
		Status: "",
		Code:   429,
	},
	TypeMeta: metav1.TypeMeta{
		Kind:       "Event",
		APIVersion: "audit.k8s.io/v1",
	},
}

var TestEventServiceAccountLocalWrite Event = Event{
	Verb:       "patch",
	RequestURI: "/api/v1/namespaces/default/configmaps",
	Level:      LevelMetadata,
	AuditID:    "00000000-0000-0000-0000000000000",
	User: authv1.UserInfo{
		UID:      "",
		Extra:    map[string]authv1.ExtraValue{},
		Username: "system:serviceaccount:openshift-console-operator:console-operator",
		Groups:   []string{"system:serviceaccounts"},
	},
	ObjectRef: &ObjectReference{
		Name:       "console",
		APIGroup:   "route.openshift.io",
		APIVersion: "v1",
		Resource:   "routes",
		Namespace:  "openshift-console",
	},
	ResponseStatus: &metav1.Status{
		Status: "",
		Code:   200,
	},
	TypeMeta: metav1.TypeMeta{
		Kind:       "Event",
		APIVersion: "audit.k8s.io/v1",
	},
}

var TestEventServiceAccountNonlocalWrite Event = Event{
	Verb:       "update",
	RequestURI: "/api/v1/namespaces/kube-system/configmaps/root-ca",
	Level:      LevelRequestResponse,
	AuditID:    "00000000-0000-0000-0000000000000",
	User: authv1.UserInfo{
		UID:      "",
		Extra:    map[string]authv1.ExtraValue{},
		Username: "system:kube-controller-manager",
		Groups:   []string{"system:masters", "system:serviceaccounts"},
	},
	ObjectRef: &ObjectReference{
		Name:       "kube-controller-manager",
		APIGroup:   "",
		APIVersion: "v1",
		Resource:   "configmaps",
		Namespace:  "kube-system",
	},
	RequestObject: &runtime.Unknown{
		TypeMeta: runtime.TypeMeta{
			APIVersion: "v1",
			Kind:       "configmaps",
		},
		Raw:         []byte(`{"kind":"ConfigMap","apiVersion":"v1","metadata":{"name":"kube-controller-manager","namespace":"kube-system","uid":"64943c77-ff75-4a8c-8e2a-3eb0872e6a73","resourceVersion":"8467256","creationTimestamp":"2021-05-13T00:19:15Z","annotations":{"control-plane.alpha.kubernetes.io/leader":"{\"holderIdentity\":\"ip-10-0-239-24_0ad3680c-3534-4daa-98da-d68a19b7e66b\",\"leaseDurationSeconds\":15,\"acquireTime\":\"2021-05-15T07:37:22Z\",\"renewTime\":\"2021-05-24T04:47:26Z\",\"leaderTransitions\":12}"}}}`),
		ContentType: runtime.ContentTypeJSON,
	},
	ResponseStatus: &metav1.Status{
		Status: "",
		Code:   200,
	},
	Annotations: map[string]string{},
	TypeMeta: metav1.TypeMeta{
		Kind:       "Event",
		APIVersion: "audit.k8s.io/v1",
	},
}

var TestEventServiceAccountNonlocalWriteUpdate Event = Event{
	Verb:       "update",
	RequestURI: "/api/v1/namespaces/kube-system/configmaps/root-ca",
	Level:      LevelRequestResponse,
	AuditID:    "00000000-0000-0000-0000000000000",
	User: authv1.UserInfo{
		UID:      "",
		Extra:    map[string]authv1.ExtraValue{},
		Username: "system:kube-controller-manager",
		Groups:   []string{"system:masters"},
	},
	ObjectRef: &ObjectReference{
		Name:       "kube-controller-manager",
		APIGroup:   "",
		APIVersion: "v1",
		Resource:   "configmaps",
		Namespace:  "kube-system",
	},
	RequestObject: &runtime.Unknown{
		TypeMeta: runtime.TypeMeta{
			APIVersion: "v1",
			Kind:       "configmaps",
		},
		Raw:         []byte(`{"kind":"ConfigMap","apiVersion":"v1","metadata":{"name":"kube-controller-manager","namespace":"kube-system","uid":"64943c77-ff75-4a8c-8e2a-3eb0872e6a73","resourceVersion":"8467256","creationTimestamp":"2021-05-13T00:19:15Z","annotations":{"control-plane.alpha.kubernetes.io/leader":"{\"holderIdentity\":\"ip-10-0-239-24_0ad3680c-3534-4daa-98da-d68a19b7e66b\",\"leaseDurationSeconds\":15,\"acquireTime\":\"2021-05-15T07:37:22Z\",\"renewTime\":\"2021-05-24T04:47:26Z\",\"leaderTransitions\":13}"}}}`),
		ContentType: runtime.ContentTypeJSON,
	},
	ResponseStatus: &metav1.Status{
		Status: "",
		Code:   200,
	},
	Annotations: map[string]string{},
	TypeMeta: metav1.TypeMeta{
		Kind:       "Event",
		APIVersion: "audit.k8s.io/v1",
	},
}

var TestEventWithAnnotations Event = Event{
	Verb:       "update",
	RequestURI: "/api/v1/namespaces/kube-system/configmaps/root-ca",
	Level:      LevelRequestResponse,
	AuditID:    "00000000-0000-0000-0000000000000",
	User: authv1.UserInfo{
		UID:      "",
		Extra:    map[string]authv1.ExtraValue{},
		Username: "system:kube-controller-manager",
		Groups:   []string{"system:masters"},
	},
	ObjectRef: &ObjectReference{
		Name:       "kube-controller-manager",
		APIGroup:   "",
		APIVersion: "v1",
		Resource:   "configmaps",
		Namespace:  "kube-system",
	},
	RequestObject: &runtime.Unknown{},
	ResponseStatus: &metav1.Status{
		Status: "",
		Code:   200,
	},
	Annotations: map[string]string{},
	TypeMeta: metav1.TypeMeta{
		Kind:       "Event",
		APIVersion: "audit.k8s.io/v1",
	},
}

var _ = Describe("splunk-exporter equivalent tests", func() {

	Context("DefaultFilter", func() {
		It("should filter events", func() {
			filter := &obs.KubeApiAudit{}
			Expect(Filtered(filter, TestEvent1)).ShouldNot(BeNil())
			Expect(Filtered(filter, TestEvent2)).ShouldNot(BeNil())
			Expect(Filtered(filter, TestEventServiceAccountReadOnly)).Should(BeNil())
			Expect(Filtered(filter, TestEventServiceAccountLocalWrite)).Should(BeNil())
			Expect(Filtered(filter, TestEventServiceAccountNonlocalWrite)).ShouldNot(BeNil())
		})
	})

	Context("DropResponseStatusCodes", func() {
		It("should drop based on response codes", func() {
			filter := &obs.KubeApiAudit{OmitResponseCodes: &[]int{429, 422}}
			Expect(Filtered(filter, TestEventResponseTooManyRequests)).Should(BeNil())
			Expect(Filtered(filter, TestEvent1)).ShouldNot(BeNil())
		})
	})
})
