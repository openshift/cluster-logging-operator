package network

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Reconcile NetworkPolicy", func() {

	defer GinkgoRecover()

	var (
		owner          metav1.OwnerReference
		policyInstance *networkingv1.NetworkPolicy
		policyKey      types.NamespacedName
		policyName     string
		componentName  string
		commonLabels   func(o runtime.Object)

		// Adding ns and label to account for addSecurityLabelsToNamespace() added in LOG-2620
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"test": "true"},
				Name:   constants.OpenshiftNS,
			},
		}
		reqClient = fake.NewFakeClient(
			namespace,
		)
		instanceName = "test-instance"
		protocolTCP  = corev1.ProtocolTCP
	)

	Context("when the collector NetworkPolicy is reconciled", func() {
		BeforeEach(func() {
			policyName = "collector-test-network-policy"
			componentName = constants.CollectorName
			commonLabels = func(o runtime.Object) {
				runtime.SetCommonLabels(o, constants.VectorName, instanceName, componentName)
			}
			owner = metav1.OwnerReference{
				APIVersion: "v1",
				Kind:       "ClusterLogForwarder",
				Name:       instanceName,
			}
			policyInstance = &networkingv1.NetworkPolicy{}
			policyKey = types.NamespacedName{Name: policyName, Namespace: namespace.Name}
		})

		It("should successfully reconcile the network policy", func() {
			// Reconcile the network policy
			Expect(ReconcileClusterLogForwarderNetworkPolicy(
				reqClient,
				constants.OpenshiftNS,
				policyName,
				instanceName,
				componentName,
				obsv1.NetworkPolicyRuleSetTypeAllowAllIngressEgress,
				nil,
				nil,
				owner,
				commonLabels)).To(Succeed())

			// Get and check the network policy
			Expect(reqClient.Get(context.TODO(), policyKey, policyInstance)).Should(Succeed())

			// Verify basic properties
			Expect(policyInstance.Name).To(Equal(policyName))
			Expect(policyInstance.Namespace).To(Equal(constants.OpenshiftNS))

			// Verify owner reference is set
			Expect(policyInstance.OwnerReferences).To(HaveLen(1))
			Expect(policyInstance.OwnerReferences[0].Name).To(Equal(instanceName))

			// Verify pod selector matches collector pods
			expectedPodSelector := runtime.Selectors(instanceName, componentName, constants.VectorName)
			Expect(policyInstance.Spec.PodSelector.MatchLabels).To(Equal(expectedPodSelector))

			// Verify policy types include both Ingress and Egress
			expectedPolicyTypes := []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			}
			Expect(policyInstance.Spec.PolicyTypes).To(Equal(expectedPolicyTypes))

			expectedIngressRules := []networkingv1.NetworkPolicyIngressRule{
				{},
			}
			// Verify ingress rules allow all traffic (empty rule)
			Expect(policyInstance.Spec.Ingress).To(Equal(expectedIngressRules))

			// Verify egress rules allow all traffic (empty rule)
			expectedEgressRules := []networkingv1.NetworkPolicyEgressRule{
				{},
			}
			Expect(policyInstance.Spec.Egress).To(Equal(expectedEgressRules))

			// Verify common labels are applied
			Expect(policyInstance.Labels).To(HaveKey(constants.LabelK8sName))
			Expect(policyInstance.Labels).To(HaveKey(constants.LabelK8sInstance))
			Expect(policyInstance.Labels).To(HaveKey(constants.LabelK8sComponent))
			Expect(policyInstance.Labels[constants.LabelK8sName]).To(Equal(constants.VectorName))
			Expect(policyInstance.Labels[constants.LabelK8sInstance]).To(Equal(instanceName))
			Expect(policyInstance.Labels[constants.LabelK8sComponent]).To(Equal(componentName))
		})

		It("should successfully reconcile the network policy with input receiver ports", func() {
			inputs := []obsv1.InputSpec{
				{
					Name: "http-receiver",
					Type: obsv1.InputTypeReceiver,
					Receiver: &obsv1.ReceiverSpec{
						Type: obsv1.ReceiverTypeHTTP,
						Port: 8080,
					},
				},
				{
					Name: "syslog-receiver",
					Type: obsv1.InputTypeReceiver,
					Receiver: &obsv1.ReceiverSpec{
						Type: obsv1.ReceiverTypeSyslog,
						Port: 5140,
					},
				},
			}

			// Reconcile the network policy with RestrictIngressEgress rule set
			Expect(ReconcileClusterLogForwarderNetworkPolicy(
				reqClient,
				constants.OpenshiftNS,
				policyName,
				instanceName,
				componentName,
				obsv1.NetworkPolicyRuleSetTypeRestrictIngressEgress,
				[]obsv1.OutputSpec{},
				inputs,
				owner,
				commonLabels)).To(Succeed())

			// Get and check the network policy
			Expect(reqClient.Get(context.TODO(), policyKey, policyInstance)).Should(Succeed())

			// Verify ingress rules include metrics port and receiver ports
			Expect(policyInstance.Spec.Ingress).To(HaveLen(1))
			ingressRule := policyInstance.Spec.Ingress[0]

			// Should have 3 ports: metrics port + 2 receiver ports
			Expect(ingressRule.Ports).To(HaveLen(3))

			portNumbers := make([]int32, 0)
			for _, port := range ingressRule.Ports {
				if port.Port != nil {
					portNumbers = append(portNumbers, port.Port.IntVal)
				}
			}

			Expect(portNumbers).To(ContainElements(int32(8080), int32(5140)))
		})

		It("should successfully reconcile the network policy with output ports", func() {
			outputs := []obsv1.OutputSpec{
				{
					Name: "elasticsearch-output",
					Type: obsv1.OutputTypeElasticsearch,
					Elasticsearch: &obsv1.Elasticsearch{
						URLSpec: obsv1.URLSpec{
							URL: "https://elasticsearch.example.com:9200",
						},
					},
				},
				{
					Name: "splunk-output",
					Type: obsv1.OutputTypeSplunk,
					Splunk: &obsv1.Splunk{
						URLSpec: obsv1.URLSpec{
							URL: "https://splunk.example.com:8088",
						},
					},
				},
				{
					Name: "loki-output",
					Type: obsv1.OutputTypeLoki,
					Loki: &obsv1.Loki{
						URLSpec: obsv1.URLSpec{
							URL: "http://loki.example.com:3100",
						},
					},
				},
			}

			// Reconcile the network policy with RestrictIngressEgress rule set
			Expect(ReconcileClusterLogForwarderNetworkPolicy(
				reqClient,
				constants.OpenshiftNS,
				policyName,
				instanceName,
				componentName,
				obsv1.NetworkPolicyRuleSetTypeRestrictIngressEgress,
				outputs,
				[]obsv1.InputSpec{},
				owner,
				commonLabels)).To(Succeed())

			// Get and check the network policy
			Expect(reqClient.Get(context.TODO(), policyKey, policyInstance)).Should(Succeed())

			// Verify egress rules include output ports
			Expect(policyInstance.Spec.Egress).To(HaveLen(1))
			egressRule := policyInstance.Spec.Egress[0]

			Expect(egressRule.Ports).To(HaveLen(5))

			expectedEgressPorts := []networkingv1.NetworkPolicyPort{
				{
					Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
					Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: 9200},
				},
				{
					Protocol: &[]corev1.Protocol{corev1.ProtocolUDP}[0],
					Port:     &intstr.IntOrString{Type: intstr.String, StrVal: "dns"},
				},
				{
					Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
					Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: 8088},
				},
				{
					Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
					Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: 3100},
				},
				{
					Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
					Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: 6443},
				},
			}

			Expect(egressRule.Ports).To(ConsistOf(expectedEgressPorts))
		})

		It("should successfully reconcile the network policy with both input and output ports", func() {
			// Setup inputs with receiver ports
			inputs := []obsv1.InputSpec{
				{
					Name: "http-receiver",
					Type: obsv1.InputTypeReceiver,
					Receiver: &obsv1.ReceiverSpec{
						Type: obsv1.ReceiverTypeHTTP,
						Port: 8080,
					},
				},
				{
					Name: "syslog-receiver",
					Type: obsv1.InputTypeReceiver,
					Receiver: &obsv1.ReceiverSpec{
						Type: obsv1.ReceiverTypeSyslog,
						Port: 5140,
					},
				},
			}

			// Setup outputs with different types
			outputs := []obsv1.OutputSpec{
				{
					Name: "elasticsearch-output",
					Type: obsv1.OutputTypeElasticsearch,
					Elasticsearch: &obsv1.Elasticsearch{
						URLSpec: obsv1.URLSpec{
							URL: "https://elasticsearch.example.com:9200",
						},
					},
				},
				{
					Name: "kafka-output",
					Type: obsv1.OutputTypeKafka,
					Kafka: &obsv1.Kafka{
						URL: "tcp://kafka.example.com:9092",
					},
				},
			}

			// Reconcile the network policy with RestrictIngressEgress rule set
			Expect(ReconcileClusterLogForwarderNetworkPolicy(
				reqClient,
				constants.OpenshiftNS,
				policyName,
				instanceName,
				componentName,
				obsv1.NetworkPolicyRuleSetTypeRestrictIngressEgress,
				outputs,
				inputs,
				owner,
				commonLabels)).To(Succeed())

			// Get and check the network policy
			Expect(reqClient.Get(context.TODO(), policyKey, policyInstance)).Should(Succeed())

			// Verify ingress rules include metrics port and receiver ports
			Expect(policyInstance.Spec.Ingress).To(HaveLen(1))
			ingressRule := policyInstance.Spec.Ingress[0]

			// Should have 3 ingress ports: metrics port + 2 receiver ports
			Expect(ingressRule.Ports).To(HaveLen(3))

			ingressPortNumbers := make([]int32, 0)
			for _, port := range ingressRule.Ports {
				if port.Port != nil {
					ingressPortNumbers = append(ingressPortNumbers, port.Port.IntVal)
				}
			}

			Expect(ingressPortNumbers).To(ContainElements(int32(8080), int32(5140)))

			// Verify egress rules include output ports
			Expect(policyInstance.Spec.Egress).To(HaveLen(1))
			egressRule := policyInstance.Spec.Egress[0]

			Expect(egressRule.Ports).To(HaveLen(4))

			expectedEgressPorts := []networkingv1.NetworkPolicyPort{
				{
					Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
					Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: 9200},
				},
				{
					Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
					Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: 9092},
				},
				{
					Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
					Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: 6443},
				},
				{
					Protocol: &[]corev1.Protocol{corev1.ProtocolUDP}[0],
					Port:     &intstr.IntOrString{Type: intstr.String, StrVal: "dns"},
				},
			}
			Expect(egressRule.Ports).To(ConsistOf(expectedEgressPorts))
		})
	})

	Context("when the logfilemetricexporter NetworkPolicy is reconciled", func() {
		BeforeEach(func() {
			policyName = "lfme-test-network-policy"
			componentName = constants.LogfilesmetricexporterName
			commonLabels = func(o runtime.Object) {
				runtime.SetCommonLabels(o, constants.LogfilesmetricexporterName, instanceName, componentName)
			}
			owner = metav1.OwnerReference{
				APIVersion: "v1",
				Kind:       "LogFileMetricExporter",
				Name:       instanceName,
			}
			policyInstance = &networkingv1.NetworkPolicy{}
			policyKey = types.NamespacedName{Name: policyName, Namespace: namespace.Name}
		})

		It("should successfully reconcile the network policy", func() {
			// Reconcile the network policy
			Expect(ReconcileLogFileMetricsExporterNetworkPolicy(
				reqClient,
				constants.OpenshiftNS,
				policyName,
				instanceName,
				componentName,
				loggingv1alpha1.NetworkPolicyRuleSetTypeAllowIngressMetrics,
				owner,
				commonLabels)).To(Succeed())
			// Get and check the network policy
			Expect(reqClient.Get(context.TODO(), policyKey, policyInstance)).Should(Succeed())

			// Verify basic properties
			Expect(policyInstance.Name).To(Equal(policyName))
			Expect(policyInstance.Namespace).To(Equal(constants.OpenshiftNS))

			// Verify owner reference is set
			Expect(policyInstance.OwnerReferences).To(HaveLen(1))
			Expect(policyInstance.OwnerReferences[0].Name).To(Equal(instanceName))

			// Verify pod selector matches LFME pods
			expectedPodSelector := runtime.Selectors(instanceName, componentName, constants.LogfilesmetricexporterName)
			Expect(policyInstance.Spec.PodSelector.MatchLabels).To(Equal(expectedPodSelector))

			// Verify policy types includes Egress and Ingress
			expectedPolicyTypes := []networkingv1.PolicyType{
				networkingv1.PolicyTypeEgress,
				networkingv1.PolicyTypeIngress,
			}
			Expect(policyInstance.Spec.PolicyTypes).To(Equal(expectedPolicyTypes))

			// Verify ingress rules allow only the named metrics port
			expectedIngressRules := []networkingv1.NetworkPolicyIngressRule{
				{
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Protocol: &protocolTCP,
							Port:     &intstr.IntOrString{Type: intstr.String, StrVal: constants.MetricsPortName},
						},
					},
				},
			}
			Expect(policyInstance.Spec.Ingress).To(Equal(expectedIngressRules))

			// Verify egress rules are not present to deny all egress traffic
			Expect(policyInstance.Spec.Egress).To(BeNil())

			// Verify common labels are applied
			Expect(policyInstance.Labels).To(HaveKey(constants.LabelK8sName))
			Expect(policyInstance.Labels).To(HaveKey(constants.LabelK8sInstance))
			Expect(policyInstance.Labels).To(HaveKey(constants.LabelK8sComponent))
			Expect(policyInstance.Labels[constants.LabelK8sName]).To(Equal(constants.LogfilesmetricexporterName))
			Expect(policyInstance.Labels[constants.LabelK8sInstance]).To(Equal(instanceName))
			Expect(policyInstance.Labels[constants.LabelK8sComponent]).To(Equal(componentName))
		})
	})

})
