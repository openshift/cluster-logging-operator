package api

type SecretType string

const (
	SecretTypeDirectory SecretType = "directory"
)

type Secret struct {
	Type SecretType `json:"type" yaml:"type" toml:"type"`
	Path string     `json:"path" yaml:"path" toml:"path"`
}

func NewDirectorySecret(path string) *Secret {
	return &Secret{
		Type: SecretTypeDirectory,
		Path: path,
	}
}
