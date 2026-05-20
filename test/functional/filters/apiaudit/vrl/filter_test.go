package apiaudit

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test"
	authv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	. "k8s.io/apiserver/pkg/apis/audit/v1"
)

var _ = Describe("Policy to VRL Filter", func() {

	Context("omit response codes", func() {
		It("should omit specified codes", func() {
			p := &obs.KubeAPIAudit{OmitResponseCodes: &[]int{1234}}
			Expect(Filtered(p, Event{ResponseStatus: &v1.Status{Code: 1234}})).To(BeNil())
			Expect(Filtered(p, Event{ResponseStatus: &v1.Status{Code: 5678}})).NotTo(BeNil())
		})
		It("should omit default codes if missing", func() {
			p := &obs.KubeAPIAudit{}
			Expect(Filtered(p, Event{ResponseStatus: &v1.Status{Code: 409}})).To(BeNil())
			Expect(Filtered(p, Event{ResponseStatus: &v1.Status{Code: 200}})).NotTo(BeNil())
		})
		It("should not omit by code if explicitly empty", func() {
			p := &obs.KubeAPIAudit{OmitResponseCodes: &[]int{}} // Explicitly Empty
			Expect(Filtered(p, Event{ResponseStatus: &v1.Status{Code: 409}})).NotTo(BeNil())
			Expect(Filtered(p, Event{ResponseStatus: &v1.Status{Code: 200}})).NotTo(BeNil())
		})
	})

	Context("omit stages", func() {
		It("for LOG-5607 should drop events when rules with omitStages have multiple stages and the ref object has a subresource", func() {
			p := &obs.KubeAPIAudit{
				OmitStages: []Stage{StageRequestReceived},
				Rules: []PolicyRule{
					{
						Level: LevelRequest,
						Resources: []GroupResources{
							{Resources: []string{"configmaps", "secrets"}},
						},
						Verbs: []string{"create", "delete", "get", "update"},
					},
					{
						Level:      LevelNone,
						OmitStages: []Stage{StageResponseComplete},
						Resources: []GroupResources{
							{Group: "policy.open-cluster-management.io", Resources: []string{"configurationpolicies", "configurationpolicies/*"}},
						},
						Verbs: []string{"update"},
					},
				},
			}
			content, err := jsonContent.ReadFile("log5607_sample.json")
			Expect(err).To(BeNil())
			e := &Event{}
			test.MustUnmarshal(string(content), e)
			Expect(Filtered(p, *e)).To(BeNil())
		})
		It("omits using policy setting", func() {
			p := &obs.KubeAPIAudit{OmitStages: []Stage{StageRequestReceived}}
			Expect(Filtered(p, Event{Stage: StageRequestReceived})).To(BeNil())
			Expect(Filtered(p, Event{Stage: StageResponseStarted})).NotTo(BeNil())
		})
		It("omits using rule setting", func() {
			p := &obs.KubeAPIAudit{Rules: []PolicyRule{
				{Verbs: []string{"update"}, OmitStages: []Stage{StagePanic}},
			}}
			Expect(Filtered(p, Event{Verb: "update", Stage: StagePanic})).To(BeNil())
			Expect(Filtered(p, Event{Verb: "create", Stage: StagePanic})).NotTo(BeNil())
		})
		It("omits using rule and policy setting", func() {
			p := &obs.KubeAPIAudit{
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
			p := &obs.KubeAPIAudit{
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
			p := &obs.KubeAPIAudit{
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
			p := &obs.KubeAPIAudit{
				Rules: []PolicyRule{
					{Level: LevelRequest, Users: []string{"barney"}},
					{Level: LevelNone, Verbs: []string{"get"}},
				},
			}
			Expect(Filtered(p, Event{Verb: "get", User: authv1.UserInfo{Username: "barney"}})).To(HaveLevel(LevelRequest))
			Expect(Filtered(p, Event{Verb: "get"})).To(HaveLevel(LevelNone))
		})

		It("matches by group", func() {
			p := &obs.KubeAPIAudit{
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
			p := &obs.KubeAPIAudit{
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
			p := &obs.KubeAPIAudit{
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
			p := &obs.KubeAPIAudit{
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
			p := &obs.KubeAPIAudit{
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
			p := &obs.KubeAPIAudit{}
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

	Context("omit managed fields", func() {
		createEventWithManagedFields := func() Event {
			reqObj := &v1.PartialObjectMetadata{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-object",
					Namespace: "test-namespace",
					ManagedFields: []v1.ManagedFieldsEntry{
						{
							Manager:   "test-manager",
							Operation: v1.ManagedFieldsOperationUpdate,
						},
					},
				},
			}
			respObj := &v1.PartialObjectMetadata{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-object",
					Namespace: "test-namespace",
					ManagedFields: []v1.ManagedFieldsEntry{
						{
							Manager:   "test-manager",
							Operation: v1.ManagedFieldsOperationUpdate,
						},
					},
				},
			}
			reqJSON, _ := json.Marshal(reqObj)
			respJSON, _ := json.Marshal(respObj)

			return Event{
				Level:          LevelRequestResponse,
				AuditID:        "test-audit-id",
				Verb:           "update",
				User:           authv1.UserInfo{Username: "test-user"},
				RequestObject:  &apiruntime.Unknown{Raw: reqJSON},
				ResponseObject: &apiruntime.Unknown{Raw: respJSON},
			}
		}

		It("should remove managedFields when OmitManagedFields is true", func() {
			omitTrue := true
			p := &obs.KubeAPIAudit{
				Rules: []PolicyRule{
					{
						Level:             LevelRequestResponse,
						OmitManagedFields: &omitTrue,
					},
				},
			}
			event := createEventWithManagedFields()
			filtered := Filtered(p, event)

			Expect(filtered).NotTo(BeNil())
			Expect(filtered.Level).To(Equal(LevelRequestResponse))

			// Verify managedFields were removed from requestObject
			reqMeta := &v1.PartialObjectMetadata{}
			err := json.Unmarshal(filtered.RequestObject.Raw, reqMeta)
			Expect(err).To(BeNil(), "Failed to unmarshal RequestObject")
			Expect(reqMeta.ObjectMeta.ManagedFields).To(BeNil())

			// Verify managedFields were removed from responseObject
			respMeta := &v1.PartialObjectMetadata{}
			err = json.Unmarshal(filtered.ResponseObject.Raw, respMeta)
			Expect(err).To(BeNil(), "Failed to unmarshal ResponseObject")
			Expect(respMeta.ObjectMeta.ManagedFields).To(BeNil())
		})

		It("should keep managedFields when OmitManagedFields is false", func() {
			omitFalse := false
			p := &obs.KubeAPIAudit{
				Rules: []PolicyRule{
					{
						Level:             LevelRequestResponse,
						OmitManagedFields: &omitFalse,
					},
				},
			}
			event := createEventWithManagedFields()
			filtered := Filtered(p, event)

			Expect(filtered).NotTo(BeNil())
			Expect(filtered.Level).To(Equal(LevelRequestResponse))

			// Verify managedFields were kept in requestObject
			reqMeta := &v1.PartialObjectMetadata{}
			err := json.Unmarshal(filtered.RequestObject.Raw, reqMeta)
			Expect(err).To(BeNil(), "Failed to unmarshal RequestObject")
			Expect(reqMeta.ObjectMeta.ManagedFields).To(HaveLen(1))

			// Verify managedFields were kept in responseObject
			respMeta := &v1.PartialObjectMetadata{}
			err = json.Unmarshal(filtered.ResponseObject.Raw, respMeta)
			Expect(err).To(BeNil(), "Failed to unmarshal ResponseObject")
			Expect(respMeta.ObjectMeta.ManagedFields).To(HaveLen(1))
		})

		It("should keep managedFields when OmitManagedFields is not set", func() {
			p := &obs.KubeAPIAudit{
				Rules: []PolicyRule{
					{
						Level: LevelRequestResponse,
						// OmitManagedFields not set (nil)
					},
				},
			}
			event := createEventWithManagedFields()
			filtered := Filtered(p, event)

			Expect(filtered).NotTo(BeNil())
			Expect(filtered.Level).To(Equal(LevelRequestResponse))

			// Verify managedFields were kept in requestObject
			reqMeta := &v1.PartialObjectMetadata{}
			err := json.Unmarshal(filtered.RequestObject.Raw, reqMeta)
			Expect(err).To(BeNil(), "Failed to unmarshal RequestObject")
			Expect(reqMeta.ObjectMeta.ManagedFields).To(HaveLen(1))

			// Verify managedFields were kept in responseObject
			respMeta := &v1.PartialObjectMetadata{}
			err = json.Unmarshal(filtered.ResponseObject.Raw, respMeta)
			Expect(err).To(BeNil(), "Failed to unmarshal ResponseObject")
			Expect(respMeta.ObjectMeta.ManagedFields).To(HaveLen(1))
		})

		Context("with event containing .metadata.managedFields", func() {
			createEventWithRootManagedFields := func() []byte {
				event := createEventWithManagedFields()
				eventJSON, _ := json.Marshal(event)

				// Parse as map to add root-level metadata.managedFields
				var eventMap map[string]interface{}
				_ = json.Unmarshal(eventJSON, &eventMap)
				eventMap["metadata"] = map[string]interface{}{
					"managedFields": []map[string]interface{}{
						{
							"manager":   "root-manager",
							"operation": "Update",
						},
					},
				}
				modifiedJSON, _ := json.Marshal(eventMap)
				return modifiedJSON
			}

			It("should remove .metadata.managedFields when OmitManagedFields is true", func() {
				omitTrue := true
				p := &obs.KubeAPIAudit{
					Rules: []PolicyRule{
						{
							Level:             LevelRequestResponse,
							OmitManagedFields: &omitTrue,
						},
					},
				}
				eventJSON := createEventWithRootManagedFields()
				filteredJSON := FilteredBytes(p, eventJSON)

				// Parse filtered output to verify .metadata.managedFields was removed
				var filteredMap map[string]interface{}
				err := json.Unmarshal(filteredJSON, &filteredMap)
				Expect(err).To(BeNil(), "Failed to unmarshal filtered output")

				// Verify root-level metadata.managedFields was removed
				if metadata, exists := filteredMap["metadata"]; exists {
					metadataMap := metadata.(map[string]interface{})
					_, hasManaged := metadataMap["managedFields"]
					Expect(hasManaged).To(BeFalse(), ".metadata.managedFields should be removed")
				}
			})

			It("should keep .metadata.managedFields when OmitManagedFields is false", func() {
				omitFalse := false
				p := &obs.KubeAPIAudit{
					Rules: []PolicyRule{
						{
							Level:             LevelRequestResponse,
							OmitManagedFields: &omitFalse,
						},
					},
				}
				eventJSON := createEventWithRootManagedFields()
				filteredJSON := FilteredBytes(p, eventJSON)

				// Parse filtered output to verify .metadata.managedFields was kept
				var filteredMap map[string]interface{}
				err := json.Unmarshal(filteredJSON, &filteredMap)
				Expect(err).To(BeNil(), "Failed to unmarshal filtered output")

				// Verify root-level metadata.managedFields was kept
				metadata, exists := filteredMap["metadata"]
				Expect(exists).To(BeTrue(), ".metadata should exist")
				metadataMap := metadata.(map[string]interface{})
				managedFields, hasManaged := metadataMap["managedFields"]
				Expect(hasManaged).To(BeTrue(), ".metadata.managedFields should be kept")
				Expect(managedFields).To(HaveLen(1))
			})

			It("should remove all managedFields from all locations when OmitManagedFields is true", func() {
				omitTrue := true
				p := &obs.KubeAPIAudit{
					Rules: []PolicyRule{
						{
							Level:             LevelRequestResponse,
							OmitManagedFields: &omitTrue,
						},
					},
				}
				eventJSON := createEventWithRootManagedFields()
				filteredJSON := FilteredBytes(p, eventJSON)

				var filteredMap map[string]interface{}
				err := json.Unmarshal(filteredJSON, &filteredMap)
				Expect(err).To(BeNil(), "Failed to unmarshal filtered output")

				// Verify root-level metadata.managedFields was removed
				if metadata, exists := filteredMap["metadata"]; exists {
					metadataMap := metadata.(map[string]interface{})
					_, hasManaged := metadataMap["managedFields"]
					Expect(hasManaged).To(BeFalse(), ".metadata.managedFields should be removed")
				}

				// Verify requestObject.metadata.managedFields was removed
				if reqObj, exists := filteredMap["requestObject"]; exists {
					reqMap := reqObj.(map[string]interface{})
					if reqMeta, exists := reqMap["metadata"]; exists {
						reqMetaMap := reqMeta.(map[string]interface{})
						_, hasManaged := reqMetaMap["managedFields"]
						Expect(hasManaged).To(BeFalse(), ".requestObject.metadata.managedFields should be removed")
					}
				}

				// Verify responseObject.metadata.managedFields was removed
				if respObj, exists := filteredMap["responseObject"]; exists {
					respMap := respObj.(map[string]interface{})
					if respMeta, exists := respMap["metadata"]; exists {
						respMetaMap := respMeta.(map[string]interface{})
						_, hasManaged := respMetaMap["managedFields"]
						Expect(hasManaged).To(BeFalse(), ".responseObject.metadata.managedFields should be removed")
					}
				}
			})
		})
	})
})
