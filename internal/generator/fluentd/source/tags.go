package source

const (
	ApplicationTags    = "kubernetes.**"
	JournalTags        = "journal.** system.var.log**"
	InfraContainerTags = "kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_**"
	InfraTags          = InfraContainerTags + " " + JournalTags
	AuditTags          = "linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
)
