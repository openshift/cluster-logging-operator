package transport

type TLS struct {
	MinTlsVersion     string `json:"min_tls_version,omitempty"`
	CipherSuites      string `json:"ciphersuites,omitempty"`
	Curves            string `json:"curves,omitempty"`
	VerifyCertificate *bool  `json:"verify_certificate,omitempty"`
	VerifyHostname    *bool  `json:"verify_hostname,omitempty"`
	KeyFile           string `json:"key_file,omitempty"`
	CRTFile           string `json:"crt_file,omitempty"`
	CAFile            string `json:"ca_file,omitempty"`
	KeyPass           string `json:"key_pass,omitempty"`
}

// TlsEnabled provides configuration for components which need to explicitly enable TLS
type TlsEnabled struct {
	TLS
	Enabled bool `json:"enabled,omitempty"`
}
