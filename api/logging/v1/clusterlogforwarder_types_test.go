package v1_test

// TODO: Migrate or remove me
//import (
//	"time"
//
//	. "github.com/onsi/ginkgo"
//	. "github.com/onsi/gomega"
//	. "github.com/openshift/cluster-logging-operator/api/logging/v1"
//	"github.com/openshift/cluster-logging-operator/internal/status"
//	. "github.com/openshift/cluster-logging-operator/test"
//	. "github.com/openshift/cluster-logging-operator/test/matchers"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//)
//
//var _ = Describe("ClusterLogForwarder", func() {
//	It("serializes with conditions correctly", func() {
//		forwarder := ClusterLogForwarder{
//			Spec: ClusterLogForwarderSpec{
//				Pipelines: []PipelineSpec{
//					{
//						InputRefs:  []string{InputNameApplication},
//						OutputRefs: []string{"X", "Y"},
//					},
//					{
//						InputRefs:  []string{InputNameInfrastructure, InputNameAudit},
//						OutputRefs: []string{"Y", "Z"},
//					},
//					{
//						InputRefs:  []string{InputNameAudit},
//						OutputRefs: []string{"X", "Z"},
//						FilterRefs: []string{"F", "G"},
//					},
//				},
//			},
//			Status: ClusterLogForwarderStatus{
//				Conditions: NewConditions(Condition{
//					Type:   "Bad",
//					Reason: "NotGood",
//					Status: "True",
//				}),
//				Inputs: NamedConditions{
//					"MyInput": NewConditions(Condition{
//						Type:   "Broken",
//						Reason: "InputBroken",
//						Status: "True",
//					}),
//				},
//			},
//		}
//		// Reset the timestamps so we get a predictable output
//		t := metav1.Date(1999, time.January, 1, 0, 0, 0, 0, time.UTC)
//		forwarder.Status.Conditions[0].LastTransitionTime = t
//		forwarder.Status.Inputs["MyInput"][0].LastTransitionTime = t
//		Expect(YAMLString(forwarder)).To(EqualLines(`
// metadata:
//   creationTimestamp: null
// spec:
//   pipelines:
//   - inputRefs:
//     - application
//     outputRefs:
//     - X
//     - "Y"
//   - inputRefs:
//     - infrastructure
//     - audit
//     outputRefs:
//     - "Y"
//     - Z
//   - filterRefs:
//     - F
//     - G
//     inputRefs:
//     - audit
//     outputRefs:
//     - X
//     - Z
// status:
//   conditions:
//   - lastTransitionTime: "1999-01-01T00:00:00Z"
//     reason: NotGood
//     status: "True"
//     type: Bad
//   inputs:
//     MyInput:
//     - lastTransitionTime: "1999-01-01T00:00:00Z"
//       reason: InputBroken
//       status: "True"
//       type: Broken
//
//`))
//		Expect(JSONString(forwarder)).To(EqualLines(`{
//   "metadata": {
//     "creationTimestamp": null
//   },
//   "spec": {
//     "pipelines": [
//       {
//         "outputRefs": [
//           "X",
//           "Y"
//         ],
//         "inputRefs": [
//           "application"
//         ]
//       },
//       {
//         "outputRefs": [
//           "Y",
//           "Z"
//         ],
//         "inputRefs": [
//           "infrastructure",
//           "audit"
//         ]
//       },
//       {
//         "outputRefs": [
//           "X",
//           "Z"
//         ],
//         "inputRefs": [
//           "audit"
//         ],
//         "filterRefs": [
//           "F",
//           "G"
//         ]
//       }
//     ]
//   },
//   "status": {
//     "conditions": [
//       {
//         "type": "Bad",
//         "status": "True",
//         "reason": "NotGood",
//         "lastTransitionTime": "1999-01-01T00:00:00Z"
//       }
//     ],
//     "inputs": {
//       "MyInput": [
//         {
//           "type": "Broken",
//           "status": "True",
//           "reason": "InputBroken",
//           "lastTransitionTime": "1999-01-01T00:00:00Z"
//         }
//       ]
//     }
//   }
// }`))
//	})
//})
//
//var _ = Describe("ClusterLogForwarderStatus", func() {
//	It("synchronizes ClusterLogForwarderStatus while handling timestamps correctly", func() {
//		original := ClusterLogForwarderStatus{
//			Conditions: status.Conditions{
//				{
//					Type:               "Ready",
//					Status:             "False",
//					Reason:             "name1 is not ready",
//					Message:            "name1 is not ready due to issue",
//					LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//				},
//				{
//					Type:               "Error",
//					Status:             "True",
//					Reason:             "name1 has an error",
//					Message:            "name1 has an error due to an issue",
//					LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//				},
//				{
//					Type:               "Foo",
//					Status:             "Foo",
//					Reason:             "Foo reason",
//					Message:            "Foo message",
//					LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//				},
//			},
//			Inputs: NamedConditions{
//				"name1": {
//					{
//						Type:               "Ready",
//						Status:             "False",
//						Reason:             "name1 is not ready",
//						Message:            "name1 is not ready due to issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//					{
//						Type:               "Available",
//						Status:             "False",
//						Reason:             "name1 is not available",
//						Message:            "name1 is not available due to issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//					{
//						Type:               "Error",
//						Status:             "True",
//						Reason:             "name1 has an error",
//						Message:            "name1 has an error due to an issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//				},
//				"name2a": {
//					{
//						Type:               "Ready",
//						Status:             "False",
//						Reason:             "name2 is not ready",
//						Message:            "name2 is not ready due to issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//					{
//						Type:               "Available",
//						Status:             "False",
//						Reason:             "name2 is not available",
//						Message:            "name2 is not available due to issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//					{
//						Type:               "Error",
//						Status:             "True",
//						Reason:             "name2 has an error",
//						Message:            "name2 has an error due to an issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//				},
//				"name2b": {
//					{
//						Type:               "Ready",
//						Status:             "False",
//						Reason:             "name2 is not ready",
//						Message:            "name2 is not ready due to issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//					{
//						Type:               "Available",
//						Status:             "False",
//						Reason:             "name2 is not available",
//						Message:            "name2 is not available due to issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//					{
//						Type:               "Error",
//						Status:             "True",
//						Reason:             "name2 has an error",
//						Message:            "name2 has an error due to an issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//				},
//				"name3": {
//					{
//						Type:               "Ready",
//						Status:             "True",
//						Reason:             "name3 is ready",
//						Message:            "name3 is ready due to issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//					{
//						Type:               "Available",
//						Status:             "True",
//						Reason:             "name3 is available",
//						Message:            "name3 is available due to issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//					{
//						Type:               "Foo",
//						Status:             "True",
//						Reason:             "Bar",
//						Message:            "Bar's message",
//						LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//					{
//						Type:               "Foo2",
//						Status:             "True",
//						Reason:             "Bar2",
//						Message:            "Bar2's message",
//						LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//				},
//			},
//		}
//
//		syncTarget := ClusterLogForwarderStatus{
//			Conditions: status.Conditions{
//				// We expect Ready to be updated to True and have the current timestamp.
//				{
//					Type:               "Ready",
//					Status:             "True",
//					Reason:             "Operator is ready",
//					Message:            "Operator is ready",
//					LastTransitionTime: metav1.NewTime(time.Date(2023, 1, 1, 12, 30, 30, 100, time.UTC)),
//				},
//				// We expect Error to be removed.
//				// We expect Available to show up with the current timestamp.
//				{
//					Type:               "Available",
//					Status:             "True",
//					Reason:             "Operator is available",
//					Message:            "Operator is available",
//					LastTransitionTime: metav1.NewTime(time.Date(2023, 1, 1, 12, 30, 30, 100, time.UTC)),
//				},
//				// We expect Foo's LastTransitionTime to remain unchanged.
//				{
//					Type:               "Foo",
//					Status:             "Foo",
//					Reason:             "Foo reason",
//					Message:            "Foo message",
//					LastTransitionTime: metav1.NewTime(time.Date(2023, 1, 1, 12, 30, 30, 100, time.UTC)),
//				},
//			},
//			Inputs: NamedConditions{
//				// Update all values within the names entry, expect all fields to match and time stamps to be recent.
//				"name1": {
//					{
//						Type:               "Ready",
//						Status:             "True",
//						Reason:             "name1 is ready",
//						Message:            "name1 is ready with no issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2024, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//					{
//						Type:               "Available",
//						Status:             "True",
//						Reason:             "name1 is available",
//						Message:            "name1 is available with no issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2024, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//					{
//						Type:               "Error",
//						Status:             "False",
//						Reason:             "name1 has no error",
//						Message:            "name1 has no error due to an issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2024, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//				},
//				// Drop name 'name2a' and 'name2b'.
//				// Drop 'Ready' and 'Available' types. Keep status of 'Foo' and expect the timestamp to remain stable.
//				// Update Foo2 and expect the timestamp to be new.
//				"name3": {
//					{
//						Type:               "Foo",
//						Status:             "True",
//						Reason:             "Bar",
//						Message:            "Bar's message",
//						LastTransitionTime: metav1.NewTime(time.Date(2023, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//					{
//						Type:               "Error",
//						Status:             "True",
//						Reason:             "name3 has an error",
//						Message:            "name3 has an error due to an issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//					{
//						Type:               "Foo2",
//						Status:             "False",
//						Reason:             "Bar2",
//						Message:            "Bar2's message",
//						LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//				},
//				// Add index name4 and expect all timestamps to be recent.
//				"name4": {
//					{
//						Type:               "Ready",
//						Status:             "True",
//						Reason:             "name4 is ready",
//						Message:            "name4 is ready with no issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2024, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//					{
//						Type:               "Available",
//						Status:             "True",
//						Reason:             "name4 is available",
//						Message:            "name4 is available with no issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2024, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//				},
//			},
//		}
//
//		expected := ClusterLogForwarderStatus{
//			Conditions: status.Conditions{
//				{
//					Type:               "Ready",
//					Status:             "True",
//					Reason:             "Operator is ready",
//					Message:            "Operator is ready",
//					LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//				},
//				{
//					Type:               "Foo",
//					Status:             "Foo",
//					Reason:             "Foo reason",
//					Message:            "Foo message",
//					LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//				},
//				{
//					Type:               "Available",
//					Status:             "True",
//					Reason:             "Operator is available",
//					Message:            "Operator is available",
//					LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//				},
//			},
//			Inputs: NamedConditions{
//				"name1": {
//					{
//						Type:               "Ready",
//						Status:             "True",
//						Reason:             "name1 is ready",
//						Message:            "name1 is ready with no issue",
//						LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//					},
//					{
//						Type:               "Available",
//						Status:             "True",
//						Reason:             "name1 is available",
//						Message:            "name1 is available with no issue",
//						LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//					},
//					{
//						Type:               "Error",
//						Status:             "False",
//						Reason:             "name1 has no error",
//						Message:            "name1 has no error due to an issue",
//						LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//					},
//				},
//				"name3": {
//					{
//						Type:               "Foo",
//						Status:             "True",
//						Reason:             "Bar",
//						Message:            "Bar's message",
//						LastTransitionTime: metav1.NewTime(time.Date(2021, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//					{
//						Type:               "Foo2",
//						Status:             "False",
//						Reason:             "Bar2",
//						Message:            "Bar2's message",
//						LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//					},
//					{
//						Type:               "Error",
//						Status:             "True",
//						Reason:             "name3 has an error",
//						Message:            "name3 has an error due to an issue",
//						LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//					},
//				},
//				"name4": {
//					{
//						Type:               "Ready",
//						Status:             "True",
//						Reason:             "name4 is ready",
//						Message:            "name4 is ready with no issue",
//						LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//					},
//					{
//						Type:               "Available",
//						Status:             "True",
//						Reason:             "name4 is available",
//						Message:            "name4 is available with no issue",
//						LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//					},
//				},
//			},
//		}
//
//		err := original.Synchronize(&syncTarget)
//		Expect(err).NotTo(HaveOccurred())
//		Expect(original.Conditions).To(MatchConditions(expected.Conditions))
//		namedConditionEquals(original.Inputs, expected.Inputs)
//	})
//
//	It("synchronizes ClusterLogForwarderStatus with empty conditions", func() {
//		var original0 *ClusterLogForwarderStatus
//		original1 := ClusterLogForwarderStatus{
//			Conditions: nil,
//			Inputs: NamedConditions{
//				"name1": nil,
//			},
//		}
//		original2 := ClusterLogForwarderStatus{
//			Conditions: nil,
//			Filters:    nil,
//			Inputs:     nil,
//			Outputs:    nil,
//			Pipelines:  nil,
//		}
//
//		syncTarget := ClusterLogForwarderStatus{
//			Conditions: status.Conditions{
//				{
//					Type:               "Ready",
//					Status:             "True",
//					Reason:             "Operator is ready",
//					Message:            "Operator is ready",
//					LastTransitionTime: metav1.NewTime(time.Date(2023, 1, 1, 12, 30, 30, 100, time.UTC)),
//				},
//				{
//					Type:               "Available",
//					Status:             "True",
//					Reason:             "Operator is available",
//					Message:            "Operator is available",
//					LastTransitionTime: metav1.NewTime(time.Date(2023, 1, 1, 12, 30, 30, 100, time.UTC)),
//				},
//			},
//			Filters: NamedConditions{
//				"name1": {
//					{
//						Type:               "Ready",
//						Status:             "True",
//						Reason:             "name1 is ready",
//						Message:            "name1 is ready with no issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2024, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//				},
//				"name2": {
//					{
//						Type:               "Foo",
//						Status:             "True",
//						Reason:             "Bar",
//						Message:            "Bar's message",
//						LastTransitionTime: metav1.NewTime(time.Date(2023, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//				},
//			},
//			Inputs: NamedConditions{
//				"name1": {
//					{
//						Type:               "Ready",
//						Status:             "True",
//						Reason:             "name1 is ready",
//						Message:            "name1 is ready with no issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2024, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//				},
//				"name2": {
//					{
//						Type:               "Foo",
//						Status:             "True",
//						Reason:             "Bar",
//						Message:            "Bar's message",
//						LastTransitionTime: metav1.NewTime(time.Date(2023, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//				},
//			},
//			Outputs: NamedConditions{
//				"name1": {
//					{
//						Type:               "Ready",
//						Status:             "True",
//						Reason:             "name1 is ready",
//						Message:            "name1 is ready with no issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2024, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//				},
//				"name2": {
//					{
//						Type:               "Foo",
//						Status:             "True",
//						Reason:             "Bar",
//						Message:            "Bar's message",
//						LastTransitionTime: metav1.NewTime(time.Date(2023, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//				},
//			},
//			Pipelines: NamedConditions{
//				"name1": {
//					{
//						Type:               "Ready",
//						Status:             "True",
//						Reason:             "name1 is ready",
//						Message:            "name1 is ready with no issue",
//						LastTransitionTime: metav1.NewTime(time.Date(2024, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//				},
//				"name2": {
//					{
//						Type:               "Foo",
//						Status:             "True",
//						Reason:             "Bar",
//						Message:            "Bar's message",
//						LastTransitionTime: metav1.NewTime(time.Date(2023, 1, 1, 12, 30, 30, 100, time.UTC)),
//					},
//				},
//			},
//		}
//
//		expected := ClusterLogForwarderStatus{
//			Conditions: status.Conditions{
//				{
//					Type:               "Ready",
//					Status:             "True",
//					Reason:             "Operator is ready",
//					Message:            "Operator is ready",
//					LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//				},
//				{
//					Type:               "Available",
//					Status:             "True",
//					Reason:             "Operator is available",
//					Message:            "Operator is available",
//					LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//				},
//			},
//			Filters: NamedConditions{
//				"name1": {
//					{
//						Type:               "Ready",
//						Status:             "True",
//						Reason:             "name1 is ready",
//						Message:            "name1 is ready with no issue",
//						LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//					},
//				},
//				"name2": {
//					{
//						Type:               "Foo",
//						Status:             "True",
//						Reason:             "Bar",
//						Message:            "Bar's message",
//						LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//					},
//				},
//			},
//			Inputs: NamedConditions{
//				"name1": {
//					{
//						Type:               "Ready",
//						Status:             "True",
//						Reason:             "name1 is ready",
//						Message:            "name1 is ready with no issue",
//						LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//					},
//				},
//				"name2": {
//					{
//						Type:               "Foo",
//						Status:             "True",
//						Reason:             "Bar",
//						Message:            "Bar's message",
//						LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//					},
//				},
//			},
//			Outputs: NamedConditions{
//				"name1": {
//					{
//						Type:               "Ready",
//						Status:             "True",
//						Reason:             "name1 is ready",
//						Message:            "name1 is ready with no issue",
//						LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//					},
//				},
//				"name2": {
//					{
//						Type:               "Foo",
//						Status:             "True",
//						Reason:             "Bar",
//						Message:            "Bar's message",
//						LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//					},
//				},
//			},
//			Pipelines: NamedConditions{
//				"name1": {
//					{
//						Type:               "Ready",
//						Status:             "True",
//						Reason:             "name1 is ready",
//						Message:            "name1 is ready with no issue",
//						LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//					},
//				},
//				"name2": {
//					{
//						Type:               "Foo",
//						Status:             "True",
//						Reason:             "Bar",
//						Message:            "Bar's message",
//						LastTransitionTime: metav1.NewTime(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)),
//					},
//				},
//			},
//		}
//
//		err := original0.Synchronize(&syncTarget)
//		Expect(err).To(HaveOccurred())
//
//		err = original1.Synchronize(&syncTarget)
//		Expect(err).NotTo(HaveOccurred())
//		Expect(original1.Conditions).To(MatchConditions(expected.Conditions))
//		namedConditionEquals(original1.Inputs, expected.Inputs)
//
//		err = original2.Synchronize(&syncTarget)
//		Expect(err).NotTo(HaveOccurred())
//		Expect(original2.Conditions).To(MatchConditions(expected.Conditions))
//		namedConditionEquals(original2.Inputs, expected.Inputs)
//	})
//})
//
//// namedConditionEquals verifies if 2 NamedConditions can be considered the same.
//func namedConditionEquals(got, expected NamedConditions) {
//	Expect(len(got)).To(Equal(len(expected)))
//	for name := range got {
//		Expect(expected).To(HaveKey(name))
//		Expect(got[name]).To(MatchConditions(expected[name]))
//	}
//}
