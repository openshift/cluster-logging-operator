package api

type SecretType string

const (
	SecretTypeDirectory SecretType = "directory"
)

type Secret struct {
	Type SecretType `json:"type"`
	Path string     `json:"path"`
}

func NewDirectorySecret(path string) *Secret {
	return &Secret{
		Type: SecretTypeDirectory,
		Path: path,
	}
}
