package sinks

type AwsAuth struct {
	AccessKeyId     string `json:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secret_access_key,omitempty"`
	AssumeRole      string `json:"assume_role,omitempty"`
	ExternalId      string `json:"external_id,omitempty"`
	SessionName     string `json:"session_name,omitempty"`
	CredentialsFile string `json:"credentials_file,omitempty"`
	Profile         string `json:"profile,omitempty"`
}
