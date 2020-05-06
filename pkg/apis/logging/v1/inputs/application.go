package inputs

// Application provides optional extra properties for input `type: application`
type Application struct {
	// FIXME(alanconway) revisit nesting

	// Only collect logs from applications in these namespaces. If empty, all application container logs will be collected.
	//
	// +optional
	Namespaces []string `json:"namespaces"`
}

// FIXME(alanconway) fix deepcopy codegen.

// DeepCopyInto deep copies ApplicationType
// This is added because operator-sdk didnt generate it
func (a *Application) DeepCopyInto(b *Application) {
	if a != nil && b != nil {
		for _, name := range (*a).Namespaces {
			(*b).Namespaces = append((*b).Namespaces, name)
		}
	}
}

// DeepCopy deep copies ApplicationType
// This is added because operator-sdk didnt generate it
func (a *Application) DeepCopy() *Application {
	b := Application{}
	if a != nil {
		for _, name := range (*a).Namespaces {
			b.Namespaces = append(b.Namespaces, name)
		}
	}
	return &b
}
