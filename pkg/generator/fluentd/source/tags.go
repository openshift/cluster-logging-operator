package source

const (
	ApplicationTags    = "kubernetes.**"
	JournalTags        = "journal.** system.var.log**"
	InfraContainerTags = "**_default_** **_kube-*_** **_openshift-*_** **_openshift_**"
	InfraTags          = InfraContainerTags + " " + JournalTags
	AuditTags          = "linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
)
