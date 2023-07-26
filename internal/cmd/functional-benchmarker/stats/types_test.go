package stats

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Evaluating log loss stats", func() {

	var sample = `{"@timestamp":"2023-06-09T16:48:18.929938124+00:00","docker":{"container_id":"8d5fb743542b8c13afb0d5d3cf4cb52e724750d63de58006f16784c9c00da641"},"epoc_in":1682023305.6504774,"epoc_out":1686329299.934293,"hostname":"ip-10-0-173-26.us-east-2.compute.internal","kubernetes":{"container_image":"quay.io/openshift-logging/cluster-logging-load-client:latest","container_image_id":"quay.io/openshift-logging/cluster-logging-load-client@sha256:c8682432425be4b6f64821f0634fb1150b7cf805da0fa724b6047135b9d4d783","container_name":"loader-0","flat_labels":["test-client=true","testname=functional","testtype=functional"],"host":"ip-10-0-173-26.us-east-2.compute.internal","labels":{"test-client":"true","testname":"functional","testtype":"functional"},"master_url":"https://kubernetes.default.svc","namespace_id":"1dca9c0d-1365-4090-9bb3-80fda9b5fec6","namespace_labels":{"kubernetes_io_metadata_name":"testhack-imh2tzge","pod-security_kubernetes_io_enforce":"privileged","security_openshift_io_scc_podSecurityLabelSync":"false","test-client":"true"},"namespace_name":"testhack-imh2tzge","pod_id":"ca9b1dbf-84f3-4aa5-a876-46c0f4bba186","pod_ip":"10.131.0.21","pod_name":"functional"},"level":"unknown","log_type":"application","message":"goloader seq - functional.0.00000000000000002A536241AD427647 - 0000000123 - hvFtZpIOtnnhugEaqJqWWHqCmVCQXExRcOnXEaGpwWgIUlRjtCXLhmQUuqgnpStPjMudPrjUFRBBqolWCNCXdGrpvCqIclJzokNTrQgSUDKmrisDaohrCGDEHeihthYcAyUXmxlQHTkWEfViIVtQHoCQEBDFFWXJfKFqrgpIjgAdqBYeillNOikhURuwFZTEmwcWLElnrlBkZICapoGmFNeBxotMRQXGQLtntyCDiYjiihtvutkzbnHixjFuXKWkhWJbzHiVTuFJvYDlXBKfqGVQejokSwYueuDCdoAxyZCLglsVPijCEmjjGQaKLzlMYZApIOcZdeWVYIWZszhVDDfXArvuVxIdCtKSWMkQJXVuukvkSqKZbkQFcvaZdbDINbYgDAXOPMlYdGTULIwQdTSYyLIlSVyrxSZCnjZfAUhOTObIAXOZJhoJOaKmOpMqzLgYgSaZBaWiMcczEXVzVVXOsjdUKNJgjmdoyhvjcRbGQcLQNpRuTRsgaTbgnziKpEOVlngQEKDIkelRRhdErcIKisDQYNOxhIOysuJXDVYKILOLvritsLpqMJFGuBJjUHcJamedAMPGzkGidwrPWZeScywFftuevYThNNTzNqaQAAyYoSpHbmoZomqSYtDxAfnpCHAOBidVMwMXNLMHUuvrGVCdRwvEignmJOqaPCnYzFSbfZQFjFauIQfjCCnRvnQHPrCKPbnDWhuphjFIFJwjFKKqqmfDzaxizIoEjCHkBGWTEfEKfDYBKhwHfqLRBsGidLzdmAzUImEqWBUDFXWCHBhLIGlyQvGOLDSBymDfydqGoqxtlwVJNzjQdGTExPgBRxQZpVuqpOXGCHIGzjvBmggNSwWqHoxsxexAmjhtfIFrhHFkpkfYlcBTOyxetSvtVCmpAsAwGTRsFmkueLsgFeTqWYiLznDvYWtyXDRBLqwqepDsPPMRslWLYhIWuPRNTWXKsfNbWXpfigOtShXEIPCSEArirHpLLeamyQJOWkIsNWRdErHMpRpzHZXu","openshift":{"cluster_id":"functional","sequence":21526},"path":"/","pipeline_metadata":{"collector":{"inputname":"fluent-plugin-systemd","ipaddr4":"10.131.0.21","name":"fluentd","received_at":"2023-06-09T16:48:18.930687+00:00","version":"1.14.6 1.6.0"}},"source_type":"http_server","timestamp":"2023-06-09T16:48:19.933871646Z","viaq_msg_id":"NGFlNDVhNzItN2ZjZi00Nzc3LTliMzItOTlmZWEyODliYWJj"}`

	It("NewPerfLog", func() {
		log := NewPerfLog(sample)
		Expect(log.Bloat()).To(BeNumerically("~", 2.4, 0.1))
		Expect(log.SequenceId).To(Equal(123))
		Expect(log.EpocIn).To(Equal(1682023305.6504774))
		Expect(log.EpocOut).To(Equal(1686329299.934293))
		Expect(log.Stream).To(Equal("functional.0.00000000000000002A536241AD427647"))
	})
})
