package inputs

// ApplicationType provides optional extra properties for input `type: application`
type ApplicationType struct {
	// Namespaces referes to the namespace names. The input to the pipeline will be
	// restricted to logs from containers belonging to these namespaces.
	//
	// If no namespaces are specified, all application container logs will be collected.
	//
	// +optional
	Namespaces []string `json:"namespaces"`
}

// DeepCopyInto deep copies ApplicationType
// This is added because operator-sdk didnt generate it
func (a *ApplicationType) DeepCopyInto(b *ApplicationType) {
	if a != nil && b != nil {
		for _, name := range (*a).Namespaces {
			(*b).Namespaces = append((*b).Namespaces, name)
		}
	}
}

// DeepCopy deep copies ApplicationType
// This is added because operator-sdk didnt generate it
func (a *ApplicationType) DeepCopy() *ApplicationType {
	b := ApplicationType{}
	if a != nil {
		for _, name := range (*a).Namespaces {
			b.Namespaces = append(b.Namespaces, name)
		}
	}
	return &b
}
