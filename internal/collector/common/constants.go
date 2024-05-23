package common

const (
	ConfigVolumeName     = "config"
	DataDir              = "datadir"
	EntrypointVolumeName = "entrypoint"

	//TrustedCABundleHashName is the environment variable name for the md5 hash value of the
	//trusted ca bundle
	TrustedCABundleHashName = "TRUSTED_CA_HASH"

	// TrustedCABundleName is the name of the configmap to inject the trusted CA bundle
	TrustedCABundleName = "trusted-ca-bundle"
)
