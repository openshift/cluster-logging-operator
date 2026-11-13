package common

const (
	ConfigVolumeName     = "config"
	DataDir              = "datadir"
	EntrypointVolumeName = "entrypoint"
	SecretDataReader     = "secret-data-reader"
	//TrustedCABundleHashName is the environment variable name for the md5 hash value of the
	//trusted ca bundle
	TrustedCABundleHashName = "TRUSTED_CA_HASH"

	// TrustedCABundleName is the name of the configmap to inject the trusted CA bundle
	TrustedCABundleName = "trusted-ca-bundle"
	// SelinuxTypeLogWriter (container_logwriter_t) is an MCS-constrained container domain that
	// grants read on /var/log/** plus inotify watch permissions on container_log_t, and write
	// on /var/lib/vector/**. Far more restrictive than spc_t (super-privileged container).
	SelinuxTypeLogWriter = "container_logwriter_t"
	// SelinuxTypeSpc (spc_t) is the Super-Privileged Container SELinux domain.
	// It is used exclusively by short-lived init containers (e.g., data-dir-init) to
	// create host directories and set SELinux labels (chcon) before the main collector
	// launches under container_logwriter_t.
	SelinuxTypeSpc = "spc_t"

	// NonRootUser is the fixed non-root UID
	NonRootUser int64 = 1000
)
