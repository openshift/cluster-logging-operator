package types

type ExtensionType string

const (
	ExtensionTypeBearerTokenAuth ExtensionType = "bearertokenauth"
)

type Extension interface {
	ID() string
	ExtensionType() ExtensionType
}
