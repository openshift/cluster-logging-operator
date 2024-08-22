
CHANGE_COMMIT=5a4dd8e0d
all: release-5.9.yaml proposed-6.0.yaml pre-change.yaml post-change.yaml changes-5.9-6.0.diff

release-5.9.yaml:
	git show release-5.9:../bundle/manifests/logging.openshift.io_clusterlogforwarders.yaml | ./cleanup.sh - > $@
pre-change.yaml:
	git show $(CHANGE_COMMIT)^:../bundle/manifests/observability.openshift.io_clusterlogforwarders.yaml | ./cleanup.sh - > $@
post-change.yaml:
	git show $(CHANGE_COMMIT):../bundle/manifests/observability.openshift.io_clusterlogforwarders.yaml | ./cleanup.sh - > $@
proposed-6.0.yaml: ../bundle/manifests/observability.openshift.io_clusterlogforwarders.yaml
	cat $< | ./cleanup.sh - > $@

changes-5.9-6.0.diff: release-5.9.yaml proposed-6.0.yaml diffdc.sh
	./diffdc.sh release-5.9.yaml proposed-6.0.yaml > $@

clean:
	rm *.yaml *.diff

