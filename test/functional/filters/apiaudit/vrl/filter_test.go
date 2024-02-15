package apiaudit

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	authv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "k8s.io/apiserver/pkg/apis/audit/v1"
)

var _ = Describe("Policy to VRL Filter", func() {

	Context("omit response codes", func() {
		It("should omit specified codes", func() {
			p := &loggingv1.KubeAPIAudit{OmitResponseCodes: &[]int{1234}}
			Expect(Filtered(p, Event{ResponseStatus: &v1.Status{Code: 1234}})).To(BeNil())
			Expect(Filtered(p, Event{ResponseStatus: &v1.Status{Code: 5678}})).NotTo(BeNil())
		})
		It("should omit default codes if missing", func() {
			p := &loggingv1.KubeAPIAudit{}
			Expect(Filtered(p, Event{ResponseStatus: &v1.Status{Code: 409}})).To(BeNil())
			Expect(Filtered(p, Event{ResponseStatus: &v1.Status{Code: 200}})).NotTo(BeNil())
		})
		It("should not omit by code if explictly empty", func() {
			p := &loggingv1.KubeAPIAudit{OmitResponseCodes: &[]int{}} // Explictly Empty
			Expect(Filtered(p, Event{ResponseStatus: &v1.Status{Code: 409}})).NotTo(BeNil())
			Expect(Filtered(p, Event{ResponseStatus: &v1.Status{Code: 200}})).NotTo(BeNil())
		})
	})

	Context("omit stages", func() {
		It("omits using policy setting", func() {
			p := &loggingv1.KubeAPIAudit{OmitStages: []Stage{StageRequestReceived}}
			Expect(Filtered(p, Event{Stage: StageRequestReceived})).To(BeNil())
			Expect(Filtered(p, Event{Stage: StageResponseStarted})).NotTo(BeNil())
		})
		It("omits using rule setting", func() {
			p := &loggingv1.KubeAPIAudit{Rules: []PolicyRule{
				{Verbs: []string{"update"}, OmitStages: []Stage{StagePanic}},
			}}
			Expect(Filtered(p, Event{Verb: "update", Stage: StagePanic})).To(BeNil())
			Expect(Filtered(p, Event{Verb: "create", Stage: StagePanic})).NotTo(BeNil())
		})
		It("omits using rule and policy setting", func() {
			p := &loggingv1.KubeAPIAudit{
				OmitStages: []Stage{StageRequestReceived},
				Rules:      []PolicyRule{{Verbs: []string{"update"}, OmitStages: []Stage{StagePanic}}}}
			Expect(Filtered(p, Event{Verb: "create", Stage: StageRequestReceived})).To(BeNil())
			Expect(Filtered(p, Event{Verb: "update", Stage: StageRequestReceived})).To(BeNil())
			Expect(Filtered(p, Event{Verb: "update", Stage: StagePanic})).To(BeNil())
			Expect(Filtered(p, Event{Verb: "create", Stage: StagePanic})).NotTo(BeNil())
		})
	})

	Context("policy rules", func() {
		It("matches by verb", func() {
			p := &loggingv1.KubeAPIAudit{
				Rules: []PolicyRule{
					{Level: LevelNone, Verbs: []string{"watch", "patch"}},
					{Level: LevelRequest, Verbs: []string{"update", "delete"}},
					{Level: LevelMetadata, Verbs: []string{"get"}},
				},
			}
			Expect(Filtered(p, Event{Verb: "watch"})).To(HaveLevel(LevelNone))
			Expect(Filtered(p, Event{Verb: "patch"})).To(HaveLevel(LevelNone))
			Expect(Filtered(p, Event{Verb: "get"})).To(HaveLevel(LevelMetadata))
			Expect(Filtered(p, Event{Verb: "update"})).To(HaveLevel(LevelRequest))
			Expect(Filtered(p, Event{Verb: "delete"})).To(HaveLevel(LevelRequest))
		})

		It("matches by user", func() {
			p := &loggingv1.KubeAPIAudit{
				Rules: []PolicyRule{
					{Level: LevelNone, Users: []string{"barney"}},
					{Level: LevelRequestResponse, Users: []string{"fred"}},
					{Level: LevelRequest, Users: []string{"pebble*", "*bam"}},
					{Level: LevelMetadata, Users: nil},
				},
			}
			Expect(Filtered(p, Event{User: authv1.UserInfo{Username: "barney"}})).To(HaveLevel(LevelNone))
			Expect(Filtered(p, Event{User: authv1.UserInfo{Username: "fred"}})).To(HaveLevel(LevelRequestResponse))
			Expect(Filtered(p, Event{User: authv1.UserInfo{Username: "pebbles"}})).To(HaveLevel(LevelRequest))
			Expect(Filtered(p, Event{User: authv1.UserInfo{Username: "bambam"}})).To(HaveLevel(LevelRequest))
			Expect(Filtered(p, Event{User: authv1.UserInfo{Username: "wilma"}})).To(HaveLevel(LevelMetadata))
		})

		It("stops on first match", func() {
			p := &loggingv1.KubeAPIAudit{
				Rules: []PolicyRule{
					{Level: LevelRequest, Users: []string{"barney"}},
					{Level: LevelNone, Verbs: []string{"get"}},
				},
			}
			Expect(Filtered(p, Event{Verb: "get", User: authv1.UserInfo{Username: "barney"}})).To(HaveLevel(LevelRequest))
			Expect(Filtered(p, Event{Verb: "get"})).To(HaveLevel(LevelNone))
		})

		It("matches by group", func() {
			p := &loggingv1.KubeAPIAudit{
				Rules: []PolicyRule{
					{Level: LevelMetadata, UserGroups: []string{"flint*", "*sons"}},
					{Level: LevelRequest, UserGroups: []string{"muppets"}},
					{Level: LevelNone, UserGroups: []string{}},
				},
			}
			Expect(Filtered(p, Event{User: authv1.UserInfo{Groups: []string{"flintstones", "muppets"}}})).To(HaveLevel(LevelMetadata))
			Expect(Filtered(p, Event{User: authv1.UserInfo{Groups: []string{"muppets"}}})).To(HaveLevel(LevelRequest))
			Expect(Filtered(p, Event{User: authv1.UserInfo{Groups: []string{"jetsons"}}})).To(HaveLevel(LevelMetadata))
			Expect(Filtered(p, Event{User: authv1.UserInfo{Groups: []string{"lostboys"}}})).To(HaveLevel(LevelNone))
		})

		It("matches by namespace", func() {
			p := &loggingv1.KubeAPIAudit{
				Rules: []PolicyRule{
					{Level: LevelMetadata, Namespaces: []string{"flintstones", "jetsons"}},
					{Level: LevelRequest, Namespaces: []string{"muppets"}},
				},
			}
			Expect(Filtered(p, Event{ObjectRef: &ObjectReference{Namespace: "flintstones"}})).To(HaveLevel(LevelMetadata))
			Expect(Filtered(p, Event{ObjectRef: &ObjectReference{Namespace: "jetsons"}})).To(HaveLevel(LevelMetadata))
			Expect(Filtered(p, Event{ObjectRef: &ObjectReference{Namespace: "muppets"}})).To(HaveLevel(LevelRequest))
			Expect(Filtered(p, Event{ObjectRef: &ObjectReference{Namespace: "other"}})).To(HaveLevel(LevelRequest))
		})

		It("matches by resource", func() {
			p := &loggingv1.KubeAPIAudit{
				Rules: []PolicyRule{
					{
						Level: LevelMetadata,
						Resources: []GroupResources{
							{Resources: []string{"configmaps"}},
							{Group: "batch", Resources: []string{"*/status"}},
						},
					},
					{
						Level: LevelRequest,
						Resources: []GroupResources{
							{
								Group:         "route.openshift.io",
								ResourceNames: []string{"console", "downloads"},
								Resources:     []string{"routes", "routes/*"},
							},
						},
					},
				},
			}
			Expect(Filtered(p, Event{ObjectRef: &ObjectReference{APIGroup: "route.openshift.io", Name: "console", Resource: "routes"}})).To(HaveLevel(LevelRequest))
			Expect(Filtered(p, Event{ObjectRef: &ObjectReference{APIGroup: "route.openshift.io", Name: "console", Resource: "routes/foo"}})).To(HaveLevel(LevelRequest))
			Expect(Filtered(p, Event{ObjectRef: &ObjectReference{APIGroup: "batch", Resource: "foo/status"}})).To(HaveLevel(LevelMetadata))
			Expect(Filtered(p, Event{ObjectRef: &ObjectReference{Resource: "configmaps"}})).To(HaveLevel(LevelMetadata))
			Expect(Filtered(p, Event{ObjectRef: &ObjectReference{APIGroup: "", Resource: "configmaps"}})).To(HaveLevel(LevelMetadata))
			// Empty but not invalid
			Expect(Filtered(p, Event{ObjectRef: nil})).To(HaveLevel(LevelRequest))
			Expect(Filtered(p, Event{ObjectRef: &ObjectReference{}})).To(HaveLevel(LevelRequest))
		})

		It("matches by non-resource", func() {
			p := &loggingv1.KubeAPIAudit{
				Rules: []PolicyRule{
					{Level: LevelMetadata, NonResourceURLs: []string{"/metrics", "/health*"}},
					{Level: LevelRequest, NonResourceURLs: []string{"/muppets"}},
				},
			}
			Expect(Filtered(p, Event{RequestURI: "/metrics"})).To(HaveLevel(LevelMetadata))
			Expect(Filtered(p, Event{RequestURI: "/metrics?foo=bar&x=y"})).To(HaveLevel(LevelMetadata))
			Expect(Filtered(p, Event{RequestURI: "/health"})).To(HaveLevel(LevelMetadata))
			Expect(Filtered(p, Event{RequestURI: "/healthy"})).To(HaveLevel(LevelMetadata))
			Expect(Filtered(p, Event{RequestURI: "/healthiest"})).To(HaveLevel(LevelMetadata))
			Expect(Filtered(p, Event{RequestURI: "/muppets?hey=ho"})).To(HaveLevel(LevelRequest))
		})

		It("matches multiple criteria", func() {
			p := &loggingv1.KubeAPIAudit{
				Rules: []PolicyRule{
					{Level: LevelRequestResponse, Verbs: []string{"get"}, Users: []string{"nobody"}},
					{Level: LevelRequest, UserGroups: []string{"flintstones"}},
					{Level: LevelMetadata, Verbs: []string{"get"}},
					{Level: LevelNone, Users: []string{"nobody"}},
				},
			}
			Expect(Filtered(p, Event{Verb: "get", User: authv1.UserInfo{Username: "nobody"}})).To(HaveLevel(LevelRequestResponse))
			Expect(Filtered(p, Event{Verb: "get", User: authv1.UserInfo{Username: "fred"}})).To(HaveLevel(LevelMetadata))
			Expect(Filtered(p, Event{Verb: "get", User: authv1.UserInfo{Username: "fred", Groups: []string{"flintstones"}}})).To(HaveLevel(LevelRequest))
			Expect(Filtered(p, Event{User: authv1.UserInfo{Username: "nobody"}})).To(HaveLevel(LevelNone))
			Expect(Filtered(p, Event{Verb: "watch", User: authv1.UserInfo{Username: "nobody"}})).To(HaveLevel(LevelNone))
		})
	})

	Context("no rules", func() {
		It("applies default rules", func() {
			p := &loggingv1.KubeAPIAudit{}
			// User events
			Expect(Filtered(p, Event{Verb: "get", User: authv1.UserInfo{Username: "fred"}})).To(HaveLevel(LevelRequestResponse))
			Expect(Filtered(p, Event{Verb: "put", User: authv1.UserInfo{Username: "fred", Groups: []string{"flintstones"}}})).To(HaveLevel(LevelRequestResponse))
			// Read-only system events
			Expect(Filtered(p, Event{Verb: "get", User: authv1.UserInfo{Username: "system:serviceaccount:foo"}})).To(HaveLevel(LevelNone))
			Expect(Filtered(p, Event{Verb: "watch", User: authv1.UserInfo{Username: "system:serviceaccount:foo"}})).To(HaveLevel(LevelNone))
			// Write events for service account in same namespace as resource
			Expect(Filtered(p, Event{
				Verb:      "update",
				User:      authv1.UserInfo{Username: "system:serviceaccount:foo"},
				ObjectRef: &ObjectReference{Namespace: "foo"}})).To(HaveLevel(LevelNone))
			// Write service account events for non-namespace resource.
			Expect(Filtered(p, Event{Verb: "patch", User: authv1.UserInfo{Username: "system:serviceaccount:foo"}})).To(HaveLevel(LevelNone))
			// Other system events
			Expect(Filtered(p, Event{Verb: "update", User: authv1.UserInfo{Username: "system:blah"}})).To(HaveLevel(LevelRequest))
		})
	})
})
