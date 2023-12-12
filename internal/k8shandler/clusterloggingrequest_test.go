package k8shandler

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

var _ = Describe("ClusterLoggingRequest", func() {

	var (
		clr *ClusterLoggingRequest
	)
	BeforeEach(func() {
		clr = &ClusterLoggingRequest{
			Cluster: &logging.ClusterLogging{
				Spec: logging.ClusterLoggingSpec{},
			},
		}

	})

	Describe("#NewClusterLoggingRequest", func() {
		var (
			forwarder *logging.ClusterLogForwarder
			cl        *logging.ClusterLogging
		)
		BeforeEach(func() {
			cl = runtime.NewClusterLogging(constants.OpenshiftNS, constants.SingletonName)
			cl.SetUID("cl-uuid")
			forwarder = runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)
			forwarder.SetUID("clf-uuid")
		})
		Context("when reconciling legacy mode", func() {
			It("should initialize the resource owner to ClusterLogging for a virtual ClusterLogForwarder", func() {
				forwarder.SetUID("")
				cr := NewClusterLoggingRequest(cl, forwarder, nil, nil, nil, "", "", nil)
				Expect(cr.ResourceOwner).To(Equal(utils.AsOwner(cl)))
			})
			It("should initialize the resource owner to ClusterLogging for an actual ClusterLogForwarder", func() {
				cr := NewClusterLoggingRequest(cl, forwarder, nil, nil, nil, "", "", nil)
				Expect(cr.ResourceOwner).To(Equal(utils.AsOwner(cl)))
			})
		})
		Context("when reconciling multi ClusterLogForwarder mode", func() {
			BeforeEach(func() {
				cl.SetNamespace("test-namespace")
				forwarder.SetNamespace("test-namespace")
			})
			It("should initialize the resource owner to ClusterLogForwarder for a virtual ClusterLogging", func() {
				cl.SetUID("")
				cr := NewClusterLoggingRequest(cl, forwarder, nil, nil, nil, "", "", nil)
				Expect(cr.ResourceOwner).To(Equal(utils.AsOwner(forwarder)))
			})
			It("should initialize the resource owner to ClusterLogForwarder for an actual ClusterLogging", func() {
				cr := NewClusterLoggingRequest(cl, forwarder, nil, nil, nil, "", "", nil)
				Expect(cr.ResourceOwner).To(Equal(utils.AsOwner(forwarder)))
			})
		})

		Context("when reconciling ClusterLogForwarder defining an HTTP receiver", func() {
			var (
				receiverName = "http-audit"
			)
			BeforeEach(func() {
				cl.SetNamespace("test-namespace")
				forwarder.SetNamespace("test-namespace")
				forwarder.Spec = logging.ClusterLogForwarderSpec{
					Inputs: []logging.InputSpec{
						{
							Name: receiverName,
							Receiver: &logging.ReceiverSpec{
								Type: logging.OutputTypeHttp,
								ReceiverTypeSpec: &logging.ReceiverTypeSpec{
									HTTP: &logging.HTTPReceiver{
										Format: logging.FormatKubeAPIAudit,
										Port:   8080,
									},
								},
							},
						},
					},
				}
			})

			Context("when an HTTP receiver is the only inputRef", func() {
				BeforeEach(func() {
					forwarder.Spec.Pipelines = []logging.PipelineSpec{
						{
							Name:       "http-audit",
							InputRefs:  []string{receiverName},
							OutputRefs: []string{logging.OutputNameDefault},
						},
					}
				})
				Context("in one pipeline", func() {
					It("should not be a deployment if AnnotationEnableCollectorAsDeployment is not defined in CLF", func() {
						cl.SetUID("")
						cr := NewClusterLoggingRequest(cl, forwarder, nil, nil, nil, "", "", nil)
						Expect(cr.isDaemonset).To(Equal(true))
					})

					It("should be a deployment if AnnotationEnableCollectorAsDeployment is defined in the CLF's annotations", func() {
						forwarder.Annotations = map[string]string{constants.AnnotationEnableCollectorAsDeployment: "true"}
						cr := NewClusterLoggingRequest(cl, forwarder, nil, nil, nil, "", "", nil)
						Expect(cr.isDaemonset).To(Equal(false))
					})
				})
				Context("in multiple pipelines", func() {
					BeforeEach(func() {
						forwarder.Spec.Pipelines = append(forwarder.Spec.Pipelines, logging.PipelineSpec{
							Name:       "http-audit-2",
							InputRefs:  []string{receiverName},
							OutputRefs: []string{logging.OutputNameDefault},
						})
					})
					It("should not be a deployment if AnnotationEnableCollectorAsDeployment is not defined in CLF", func() {
						cl.SetUID("")
						cr := NewClusterLoggingRequest(cl, forwarder, nil, nil, nil, "", "", nil)
						Expect(cr.isDaemonset).To(Equal(true))
					})

					It("should be a deployment if AnnotationEnableCollectorAsDeployment is defined in the CLF's annotations", func() {
						forwarder.Annotations = map[string]string{constants.AnnotationEnableCollectorAsDeployment: "true"}
						cr := NewClusterLoggingRequest(cl, forwarder, nil, nil, nil, "", "", nil)
						Expect(cr.isDaemonset).To(Equal(false))
					})
				})
				Context("in one pipeline but another pipeline references a reserved input name only", func() {
					BeforeEach(func() {
						forwarder.Spec.Pipelines = append(forwarder.Spec.Pipelines, logging.PipelineSpec{
							Name:       "app-logs",
							InputRefs:  []string{logging.InputNameApplication},
							OutputRefs: []string{logging.OutputNameDefault},
						})
					})
					It("should not be a deployment if AnnotationEnableCollectorAsDeployment is not defined in CLF", func() {
						cl.SetUID("")
						cr := NewClusterLoggingRequest(cl, forwarder, nil, nil, nil, "", "", nil)
						Expect(cr.isDaemonset).To(Equal(true))
					})

					It("should not be a deployment if AnnotationEnableCollectorAsDeployment is defined in the CLF's annotations", func() {
						forwarder.Annotations = map[string]string{constants.AnnotationEnableCollectorAsDeployment: "true"}
						cr := NewClusterLoggingRequest(cl, forwarder, nil, nil, nil, "", "", nil)
						Expect(cr.isDaemonset).To(Equal(true))
					})
				})
			})

			Context("when an HTTP receiver is not the only inputRef", func() {
				BeforeEach(func() {
					forwarder.Spec.Pipelines = []logging.PipelineSpec{
						{
							Name:       "http-audit-app",
							InputRefs:  []string{receiverName, logging.InputNameApplication},
							OutputRefs: []string{logging.OutputNameDefault},
						},
					}
				})
				Context("in one pipeline", func() {
					It("should not be a deployment if AnnotationEnableCollectorAsDeployment is not defined in CLF", func() {
						cl.SetUID("")
						cr := NewClusterLoggingRequest(cl, forwarder, nil, nil, nil, "", "", nil)
						Expect(cr.isDaemonset).To(Equal(true))
					})

					It("should not be a deployment if AnnotationEnableCollectorAsDeployment is defined in the CLF's annotations", func() {
						forwarder.Annotations = map[string]string{constants.AnnotationEnableCollectorAsDeployment: "true"}
						cr := NewClusterLoggingRequest(cl, forwarder, nil, nil, nil, "", "", nil)
						Expect(cr.isDaemonset).To(Equal(true))
					})
				})
				Context("in multiple pipelines", func() {
					BeforeEach(func() {
						forwarder.Spec.Pipelines = append(forwarder.Spec.Pipelines, logging.PipelineSpec{
							Name:       "http-audit-app-2",
							InputRefs:  []string{receiverName, logging.InputNameApplication},
							OutputRefs: []string{logging.OutputNameDefault},
						})
					})
					It("should not be a deployment if AnnotationEnableCollectorAsDeployment is not defined in CLF", func() {
						cl.SetUID("")
						cr := NewClusterLoggingRequest(cl, forwarder, nil, nil, nil, "", "", nil)
						Expect(cr.isDaemonset).To(Equal(true))
					})

					It("should not be a deployment if AnnotationEnableCollectorAsDeployment is defined in the CLF's annotations", func() {
						forwarder.Annotations = map[string]string{constants.AnnotationEnableCollectorAsDeployment: "true"}
						cr := NewClusterLoggingRequest(cl, forwarder, nil, nil, nil, "", "", nil)
						Expect(cr.isDaemonset).To(Equal(true))
					})
				})
			})

		})
	})

	Describe("#IncludesManagedStorage", func() {
		Context("when logstore", func() {
			Context("is defined", func() {

				It("should return true because we are writing to the managed store", func() {
					clr.Cluster.Spec.LogStore = &logging.LogStoreSpec{}
					Expect(clr.IncludesManagedStorage()).To(BeTrue())
				})

			})
			Context("is not defined", func() {
				It("should return false because there is nowhere to write logs", func() {
					Expect(clr.IncludesManagedStorage()).To(BeFalse())
				})
			})

		})
	})
})
