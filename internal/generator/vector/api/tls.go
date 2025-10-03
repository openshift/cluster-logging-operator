package api

// TLS configures the TLS options for the sink's connection.
type TLS struct {
	// Enabled: Whether to require TLS. Default: true for some systems, check Vector docs.
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty" toml:"enabled,omitempty"`
	// VerifyCertificate: Whether to verify the server's certificate. Default: true.
	VerifyCertificate *bool `json:"verify_certificate,omitempty" yaml:"verify_certificate,omitempty" toml:"verify_certificate,omitempty"`
	// VerifyHostname: Whether to verify the server's hostname. Default: true.
	VerifyHostname *bool `json:"verify_hostname,omitempty" yaml:"verify_hostname,omitempty" toml:"verify_hostname,omitempty"`
	// CaFile: Path to the CA certificate file.
	CaFile string `json:"ca_file,omitempty" yaml:"ca_file,omitempty" toml:"ca_file,omitempty"`
	// CrtFile: Path to the client certificate file.
	CrtFile string `json:"crt_file,omitempty" yaml:"crt_file,omitempty" toml:"crt_file,omitempty"`
	// KeyFile: Path to the client key file.
	KeyFile string `json:"key_file,omitempty" yaml:"key_file,omitempty" toml:"key_file,omitempty"`
	// KeyPass: Password for the client key file.
	KeyPass string `json:"key_pass,omitempty" yaml:"key_pass,omitempty" toml:"key_pass,omitempty"`
}
