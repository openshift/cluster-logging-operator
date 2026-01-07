package sinks

type AwsAuth struct {
	AccessKeyId     string `json:"access_key_id,omitempty" yaml:"access_key_id,omitempty" toml:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secret_access_key,omitempty" yaml:"secret_access_key,omitempty" toml:"secret_access_key,omitempty"`
	AssumeRole      string `json:"assume_role,omitempty" yaml:"assume_role,omitempty" toml:"assume_role,omitempty"`
	ExternalId      string `json:"external_id,omitempty" yaml:"external_id,omitempty" toml:"external_id,omitempty"`
	SessionName     string `json:"session_name,omitempty" yaml:"session_name,omitempty" toml:"session_name,omitempty"`
	CredentialsFile string `json:"credentials_file,omitempty" yaml:"credentials_file,omitempty" toml:"credentials_file,omitempty"`
	Profile         string `json:"profile,omitempty" yaml:"profile,omitempty" toml:"profile,omitempty"`
}
